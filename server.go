package boring

import (
	"context"
	"net"

	"github.com/shellfly/boring/pkg/crypto"
	"github.com/shellfly/boring/pkg/log"
	"github.com/shellfly/boring/pkg/socks"
	"github.com/shellfly/boring/pkg/tcp"
)

type Server struct {
	*tcp.Server

	crypto crypto.Crypto
}

func (s *Server) ServeConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	conn = s.crypto.Conn(conn)
	addr, err := socks.ReadAddr(conn)
	if err != nil {
		log.Error("ReadAddr error: ", err)
		return
	}
	log.Info("handle boring proxy request ", addr.String())
	remote, err := net.Dial("tcp", addr.String())
	if err != nil {
		log.Error("Connect to addr error", err)
		return
	}
	defer remote.Close()
	log.Debug("start boring process for: ", addr)
	Boring(conn, remote)
}

// NewServer returns a tcp server which handles requests from boring client
func NewServer(addr string, crypto crypto.Crypto) *Server {
	s := &Server{}
	s.Server = tcp.NewServer(addr, s)
	s.crypto = crypto
	return s
}
