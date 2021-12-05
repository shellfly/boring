package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"fmt"
	"io"
	"net"

	"github.com/shellfly/boring/pkg/log"
)

const (
	MaxPayloadSize = 32 * 1024
)

type CipherFunc func([]byte) (cipher.Block, error)
type CryptFunc func([]byte) []byte

type Crypto interface {
	Conn(net.Conn) net.Conn
}

type dummy struct{}

func (dummy) Conn(c net.Conn) net.Conn { return c }

type aesC struct {
	key string
}

func NewCrypto(method, key string) Crypto {
	switch method {
	case "dummy":
		return &dummy{}
	case "aes":
		return &aesC{key}
	default:
		panic("unimplemented method")
	}
}

func (a *aesC) Conn(conn net.Conn) net.Conn {
	var k []byte
	for _, b := range md5.Sum([]byte(a.key)) {
		k = append(k, b)
	}
	block, _ := aes.NewCipher(k)
	return &CryptConn{
		Conn: conn,
		// bufr:  bufio.NewReader(conn),
		// bufw:  bufio.NewWriter(conn),
		block: block,
		buf:   make([]byte, MaxPayloadSize),
	}
}

type CryptConn struct {
	net.Conn
	block cipher.Block

	//r *reader
	//	bufr  *bufio.Reader
	//	bufw  *bufio.Writer
	buf      []byte
	leftover []byte
}

func (cc *CryptConn) Read(b []byte) (n int, err error) {
	defer func() {
		fmt.Println("read data :", len(b), b)
	}()
	fmt.Println("want to read length, leftover", len(b), len(cc.leftover))
	if len(cc.leftover) > 0 {
		n := copy(b, cc.leftover)
		cc.leftover = cc.leftover[n:]
		return n, nil
	}

	n, err = cc.read()
	if err != nil {
		log.Error("read decrypt error: ", err)
		return
	}
	m := copy(b, cc.buf[:n])
	fmt.Println("m,n", m, n)
	if m < n { // insufficient len(b), keep leftover for next read
		cc.leftover = cc.buf[m:n]
	}
	return m, err
}

func (cc *CryptConn) read() (n int, err error) {
	buf := make([]byte, MaxPayloadSize)
	n, err = io.ReadFull(cc.Conn, buf[:4])
	if err != nil {
		log.Error("read from cryptconn error: ", err)
		return
	}

	datalen := int(buf[0])<<24 | int(buf[1])<<16 | int(buf[2])<<8 | int(buf[3])
	fmt.Println("cryptconn read lenbuf:", buf[:4], datalen)
	n, err = io.ReadFull(cc.Conn, buf[:datalen])
	if err != nil {
		log.Error("read from cryptconn error: ", err)
		return
	}
	data := cc.Decrypt(buf[:datalen])
	fmt.Println("read decrypt data: ", data, len(data))
	return copy(cc.buf, data), nil
}

func (cc *CryptConn) Write(b []byte) (n int, err error) {
	fmt.Println("write data: ", len(b), b)
	data := cc.Encrypt(b)
	length := len(data)
	cc.Conn.Write([]byte{byte(length >> 24), byte(length >> 16), byte(length >> 8), byte(0xFF & length)})
	_, err = cc.Conn.Write(data)
	if err != nil {
		log.Error("write data to cryptconn error: ", err)
		return
	}
	// pretend to return length of original data
	return len(b), nil
}

// https://golang.org/src/crypto/cipher/example_test.go
func (cc *CryptConn) Encrypt(src []byte) []byte {
	blockSize := cc.block.BlockSize()
	src = pkcs7Padding(src, blockSize)
	dst := make([]byte, blockSize+len(src))
	iv := dst[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(cc.block, iv)
	mode.CryptBlocks(dst[blockSize:], src)
	return dst
}

func (cc *CryptConn) Decrypt(src []byte) []byte {
	blockSize := cc.block.BlockSize()
	iv := src[:blockSize]
	src = src[blockSize:]
	mode := cipher.NewCBCDecrypter(cc.block, iv)

	// CryptBlocks can work in-place if the two arguments are the same.
	mode.CryptBlocks(src, src)
	src = pkcs7UnPadding(src)
	return src
}

func pkcs7Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func pkcs7UnPadding(origData []byte) []byte {
	length := len(origData)
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
