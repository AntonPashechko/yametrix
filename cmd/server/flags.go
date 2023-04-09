package main

import (
	"flag"
)

var endpoint string

func parseFlags() {
	flag.StringVar(&endpoint, "a", "localhost:8080", "address and port to run server")
	flag.Parse()
}
