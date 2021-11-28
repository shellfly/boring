package main

import (
	"flag"

	"github.com/shellfly/boring"
	log "github.com/shellfly/boring/pkg/log"
)

var (
	addr     = flag.String("addr", ":1081", "tunnel client listen addr")
	server   = flag.String("server", "", "tunnel server addr")
	key      = flag.String("key", "", "encryption key")
	logLevel = flag.String("log.level", "info", "log level")
)

func main() {
	flag.Parse()
	log.SetLevel(*logLevel)
	log.Infof("Start boring client on %s", *addr)
	cli := boring.NewClient(*addr, *server, *key)
	log.Panic(cli.ListenAndServe())
}
