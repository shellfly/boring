package socks5

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	"github.com/shellfly/socks5/pkg/log"
)

type conn struct {
	server *Server
	r      *bufio.Reader
	rwc    net.Conn
}

func (c *conn) serve(ctx context.Context) {
	defer c.rwc.Close()
	if err := c.shake(); err != nil {
		log.Errorln(err)
		return
	}

	addr, err := c.getAddr()
	if err != nil {
		log.Errorln(err)
		return
	}

	log.Infoln("proxy to addr: ", addr)
	remote, err := net.Dial("tcp", addr)
	if err != nil {
		log.Errorln(err)
		return
	}
	defer remote.Close()

	wg := new(sync.WaitGroup)
	wg.Add(2)
	go func() {
		defer wg.Done()
		// copy source request to remote host
		// TODO: use c.conn? does c.r has retry mechanism?
		io.Copy(remote, c.r)
	}()

	go func() {
		defer wg.Done()
		// copy remote response to source host
		io.Copy(c.rwc, remote)
	}()
	wg.Wait()
}

func (c *conn) shake() error {
	// read first byte, get socks version
	version, _ := c.r.ReadByte()
	log.Debugf("client version：%d", version)
	if version != 5 {
		return fmt.Errorf("invalid socks version: %d", version)
	}

	nMethods, _ := c.r.ReadByte()
	log.Debugf("methods length：%d", nMethods)

	// 0 no authentication, 1 GSSAPI, 2 user password etc.
	buf := make([]byte, nMethods)
	io.ReadFull(c.r, buf)
	log.Debugf("authentication：%v", buf)

	resp := []byte{5, 0}
	c.rwc.Write(resp)
	return nil
}

func (c *conn) getAddr() (string, error) {
	version, _ := c.r.ReadByte()
	log.Debugf("client version：%d", version)
	if version != 5 {
		return "", fmt.Errorf("invalid socks version: %d", version)
	}

	// 1 CONNECT, 2 BIND, 3 UDP ASSOCIATE
	cmd, _ := c.r.ReadByte()
	log.Debugf("client request type：%d\n", cmd)
	if cmd != 1 {
		return "", fmt.Errorf("unimplemented client request: %d", cmd)
	}

	// skip RSV field
	c.r.ReadByte()

	// 1 IPV4, 3 DOMAIN NAME, 4 IPV6
	addrtype, _ := c.r.ReadByte()
	// TODO: support more type
	if addrtype != 3 {
		return "", fmt.Errorf("unimplemented addr type: %d", addrtype)
	}
	addrlen, _ := c.r.ReadByte()
	addr := make([]byte, addrlen)
	io.ReadFull(c.r, addr)
	log.Debugf("get proxy domain: %s\n", addr)

	var port int16
	binary.Read(c.r, binary.BigEndian, &port)
	resp := []byte{0x05, 0x00, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
	c.rwc.Write(resp)
	return fmt.Sprintf("%s:%d", addr, port), nil
}

type Server struct {
	Addr string
}

// ListenAndServe always returns a non-nil error.
func (srv *Server) ListenAndServe() error {
	addr := srv.Addr
	if addr == "" {
		addr = ":1080"
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

		c := srv.newConn(rw)
		baseCtx := context.Background()
		go c.serve(baseCtx)
	}
}

// Create new connection from rwc.
func (srv *Server) newConn(rwc net.Conn) *conn {
	r := bufio.NewReader(rwc)
	return &conn{
		server: srv,
		r:      r,
		rwc:    rwc,
	}
}

// ListenAndServe always returns a non-nil error.
func ListenAndServe(addr string) error {
	server := &Server{Addr: addr}
	return server.ListenAndServe()
}
