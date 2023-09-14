package app

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/compress"
	"github.com/AntonPashechko/yametrix/internal/encrypt"
	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/handlers"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/sign"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/trustedsubnets"
	"github.com/go-chi/chi/v5"
)

// HTTPApp управляет жизненным циклом Http сервера.
type HTTPApp struct {
	server     *http.Server           // экземпляр нашего http сервиса
	storage    storage.MetricsStorage // хранилище метрик
	notifyStop context.CancelFunc     // cancel функция для вызова stop сигнала
}

// Run запускает сервис в работу.
func (m *HTTPApp) Run() {
	if err := m.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("cannot listen: %s\n", err)
	}
}

// ServerDone возвращает канал по которому определяется признак завершения работы.
func (m *HTTPApp) ServerDone() <-chan struct{} {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	m.notifyStop = stop
	return ctx.Done()
}

// Shutdown корректно останавливает сервис.
func (m *HTTPApp) Shutdown() error {
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

// CreateHTTPApp создает экземпляр HttpApp.
func CreateHTTPApp(storage storage.MetricsStorage, cfg *config.Config) (*HTTPApp, error) {

	//Наш роутер, регистрируем хэндлеры
	router := chi.NewRouter()
	//Подключаем middleware логирования
	router.Use(logger.Middleware)
	//Подключаем middleware декомпрессии
	router.Use(compress.Middleware)
	//Для контроля ip клиента
	router.Use(trustedsubnets.Middleware)

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

	return &HTTPApp{
		server: &http.Server{
			Addr:    cfg.Endpoint,
			Handler: router,
		},
		storage: storage,
	}, nil
}
