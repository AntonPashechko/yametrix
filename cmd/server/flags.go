package main

import (
	"flag"
	"log"
	"os"
	"strconv"

	"github.com/AntonPashechko/yametrix/pkg/utils"
)

type Options struct {
	endpoint      string
	logLevel      string
	storeInterval uint64 //0 - синхронная запись
	storePath     string
	restore       bool
}

var options Options

func parseFlags() {
	/*Разбираем командную строку*/
	flag.StringVar(&options.endpoint, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&options.logLevel, "l", "info", "log level")

	flag.Uint64Var(&options.storeInterval, "i", 300, "store metrics interval")
	flag.StringVar(&options.storePath, "а", "/tmp/metrics-db.json", "store metrics path")

	/*flag.Uint64Var(&options.storeInterval, "i", 10, "store metrics interval")
	flag.StringVar(&options.storePath, "а", "metrics-db.json", "store metrics path")*/

	flag.BoolVar(&options.restore, "r", true, "is restore")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		options.endpoint = addr
	}

	if lvl, exist := os.LookupEnv("LOG_LEVEL"); exist {
		options.logLevel = lvl
	}

	if storeIntStr, exist := os.LookupEnv("STORE_INTERVAL"); exist {
		interval, err := utils.StrToInt64(storeIntStr)
		if err != nil {
			log.Fatalf("bad Env STORE_INTERVAL: %s\n", err)
		}

		options.storeInterval = uint64(interval)
	}

	if storePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		options.storePath = storePath
	}

	if storePath, exist := os.LookupEnv("RESTORE "); exist {
		boolValue, err := strconv.ParseBool(storePath)
		if err != nil {
			log.Fatalf("bad Env RESTORE: %s\n", err)
		}
		options.restore = boolValue
	}
}
