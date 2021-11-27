package main

import (
	"flag"

	"github.com/shellfly/socks5"
	log "github.com/shellfly/socks5/pkg/log"
)

var (
	addr     = flag.String("addr", ":1080", "socket server listen addr")
	logLevel = flag.String("log.level", "info", "log level")
)

func main() {
	flag.Parse()
	log.SetLevel(*logLevel)
	log.Infof("Start socks5 server on %s", *addr)
	log.Panic(socks5.ListenAndServe(*addr))
}
