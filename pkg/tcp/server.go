package tcp

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/shellfly/boring/pkg/log"
)

type ConnHandler interface {
	ServeConn(context.Context, net.Conn)
}

type Server struct {
	addr    string
	handler ConnHandler
}

// ListenAndServe always returns a non-nil error.
func (srv *Server) ListenAndServe() error {
	addr := srv.addr
	if addr == "" {
		return errors.New("listen address not specified")
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	return srv.Serve(ln)
}

// Serve always returns a non-nil error.
func (srv *Server) Serve(l net.Listener) error {
	defer l.Close()
	var tempDelay time.Duration // how long to sleep on accept failure

	for {
		conn, e := l.Accept()
		if e != nil {
			if ne, ok := e.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				log.Errorf("Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}

		baseCtx := context.Background()
		go srv.handler.ServeConn(baseCtx, conn)
	}
}

func NewServer(addr string, handler ConnHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}
