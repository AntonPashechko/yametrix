// Package config предназначен для инициализации конфигурации клиента.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AntonPashechko/yametrix/pkg/utils"
)

// Config содержит список параметров для работы клиента.
type Config struct {
	ServerEndpoint string //эндпонт сервера
	SignKey        string //ключ подписи для контроля целостности запроса/ответа
	CryptoKey      string //путь до файла с публичным ключом сервера для шифрования данных
	ConfigJson     string //путь до файла с json конфигурацией
	ReportInterval int64  //интервал отправки обновленных метрик
	PollInterval   int64  //интервал обновления метрик
}

// formJson дополняет отсутствующие параметры из json
func (m *Config) formJson() error {

	data, err := os.ReadFile(m.ConfigJson)
	if err != nil {
		return fmt.Errorf("cannot read json config: %w", err)
	}

	var settings map[string]interface{}

	err = json.Unmarshal(data, &settings)
	if err != nil {
		return fmt.Errorf("cannot unmarshal json settings: %w", err)
	}

	for stype, value := range settings {
		switch stype {
		case "address":
			if m.ServerEndpoint == `` {
				m.ServerEndpoint = value.(string)
			}
		case "report_interval":
			if m.ReportInterval == 0 {
				duration, err := time.ParseDuration(value.(string))
				if err != nil {
					return fmt.Errorf("bad json param 'report_interval': %w", err)
				}
				m.ReportInterval = int64(duration.Seconds())
			}
		case "poll_interval":
			if m.PollInterval == 0 {
				duration, err := time.ParseDuration(value.(string))
				if err != nil {
					return fmt.Errorf("bad json param 'poll_interval': %w", err)
				}
				m.PollInterval = int64(duration.Seconds())
			}
		case "sign_key":
			if m.SignKey == `` {
				m.SignKey = value.(string)
			}
		case "crypto_key":
			if m.CryptoKey == `` {
				m.CryptoKey = value.(string)
			}
		}
	}

	return nil
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
	flag.StringVar(&cfg.ConfigJson, "c", "", "json config")
	flag.StringVar(&cfg.ConfigJson, "config", "", "json config")

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

	if config, exist := os.LookupEnv("CONFIG"); exist {
		cfg.ConfigJson = config
	}

	if cfg.ConfigJson != `` {
		err := cfg.formJson()
		if err != nil {
			return nil, fmt.Errorf("cannot full setting from json config: %w", err)
		}
	}

	if !strings.HasPrefix(cfg.ServerEndpoint, "http") && !strings.HasPrefix(cfg.ServerEndpoint, "https") {
		cfg.ServerEndpoint = "http://" + cfg.ServerEndpoint
	}

	return cfg, nil
}
