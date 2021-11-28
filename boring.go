package boring

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/shellfly/boring/pkg/crypto"
	"github.com/shellfly/boring/pkg/log"
)

const (
	Version1        = 0x1
	StatusSucceeded = 0x00
)

type Tunnel struct {
	crypto crypto.Crypto
}

func (t Tunnel) checkVersion(ctx context.Context, remote net.Conn) error {
	log.Debugln("checking version")
	version, err := t.receive(ctx, remote)
	if err != nil {
		return err
	}
	if len(version) != 1 || version[0] != byte(Version1) {
		return fmt.Errorf("unknown version: %v", version)
	}
	return nil
}

func (t Tunnel) send(ctx context.Context, remote net.Conn, data []byte) error {
	log.Debugln("send data into tunnel", data, string(data))
	data = t.crypto.Encrypt(data)
	length := len(data)
	b := append([]byte{byte(length)}, data...)
	nw, err := remote.Write(b)
	if nw != len(b) {
		return fmt.Errorf("invalid write length, want %d, got %d", len(b), nw)
	}
	return err
}

func (t Tunnel) receive(ctx context.Context, remote net.Conn) (data []byte, err error) {
	defer func() {
		// catch possible panic from Decrypt function
		if e := recover(); e != nil {
			err = fmt.Errorf("receive data error :%v", e)
		}
		if err == nil {
			log.Debugln("received data:", data, string(data))
		}
	}()
	log.Debugln("receiving data...")
	b := make([]byte, 32) // 10 is an estimated length
	if _, err := io.ReadFull(remote, b[:1]); err != nil {
		return nil, err
	}
	length := int(b[0])
	if length > len(b) {
		b = make([]byte, length)
	}
	if _, err := io.ReadFull(remote, b[:length]); err != nil {
		return nil, err
	}
	return t.crypto.Decrypt(b[:length]), nil
}

// Borning copy data into and out of tunnel
// edge(local) --> [encrypt] -->  tunnel --> [decrypt] --> edge(remote)
/// edge(local) --> [encrypt] -->  tunnel --> [decrypt] --> internet(remote)
func Boring(edge, tunnel net.Conn, c crypto.Crypto) {
	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		if _, err := Copy(tunnel, edge, c.Encrypt); err != nil {
			log.Errorln("copy data into tunnel error", err)
		}

	}()
	go func() {
		defer wg.Done()
		if _, err := Copy(edge, tunnel, c.Decrypt); err != nil {
			log.Errorln("copy data from tunnel error", err)
		}

	}()
	wg.Wait()
}

// Copy...
func Copy(dst io.Writer, src io.Reader, crypt crypto.CryptFunc) (written int64, err error) {
	size := 32 * 1024
	if l, ok := src.(*io.LimitedReader); ok && int64(size) > l.N {
		if l.N < 1 {
			size = 1
		} else {
			size = int(l.N)
		}
	}
	buf := make([]byte, size)
	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			// data := crypt(buf[0:nr])
			// nr = len(data)
			nw, ew := dst.Write(buf[0:nr])
			if nw < 0 || nr < nw {
				nw = 0
				if ew == nil {
					ew = errors.New("invalid write")
				}
			}
			written += int64(nw)
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er != nil {
			if er != io.EOF {
				err = er
			}
			break
		}
	}
	return written, err
}
