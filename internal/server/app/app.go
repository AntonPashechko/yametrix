// Package app нужен для контроля жизненного цикла сервера.
package app

import (
	"fmt"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/AntonPashechko/yametrix/internal/storage/sqlstorage"
	"github.com/AntonPashechko/yametrix/internal/trustedsubnets"
)

const (
	shutdownTime = 5 * time.Second // время на shutdown
)

type App interface {
	Run()
	ServerDone() <-chan struct{}
	Shutdown() error
}

// Create создает экземпляр App.
func Create(cfg *config.Config) (App, error) {

	var storage storage.MetricsStorage
	if cfg.DataBaseDNS != "" {
		var err error
		storage, err = sqlstorage.NewStorage(cfg.DataBaseDNS)
		if err != nil {
			return nil, fmt.Errorf("cannot create db store: %w", err)
		}

	} else {
		//Хранилище метрик в памяти
		memStorage := memstorage.NewStorage()
		//Сторер
		restorer.Initialize(memStorage, cfg)

		storage = memStorage
	}

	if cfg.TrustedSubnet != `` {
		if err := trustedsubnets.Initialize(cfg.TrustedSubnet); err != nil {
			return nil, fmt.Errorf("cannot initialize trusted subnets: %w", err)
		}
	}

	if cfg.AppType == "http" {
		return CreateHTTPApp(storage, cfg)
	} else if cfg.AppType == "grpc" {
		return CreateGRPCApp(storage, cfg)
	} else {
		return nil, fmt.Errorf("unknown app type: %s", cfg.AppType)
	}
}
