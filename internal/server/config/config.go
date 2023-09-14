// Package config предназначен для инициализации конфигурации сервера.
package config

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AntonPashechko/yametrix/internal/logger"

	"github.com/AntonPashechko/yametrix/pkg/utils"
)

// Config содержит список параметров для работы сервера.
type Config struct {
	Endpoint      string // эндпоинт сервера
	StorePath     string // путь к файлу синхронизации метрик
	DataBaseDNS   string // строка подключения к БД
	SignKey       string // ключ подписи
	CryptoKey     string // путь до файла с приватным ключом сервера для расшифровывания данных
	ConfigJson    string // путь до файла с json конфигурацией
	TrustedSubnet string // строковое представление бесклассовой адресации (CIDR)
	AppType       string // тип сервера http или grpc
	StoreInterval uint64 // интервал синхронизации метрик (0 - синхронная запись)
	Restore       bool   // флаг синхронизации метрик из файла при запуске
}

// formJson дополняет отсутствующие параметры из json
func (m *Config) formFile() error {

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
			if m.Endpoint == `` {
				m.Endpoint = value.(string)
			}
		case "restore":
			if !m.Restore {
				m.Restore = value.(bool)
			}
		case "store_interval":
			if m.StoreInterval == 0 {
				duration, err := time.ParseDuration(value.(string))
				if err != nil {
					return fmt.Errorf("bad json param 'store_interval': %w", err)
				}
				m.StoreInterval = uint64(duration.Seconds())
			}
		case "store_file":
			if m.StorePath == `` {
				m.StorePath = value.(string)
			}
		case "database_dsn":
			if m.DataBaseDNS == `` {
				m.DataBaseDNS = value.(string)
			}
		case "sign_key":
			if m.SignKey == `` {
				m.SignKey = value.(string)
			}
		case "crypto_key":
			if m.CryptoKey == `` {
				m.CryptoKey = value.(string)
			}
		case "trusted_subnet":
			if m.TrustedSubnet == `` {
				m.TrustedSubnet = value.(string)
			}
		case "app_type":
			if m.AppType == `` {
				m.AppType = value.(string)
			}
		}
	}

	return nil
}

// newConfig создает экземпляр Config на онове опций в строковом представлении.
func newConfig(opt options) (*Config, error) {
	cfg := &Config{
		Endpoint:      opt.endpoint,
		StorePath:     opt.storePath,
		DataBaseDNS:   opt.dbDNS,
		SignKey:       opt.signKey,
		CryptoKey:     opt.сryptoKey,
		ConfigJson:    opt.configJson,
		TrustedSubnet: opt.trustedSubnet,
	}

	restore, err := strconv.ParseBool(opt.restore)
	if err != nil {
		return nil, fmt.Errorf("bad param RESTORE: %w", err)
	}
	cfg.Restore = restore

	//В тестах на гитхаб данный параметр от инкремента к инкременту задается по разному, или 10 или 10s
	//Буду тогда по очереди пытаться его разобрать, сперва как 10s
	duration, err := time.ParseDuration(opt.storeInterval)
	if err != nil {
		//Теперь как 10
		duration, err := utils.StrToInt64(opt.storeInterval)
		if err != nil {
			return nil, fmt.Errorf("bad param STORE_INTERVAL: %w", err)
		}
		cfg.StoreInterval = uint64(duration)
	} else {
		cfg.StoreInterval = uint64(duration.Seconds())
	}

	if cfg.ConfigJson != `` {
		err := cfg.formFile()
		if err != nil {
			return nil, fmt.Errorf("cannot full setting from json config: %w", err)
		}
	}

	return cfg, nil
}

// options содержит список параметров для работы сервера в строковом представлении.
type options struct {
	endpoint      string
	storeInterval string
	storePath     string
	restore       string
	dbDNS         string
	signKey       string
	сryptoKey     string
	configJson    string
	trustedSubnet string
}

// LoadServerConfig загружает настройки сервера из командной строки или переменных окружения.
func LoadServerConfig() (*Config, error) {

	logger.Info(strings.Join(os.Args, " "))
	var opt options

	/*Разбираем командную строку сперва в структуру только со string полями*/
	flag.StringVar(&opt.endpoint, "a", "localhost:8080", "address and port to run server")

	flag.StringVar(&opt.storeInterval, "i", "300s", "store metrics interval")
	flag.StringVar(&opt.storePath, "а", "/tmp/metrics-db.json", "store metrics path")

	flag.StringVar(&opt.restore, "r", "true", "is restore")
	flag.StringVar(&opt.dbDNS, "d", "", "db dns")

	flag.StringVar(&opt.signKey, "k", "", "sign key")
	flag.StringVar(&opt.сryptoKey, "crypto-key", "", "private crypto key")

	flag.StringVar(&opt.configJson, "c", "", "json config")
	flag.StringVar(&opt.configJson, "config", "", "json config")

	flag.StringVar(&opt.configJson, "t", "", "trusted subnet")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		opt.endpoint = addr
		logger.Info("ADDRESS env: %s", addr)
	}

	if storeIntStr, exist := os.LookupEnv("STORE_INTERVAL"); exist {
		opt.storeInterval = storeIntStr
		logger.Info("STORE_INTERVAL env: %s", storeIntStr)
	}

	if storePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		opt.storePath = storePath
		logger.Info("FILE_STORAGE_PATH env: %s", storePath)
	}

	if restore, exist := os.LookupEnv("RESTORE"); exist {
		opt.restore = restore
		logger.Info("RESTORE env: %s", restore)
	}

	if dns, exist := os.LookupEnv("DATABASE_DSN"); exist {
		logger.Info("DATABASE_DSN env: %s", dns)
		opt.dbDNS = dns
	}

	if signKey, exist := os.LookupEnv("KEY"); exist {
		logger.Info("SIGN_KEY env: %s", signKey)
		opt.signKey = signKey
	}

	if cryptoKey, exist := os.LookupEnv("CRYPTO_KEY"); exist {
		logger.Info("CRYPTO_KEY env: %s", cryptoKey)
		opt.сryptoKey = cryptoKey
	}

	if config, exist := os.LookupEnv("CONFIG"); exist {
		logger.Info("CONFIG env: %s", config)
		opt.configJson = config
	}

	if subnet, exist := os.LookupEnv("TRUSTED_SUBNET"); exist {
		logger.Info("TRUSTED_SUBNET env: %s", subnet)
		opt.trustedSubnet = subnet
	}

	return newConfig(opt)
}
