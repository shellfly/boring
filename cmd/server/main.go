package main

import (
	"flag"

	"github.com/shellfly/boring"
	log "github.com/shellfly/boring/pkg/log"
)

var (
	addr     = flag.String("addr", ":1080", "server listen addr")
	key      = flag.String("key", "", "encryption key")
	logLevel = flag.String("log.level", "info", "log level")
)

func main() {
	flag.Parse()
	log.SetLevel(*logLevel)
	log.Infof("Start boring server on %s", *addr)
	srv := boring.NewServer(*addr, *key)
	log.Panic(srv.ListenAndServe())
}
