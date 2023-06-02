package config

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/AntonPashechko/yametrix/pkg/utils"
)

type Config struct {
	ServerEndpoint string
	ReportInterval int64
	PollInterval   int64
	SignKey        string
}

func LoadAgentConfig() (*Config, error) {
	cfg := new(Config)
	/*Получаем параметры из командной строки*/
	flag.StringVar(&cfg.ServerEndpoint, "a", "http://localhost:8080", "server address and port")
	flag.Int64Var(&cfg.ReportInterval, "r", 10, "report interval")
	flag.Int64Var(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.SignKey, "k", "", "sign key")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		cfg.ServerEndpoint = addr
	}

	if interval, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		val, err := utils.StrToInt64(interval)
		if err != nil {
			//Тут бы я продолжил с параметром по умолчанию... но да ладно
			return nil, fmt.Errorf("cannot parse REPORT_INTERVAL env: %w", err)
		}
		cfg.ReportInterval = val
	}

	if interval, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		val, err := utils.StrToInt64(interval)
		if err != nil {
			//Тут бы я продолжил с параметром по умолчанию... но да ладно
			return nil, fmt.Errorf("cannot parse POLL_INTERVAL env: %w", err)
		}
		cfg.PollInterval = val
	}

	if signKey, exist := os.LookupEnv("KEY"); exist {
		cfg.SignKey = signKey
		log.Printf("AGENT SIGN_KEY env: %s", signKey)
	}

	if !strings.HasPrefix(cfg.ServerEndpoint, "http") && !strings.HasPrefix(cfg.ServerEndpoint, "https") {
		cfg.ServerEndpoint = "http://" + cfg.ServerEndpoint
	}

	return cfg, nil
}
