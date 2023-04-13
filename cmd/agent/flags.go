package main

import (
	"flag"
	"os"
	"strings"

	"github.com/AntonPashechko/yametrix/pkg/utils"
)

type Options struct {
	serverEndpoint string
	reportInterval int64
	pollInterval   int64
}

var options Options

func parseFlags() {
	/*Получаем параметры из командной строки*/
	flag.StringVar(&options.serverEndpoint, "a", "http://localhost:8080", "server address and port")
	flag.Int64Var(&options.reportInterval, "r", 10, "report interval")
	flag.Int64Var(&options.pollInterval, "p", 2, "poll interval")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		options.serverEndpoint = addr
	}

	if interval, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		if val, err := utils.StrToInt64(interval); err == nil {
			options.reportInterval = val
		}
	}

	if interval, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		if val, err := utils.StrToInt64(interval); err == nil {
			options.pollInterval = val
		}
	}

	if !strings.HasPrefix(options.serverEndpoint, "http") && !strings.HasPrefix(options.serverEndpoint, "https") {
		options.serverEndpoint = "http://" + options.serverEndpoint
	}
}
