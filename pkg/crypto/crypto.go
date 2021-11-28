package crypto

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/md5"
	"crypto/rand"
	"io"
)

type CipherFunc func([]byte) (cipher.Block, error)
type CryptFunc func([]byte) []byte

type Crypto interface {
	Encrypt([]byte) []byte
	Decrypt([]byte) []byte
}

type crypto struct {
	method string
	block  cipher.Block
}

var ciphers = map[string]CipherFunc{
	"aes": aes.NewCipher,
}

func NewCrypto(method, key string) Crypto {
	cipher, ok := ciphers[method]
	if !ok {
		panic("unimplemented method")
	}
	var k []byte
	for _, b := range md5.Sum([]byte(key)) {
		k = append(k, b)
	}
	block, _ := cipher(k)
	return &crypto{method: method, block: block}
}

// https://golang.org/src/crypto/cipher/example_test.go
func (c *crypto) Encrypt(src []byte) []byte {
	blockSize := c.block.BlockSize()
	src = pkcs7Padding(src, blockSize)
	dst := make([]byte, blockSize+len(src))
	iv := dst[:blockSize]
	if _, err := io.ReadFull(rand.Reader, iv); err != nil {
		panic(err)
	}

	mode := cipher.NewCBCEncrypter(c.block, iv)
	mode.CryptBlocks(dst[blockSize:], src)
	return dst
}

func (c *crypto) Decrypt(src []byte) []byte {
	blockSize := c.block.BlockSize()
	iv := src[:blockSize]
	src = src[blockSize:]
	mode := cipher.NewCBCDecrypter(c.block, iv)

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
