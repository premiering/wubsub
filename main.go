package main

import (
	"flag"

	"github.com/premiering/wubsub/log"
)

func main() {
	port := flag.Int("port", 9190, "The port the server will listen on")
	debugMode := flag.Bool("debug", false, "When enabled, pub/sub events are logged to output")
	tls := flag.Bool("tls", false, "Enables TLS, (requires key and cert file)")
	keyfile := flag.String("keyfile", "", "Path to key file relative to working dir")
	certfile := flag.String("certfile", "", "Path to cert file relative to working dir")
	flag.Parse()

	log.SetAppName("wubsub")
	app := CreateApp(*port, *debugMode, *tls, *keyfile, *certfile)
	app.Start()
}
