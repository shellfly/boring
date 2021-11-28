package boring

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"

	"github.com/shellfly/boring/pkg/crypto"
	"github.com/shellfly/boring/pkg/log"
	"github.com/shellfly/boring/pkg/tcp"
)

type Client struct {
	*tcp.Server
	Tunnel
	remoteAddr string
}

// socksHandler handle socks5 proxy request coming from local application
func (c *Client) socksHandler(ctx context.Context, sc net.Conn, addr string) {
	log.Infoln("Handle socks5 proxy request", addr)
	remote, err := c.connect(ctx, addr)
	if err != nil {
		log.Errorln("Connect to boring remote server error", err)
		return
	}
	defer remote.Close()
	log.Debugln("start boring process for: ", addr)
	Boring(sc, remote, c.crypto)
}

func (c *Client) handshake(ctx context.Context, remote net.Conn) error {
	// send client version with encryption
	if err := c.send(ctx, remote, []byte{Version1}); err != nil {
		return err
	}
	// check server version
	return c.checkVersion(ctx, remote)
}

func (c *Client) sendAddr(ctx context.Context, remote net.Conn, addr string) error {
	log.Debugln("sending addr")
	// send addr with encryption
	if err := c.send(ctx, remote, []byte(addr)); err != nil {
		return err
	}
	// read remote version
	version, err := c.receive(ctx, remote)
	if err != nil {
		return err
	}
	if len(version) != 1 || version[0] != byte(Version1) {
		return fmt.Errorf("unknown version: %v", version)
	}

	// read response, 1 byte resp, 2 byte RSV
	b := make([]byte, 3)
	if _, err = io.ReadFull(remote, b); err != nil {
		return err
	}

	if resp := byte(b[0]); resp != StatusSucceeded {
		return errors.New("unknown status " + string(resp))
	}
	return nil
}

// Connect create an encrypted tunnel to boring server
func (c *Client) connect(ctx context.Context, addr string) (net.Conn, error) {
	remote, err := net.Dial("tcp", c.remoteAddr)
	if err != nil {
		return nil, err
	}
	if err := c.handshake(ctx, remote); err != nil {
		log.Debug("handshake error: ", err)
		return nil, err
	}
	if err := c.sendAddr(ctx, remote, addr); err != nil {
		log.Debug("sendAddr error: ", err)
		return nil, err
	}
	return remote, nil
}

// NewClient returns a tunnel client, it's also a socks5 server
// to accept proxy requests.
func NewClient(localAddr, remoteAddr, key string) *Client {
	c := &Client{}
	socks5 := tcp.NewSocks5(c.socksHandler)
	c.Server = tcp.NewServer(localAddr, socks5)
	c.Tunnel = Tunnel{crypto.NewCrypto("aes", key)}
	c.remoteAddr = remoteAddr
	return c
}
