package main

import (
	"flag"

	"github.com/premiering/wubsub/app"
)

func main() {
	port := flag.Int("port", 9190, "The port the server will listen on")
	debugMode := flag.Bool("debug", false, "When enabled, pub/sub events are logged to output")
	flag.Parse()

	app := app.CreateApp(*port, *debugMode)
	app.Start()
}
