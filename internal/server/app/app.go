// Package app нужен для контроля жизненного цикла сервера.
package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/encrypt"
	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/handlers"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/sign"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/AntonPashechko/yametrix/internal/storage/sqlstorage"
)

const (
	shutdownTime = 5 * time.Second // время на shutdown
)

// App управляет жизненным циклов сервера.
type App struct {
	server     *http.Server           // экземпляр нашего http сервиса
	storage    storage.MetricsStorage // хранилище метрик
	notifyStop context.CancelFunc     // cancel функция для вызова stop сигнала
}

// Create создает экземпляр App.
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
		restorer.Initialize(memStorage, cfg)

		storage = memStorage
	}

	//Наш роутер, регистрируем хэндлеры
	router := chi.NewRouter()
	//Подключаем middleware логирования
	router.Use(logger.Middleware)
	//Подключаем middleware декомпрессии
	router.Use(compress.Middleware)

	//Если задан ключ для подписи - инициализируем объект, добавляем Middleware
	if cfg.SignKey != `` {
		sign.Initialize([]byte(cfg.SignKey))
		router.Use(sign.Middleware)
	}

	if cfg.CryptoKey != `` {
		if err := encrypt.InitializeDecryptor(cfg.CryptoKey); err != nil {
			return nil, fmt.Errorf("cannot create encryptor: %w", err)
		}
		router.Use(encrypt.Middleware)
	}

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

// Run запускает сервис в работу.
func (m *App) Run() {
	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("cannot listen: %s\n", err)
	}
}

// ServerDone возвращает канал по которому определяется признак завершения работы.
func (m *App) ServerDone() <-chan struct{} {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	m.notifyStop = stop
	return ctx.Done()
}

// Shutdown корректно останавливает сервис.
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
