package main

import (
	"flag"
	"os"
	"strings"

	config "github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/pkg/utils"
)

func parseFlags(cfg *config.Config) {
	/*Получаем параметры из командной строки*/
	flag.StringVar(&cfg.ServerEndpoint, "a", "http://localhost:8080", "server address and port")
	flag.Int64Var(&cfg.ReportInterval, "r", 10, "report interval")
	flag.Int64Var(&cfg.PollInterval, "p", 2, "poll interval")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		cfg.ServerEndpoint = addr
	}

	if interval, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		if val, err := utils.StrToInt64(interval); err == nil {
			cfg.ReportInterval = val
		}
	}

	if interval, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		if val, err := utils.StrToInt64(interval); err == nil {
			cfg.PollInterval = val
		}
	}

	if !strings.HasPrefix(cfg.ServerEndpoint, "http") && !strings.HasPrefix(cfg.ServerEndpoint, "https") {
		cfg.ServerEndpoint = "http://" + cfg.ServerEndpoint
	}
}
