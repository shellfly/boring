package tcp

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"

	"github.com/shellfly/boring/pkg/log"
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
)

type handler func(context.Context, net.Conn, string)

// socks5 implements ConnHandler interface
type socks5 struct {
	r       *bufio.Reader
	rwc     net.Conn
	handler handler
}

func NewSocks5(h handler) *socks5 {
	return &socks5{
		handler: h,
	}
}

func (s *socks5) Init(conn net.Conn) {
	s.r = bufio.NewReader(conn)
	s.rwc = conn
}

func (s *socks5) ServeConn(ctx context.Context) {
	defer s.rwc.Close()

	if err := s.handshake(); err != nil {
		log.Errorln(err)
		return
	}

	addr, err := s.getAddr()
	if err != nil {
		log.Errorln(err)
		return
	}
	s.handler(ctx, s.rwc, addr)
}

func (s *socks5) checkVersion() error {
	// read first byte, get socks version
	version, _ := s.r.ReadByte()
	log.Debugf("client version：%d", version)
	if version != Version5 {
		return fmt.Errorf("invalid socks version: %d", version)
	}
	return nil
}
func (s *socks5) handshake() error {
	if err := s.checkVersion(); err != nil {
		return err
	}

	nMethods, _ := s.r.ReadByte()
	log.Debugf("methods length：%d", nMethods)
	buf := make([]byte, nMethods)
	io.ReadFull(s.r, buf)
	resp := []byte{Version5, byte(AuthMethodNotRequired)}
	s.rwc.Write(resp)
	return nil
}

func (s *socks5) getAddr() (string, error) {
	if err := s.checkVersion(); err != nil {
		return "", err
	}

	// 1 CONNECT, 2 BIND, 3 UDP ASSOCIATE
	cmd, _ := s.r.ReadByte()
	log.Debugf("client request type：%d\n", cmd)
	if cmd != byte(CmdConnect) {
		return "", fmt.Errorf("unimplemented client request: %d", cmd)
	}

	// skip RSV field
	s.r.ReadByte()

	addrtype, _ := s.r.ReadByte()
	// TODO: support more type
	if addrtype != AddrTypeFQDN {
		return "", fmt.Errorf("unimplemented addr type: %d", addrtype)
	}
	addrlen, _ := s.r.ReadByte()
	addr := make([]byte, addrlen)
	io.ReadFull(s.r, addr)
	log.Debugf("get proxy domain: %s\n", addr)

	var port int16
	binary.Read(s.r, binary.BigEndian, &port)
	resp := []byte{Version5, byte(StatusSucceeded), 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	s.rwc.Write(resp)
	return fmt.Sprintf("%s:%d", addr, port), nil
}
