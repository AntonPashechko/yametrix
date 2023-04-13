package main

import (
	"flag"
	"os"
)

var endpoint string

func parseFlags() {
	flag.StringVar(&endpoint, "a", "localhost:8080", "address and port to run server")
	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		endpoint = addr
	}
}
