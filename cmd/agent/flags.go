package main

import (
	"flag"
)

type Options struct {
	serverEndpoint string
	reportInterval int64
	pollInterval   int64
}

var options Options

func parseFlags() {
	flag.StringVar(&options.serverEndpoint, "a", "http://localhost:8080", "server address and port")
	flag.Int64Var(&options.reportInterval, "r", 10, "report interval")
	flag.Int64Var(&options.pollInterval, "p", 2, "poll interval")

	flag.Parse()
}
