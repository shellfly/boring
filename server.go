package boring

import (
	"bufio"
	"context"
	"net"

	"github.com/shellfly/boring/pkg/crypto"
	"github.com/shellfly/boring/pkg/log"
	"github.com/shellfly/boring/pkg/tcp"
)

type Server struct {
	*tcp.Server
	Tunnel
	r    *bufio.Reader
	rwc  net.Conn
	done chan struct{}
}

func (s *Server) Close(conn net.Conn) {
	s.rwc.Close()
}
func (s *Server) Init(conn net.Conn) {
	s.r = bufio.NewReader(conn)
	s.rwc = conn
}

func (s *Server) ServeConn(ctx context.Context) {
	defer s.rwc.Close()
	if err := s.handshake(ctx); err != nil {
		log.Errorln("handshake error: ", err)
		return
	}
	addr, err := s.getAddr(ctx)
	if err != nil {
		log.Errorln("getAddr error: ", err)
		return
	}
	s.handler(ctx, s.rwc, addr)
}

func (s *Server) handshake(ctx context.Context) error {
	// check client version
	if err := s.checkVersion(ctx, s.rwc); err != nil {
		return err
	}
	// send server version
	return s.send(ctx, s.rwc, []byte{Version1})
}
func (s *Server) getAddr(ctx context.Context) (string, error) {
	data, err := s.receive(ctx, s.rwc)
	if err != nil {
		return "", err
	}
	addr := string(data)
	if err := s.send(ctx, s.rwc, []byte{Version1}); err != nil {
		return "", err
	}
	// send response, 1 byte resp, 2 byte RSV
	b := []byte{0x00, 0x00, 0x00}
	if _, err := s.rwc.Write(b); err != nil {
		return "", err
	}
	return addr, nil
}

func (s *Server) handler(ctx context.Context, sc net.Conn, addr string) {
	log.Info("handle boring proxy request ", addr)
	remote, err := net.Dial("tcp", addr)
	if err != nil {
		log.Errorln("Connect to addr error", err)
		return
	}
	defer remote.Close()
	log.Debugln("start boring process for: ", addr)
	Boring(remote, sc, s.crypto)
}

// NewServer returns a tcp server which handles requests from boring client
func NewServer(addr, key string) *Server {
	s := &Server{done: make(chan struct{})}
	s.Server = tcp.NewServer(addr, s)
	s.Tunnel = Tunnel{crypto.NewCrypto("aes", key)}
	return s
}
