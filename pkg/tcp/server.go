package tcp

import (
	"context"
	"errors"
	"net"
	"time"

	"github.com/shellfly/boring/pkg/log"
)

type ConnHandler interface {
	Init(net.Conn)
	ServeConn(context.Context)
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
		rw, e := l.Accept()
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
				log.Errorf("socks5: Accept error: %v; retrying in %v", e, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return e
		}

		baseCtx := context.Background()
		srv.handler.Init(rw)
		go srv.handler.ServeConn(baseCtx)
	}
}

func NewServer(addr string, handler ConnHandler) *Server {
	return &Server{
		addr:    addr,
		handler: handler,
	}
}
