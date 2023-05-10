package app

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/handlers"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-chi/chi/v5"
	_ "github.com/jackc/pgx/v5/stdlib"
)

const (
	shutdownTime = 5 * time.Second
)

type App struct {
	server     *http.Server
	notifyStop context.CancelFunc
}

func Create(cfg *config.Config) *App {

	//Хранилище метрик
	storage := memstorage.NewMemStorage()

	db, err := sql.Open("pgx", cfg.DataBaseDNS)
	if err != nil {
		log.Fatalf("cannot create connection db: %s\n", err)
	}
	defer db.Close()

	//Сторер
	restorer.Initialize(storage, restorer.FileRestorer, cfg)

	//Наш роутер, регистрируем хэндлеры
	router := chi.NewRouter()
	metrixHandler := handlers.NewMetrixHandler(storage, db)
	metrixHandler.Register(router)

	app := &App{
		server: &http.Server{
			Addr:    cfg.Endpoint,
			Handler: router,
		},
	}

	return app
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
	defer restorer.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), shutdownTime)
	defer cancel()

	if err := m.server.Shutdown(ctx); err != nil {
		return fmt.Errorf("server shutdown failed: %w", err)
	}

	return nil
}
