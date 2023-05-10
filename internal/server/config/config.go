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

type Config struct {
	Endpoint      string
	LogLevel      string
	StoreInterval uint64 //0 - синхронная запись
	StorePath     string
	Restore       bool
	DataBaseDNS   string
}

func newConfig(opt options) (*Config, error) {
	cfg := &Config{
		Endpoint:    opt.endpoint,
		LogLevel:    opt.logLevel,
		StorePath:   opt.storePath,
		DataBaseDNS: opt.dbDNS,
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

type options struct {
	endpoint      string
	logLevel      string
	storeInterval string
	storePath     string
	restore       string
	dbDNS         string
}

func LoadServerConfig() (*Config, error) {

	logger.Info(strings.Join(os.Args, " "))
	var opt options

	/*Разбираем командную строку сперва в структуру только со string полями*/
	flag.StringVar(&opt.endpoint, "a", "localhost:8080", "address and port to run server")
	flag.StringVar(&opt.logLevel, "l", "info", "log level")

	flag.StringVar(&opt.storeInterval, "i", "300s", "store metrics interval")
	flag.StringVar(&opt.storePath, "а", "/tmp/metrics-db.json", "store metrics path")

	flag.StringVar(&opt.restore, "r", "true", "is restore")
	flag.StringVar(&opt.dbDNS, "d", "host=localhost user=postgres password=postgres dbname=postgres sslmode=disable", "db dns")

	flag.Parse()

	/*Но если заданы в окружении - берем оттуда*/
	if addr, exist := os.LookupEnv("ADDRESS"); exist {
		opt.endpoint = addr
	}

	if lvl, exist := os.LookupEnv("LOG_LEVEL"); exist {
		opt.logLevel = lvl
	}

	if storeIntStr, exist := os.LookupEnv("STORE_INTERVAL"); exist {
		opt.storeInterval = storeIntStr
	}

	if storePath, exist := os.LookupEnv("FILE_STORAGE_PATH"); exist {
		opt.storePath = storePath
	}

	if restore, exist := os.LookupEnv("RESTORE"); exist {
		opt.restore = restore
	}

	if storePath, exist := os.LookupEnv("DATABASE_DSN "); exist {
		opt.dbDNS = storePath
	}

	return newConfig(opt)
}
