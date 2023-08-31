// Package config предназначен для инициализации конфигурации клиента.
package config

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/AntonPashechko/yametrix/pkg/utils"
)

// Config содержит список параметров для работы клиента.
type Config struct {
	ServerEndpoint string //эндпонт сервера
	SignKey        string //ключ подписи для контроля целостности запроса/ответа
	CryptoKey      string //путь до файла с публичным ключом сервера для шифрования данных
	ReportInterval int64  //интервал отправки обновленных метрик
	PollInterval   int64  //интервал обновления метрик
}

// LoadAgentConfig загружает настройки клиента из командной строки или переменных окружения.
func LoadAgentConfig() (*Config, error) {
	cfg := new(Config)
	/*Получаем параметры из командной строки*/
	flag.StringVar(&cfg.ServerEndpoint, "a", "http://localhost:8080", "server address and port")
	flag.Int64Var(&cfg.ReportInterval, "r", 10, "report interval")
	flag.Int64Var(&cfg.PollInterval, "p", 2, "poll interval")
	flag.StringVar(&cfg.SignKey, "k", "", "sign key")
	flag.StringVar(&cfg.CryptoKey, "crypto-key", "", "open crypto key")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		cfg.ServerEndpoint = addr
	}

	if interval, exist := os.LookupEnv("REPORT_INTERVAL"); exist {
		val, err := utils.StrToInt64(interval)
		if err != nil {
			return nil, fmt.Errorf("cannot parse REPORT_INTERVAL env: %w", err)
		}
		cfg.ReportInterval = val
	}

	if interval, exist := os.LookupEnv("POLL_INTERVAL"); exist {
		val, err := utils.StrToInt64(interval)
		if err != nil {
			return nil, fmt.Errorf("cannot parse POLL_INTERVAL env: %w", err)
		}
		cfg.PollInterval = val
	}

	if signKey, exist := os.LookupEnv("KEY"); exist {
		cfg.SignKey = signKey
	}

	if cryptoKey, exist := os.LookupEnv("CRYPTO_KEY"); exist {
		cfg.CryptoKey = cryptoKey
	}

	if !strings.HasPrefix(cfg.ServerEndpoint, "http") && !strings.HasPrefix(cfg.ServerEndpoint, "https") {
		cfg.ServerEndpoint = "http://" + cfg.ServerEndpoint
	}

	return cfg, nil
}
