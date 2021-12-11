package boring

import (
	"context"
	"net"

	"github.com/shellfly/boring/pkg/crypto"
	"github.com/shellfly/boring/pkg/log"
	"github.com/shellfly/boring/pkg/socks"
	"github.com/shellfly/boring/pkg/tcp"
)

type Client struct {
	*tcp.Server
	crypto     crypto.Crypto
	remoteAddr string
}

// socksHandler handle socks5 proxy request coming from local application
func (c *Client) socksHandler(ctx context.Context, rw net.Conn, addr socks.Addr) {
	log.Info("Handle socks5 proxy request: ", addr.String())
	remote, err := net.Dial("tcp", c.remoteAddr)
	if err != nil {
		log.Error("Connect to boring remote server error", err)
		return
	}
	defer remote.Close()
	remote = c.crypto.EncryptConn(remote)
	if _, err := remote.Write(addr); err != nil {
		log.Error("Send addr to boring remote server err:", err)
		return
	}

	log.Debug("Start boring process for: ", addr)
	Boring(rw, remote)
}

// NewClient returns a tunnel client, it's also a socks5 server
// to accept proxy requests.
func NewClient(localAddr, remoteAddr string, crypto crypto.Crypto) *Client {
	c := &Client{}
	c.Server = socks.NewSocksServer(localAddr, c.socksHandler)
	c.crypto = crypto
	c.remoteAddr = remoteAddr
	return c
}
