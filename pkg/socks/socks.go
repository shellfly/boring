package socks

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/shellfly/boring/pkg/log"
	"github.com/shellfly/boring/pkg/tcp"
)

// Wire protocol constants.
const (
	Version5     = 0x05
	AddrTypeIPv4 = 0x01
	AddrTypeFQDN = 0x03
	AddrTypeIPv6 = 0x04

	CmdConnect = 0x01 // establishes an active-open forward proxy connection
	cmdBind    = 0x02 // establishes a passive-open forward proxy connection

	AuthMethodNotRequired         = 0x00 // no authentication required
	AuthMethodUsernamePassword    = 0x02 // use username/password
	AuthMethodNoAcceptableMethods = 0xff // no acceptable authentication methods

	StatusSucceeded = 0x00

	MaxAddrLen = 1 + 1 + 255 + 2
)

type Handler func(context.Context, net.Conn, Addr)

// Addr represents a SOCKS address as defined in RFC 1928 section 5.
type Addr []byte

// String serializes bytes address to string form.
func (a Addr) String() string {
	var host, port string

	switch a[0] { // address type
	case AddrTypeFQDN:
		host = string(a[2 : 2+int(a[1])])
		port = strconv.Itoa((int(a[2+int(a[1])]) << 8) | int(a[2+int(a[1])+1]))
	case AddrTypeIPv4:
		host = net.IP(a[1 : 1+net.IPv4len]).String()
		port = strconv.Itoa((int(a[1+net.IPv4len]) << 8) | int(a[1+net.IPv4len+1]))
	case AddrTypeIPv6:
		host = net.IP(a[1 : 1+net.IPv6len]).String()
		port = strconv.Itoa((int(a[1+net.IPv6len]) << 8) | int(a[1+net.IPv6len+1]))
	}

	return net.JoinHostPort(host, port)
}

// socks5 implements RFC1928
type socks struct {
	handler Handler
}

func NewSocks5Server(addr string, handler Handler) *tcp.Server {
	s := &socks{handler: handler}
	return tcp.NewServer(addr, s)
}

func (s *socks) ServeConn(ctx context.Context, rw net.Conn) {
	defer rw.Close()

	addr, err := s.handshake(ctx, rw)
	if err != nil {
		log.Errorf("socks handshake error: %v", err)
		return
	}

	s.handler(ctx, rw, addr)
}

func (s *socks) handshake(ctx context.Context, rw io.ReadWriter) (addr []byte, err error) {
	// VER, NMETHODS
	buf := make([]byte, MaxAddrLen)
	if _, err = io.ReadFull(rw, buf[:2]); err != nil {
		return
	}
	version := buf[0]
	log.Debugf("client socks version：%d", version)
	if version != Version5 {
		err = fmt.Errorf("invalid socks version: %v", version)
		return
	}
	nMethods := int(buf[1])
	log.Debugf("methods length：%d", nMethods)
	if _, err = io.ReadFull(rw, buf[:nMethods]); err != nil {
		return
	}
	resp := []byte{Version5, byte(AuthMethodNotRequired)}
	rw.Write(resp)

	//VER CMD RSV
	if _, err = io.ReadFull(rw, buf[:3]); err != nil {
		return
	}
	cmd := buf[1]
	addr, err = readAddr(rw, buf)
	if err != nil {
		return
	}

	switch cmd {
	case CmdConnect:
		_, err = rw.Write([]byte{Version5, byte(StatusSucceeded), 0, 1, 0, 0, 0, 0, 0, 0})
	default:
		err = errors.New("command not supported")
	}
	return
}

func readAddr(r io.Reader, b []byte) (addr Addr, err error) {
	if len(b) < MaxAddrLen {
		err = io.ErrShortBuffer
		return
	}
	_, err = io.ReadFull(r, b[:1]) // read 1st byte for address type
	if err != nil {
		return
	}
	addrtype := b[0]
	addrlen := 0
	start := 1
	switch addrtype {
	case AddrTypeFQDN:
		_, err = io.ReadFull(r, b[1:2])
		if err != nil {
			return
		}
		addrlen = int(b[1])
		start = 2
	case AddrTypeIPv4:
		addrlen = net.IPv4len
	case AddrTypeIPv6:
		addrlen = net.IPv6len
	}

	_, err = io.ReadFull(r, b[start:start+addrlen+2]) // 2 bytes for port
	if err != nil {
		return
	}
	return b[:start+addrlen+2], nil
}

// ReadAddr reads get a valid addr string from socks connection
func ReadAddr(r io.Reader) (Addr, error) {
	return readAddr(r, make([]byte, MaxAddrLen))
}
