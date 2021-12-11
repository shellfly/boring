package main

import (
	"flag"

	"github.com/shellfly/boring"
	"github.com/shellfly/boring/pkg/crypto"
	log "github.com/shellfly/boring/pkg/log"
)

var (
	addr     = flag.String("addr", ":1080", "boring server listen addr")
	method   = flag.String("method", "", "encryption method")
	key      = flag.String("key", "", "encryption key")
	logLevel = flag.String("log.level", "info", "log level")
)

func main() {
	flag.Parse()
	log.SetLevel(*logLevel)
	log.Infof("Start boring server on %s", *addr)
	crypto := crypto.NewCrypto(*method, *key)
	srv := boring.NewServer(*addr, crypto)
	log.Panic(srv.ListenAndServe())
}
