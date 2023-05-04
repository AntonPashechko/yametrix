package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"

	config "github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

func parseFlags(cfg *config.Config) error {
	/*Разбираем командную строку*/
	flag.StringVar(&cfg.Endpoint, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&cfg.LogLevel, "l", "info", "log level")

	/*flag.Uint64Var(&cfg.StoreInterval, "i", 300, "store metrics interval")
	flag.StringVar(&cfg.StorePath, "а", "/tmp/metrics-db.json", "store metrics path")*/

	flag.Uint64Var(&cfg.StoreInterval, "i", 0, "store metrics interval")
	flag.StringVar(&cfg.StorePath, "а", "metrics-db.json", "store metrics path")

	flag.BoolVar(&cfg.Restore, "r", true, "is restore")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		cfg.Endpoint = addr
	}

	if lvl, exist := os.LookupEnv("LOG_LEVEL"); exist {
		cfg.LogLevel = lvl
	}

	if storeIntStr, exist := os.LookupEnv("STORE_INTERVAL"); exist {
		interval, err := utils.StrToInt64(storeIntStr)
		if err != nil {
			return fmt.Errorf("bad Env STORE_INTERVAL: %s", err)
		}

		cfg.StoreInterval = uint64(interval)
	}

	if storePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		cfg.StorePath = storePath
	}

	if storePath, exist := os.LookupEnv("RESTORE "); exist {
		boolValue, err := strconv.ParseBool(storePath)
		if err != nil {
			return fmt.Errorf("bad Env RESTORE: %s", err)
		}
		cfg.Restore = boolValue
	}

	return nil
}
