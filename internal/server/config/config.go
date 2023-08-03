// Пакет config предназначен для инициализации конфигурации сервера.
package config

import (
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
	StoreInterval uint64 // интервал синхронизации метрик (0 - синхронная запись)
	StorePath     string // путь к файлу синхронизации метрик
	Restore       bool   // флаг синхронизации метрик из файла при запуске
	DataBaseDNS   string // строка подключения к БД
	SignKey       string // ключ подписи
}

// newConfig создает экземпляр Config на онове опций в строковом представлении.
func newConfig(opt options) (*Config, error) {
	cfg := &Config{
		Endpoint:    opt.endpoint,
		StorePath:   opt.storePath,
		DataBaseDNS: opt.dbDNS,
		SignKey:     opt.signKey,
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

	return newConfig(opt)
}
