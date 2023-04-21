package main

import (
	"flag"
	"os"
)

type Options struct {
	endpoint string
	logLevel string
}

var options Options

func parseFlags() {
	/*Разбираем командную строку*/
	flag.StringVar(&options.endpoint, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&options.logLevel, "l", "info", "log level")
	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		options.endpoint = addr
	}

	if lvl, exist := os.LookupEnv("LOG_LEVEL"); exist {
		options.logLevel = lvl
	}
}
