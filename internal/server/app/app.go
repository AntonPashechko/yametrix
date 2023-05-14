package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/handlers"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/AntonPashechko/yametrix/internal/storage/sqlstorage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	shutdownTime = 5 * time.Second
)

type App struct {
	server     *http.Server
	storage    storage.MetricsStorage //Нужно как то закрыть базу при shutsown? TODO
	notifyStop context.CancelFunc
}

func Create(cfg *config.Config) (*App, error) {

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
		restorer.Initialize(memStorage, restorer.FileRestorer, cfg)

		storage = memStorage
	}

	//Наш роутер, регистрируем хэндлеры
	router := chi.NewRouter()
	//Подключаем middleware логирования
	router.Use(logger.Middleware)
	//Подключаем middleware декомпрессии
	router.Use(compress.Middleware)

	metricsHandler := handlers.NewMetricsHandler(storage)
	metricsHandler.Register(router)

	return &App{
		server: &http.Server{
			Addr:    cfg.Endpoint,
			Handler: router,
		},
		storage: storage,
	}, nil
}

func (m *App) Run() {
	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("cannot listen: %s\n", err)
	}
}

func (m *App) ServerDone() <-chan struct{} {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	m.notifyStop = stop
	return ctx.Done()
}

func (m *App) Shutdown() error {
	defer m.notifyStop()
	defer m.storage.Close()
	defer restorer.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
