package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/handlers"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	memstorage "github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-chi/chi/v5"
)

func main() {
	cfg := new(config.Config)
	if err := parseFlags(cfg); err != nil {
		log.Fatalf("cannot parse config: %s\n", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	//Инициализируем синглтон логера
	logger.Initialize(cfg.LogLevel)

	//Хранилище метрик
	storage := memstorage.NewMemStorage()

	//Наш роутер, регистрируем хэндлеры
	router := chi.NewRouter()
	metrixHandler := handlers.NewMetrixHandler(storage)
	metrixHandler.Register(router)

	//Работа по синхронизированию данных
	restorer.Initialize(storage, restorer.FileRestorer, cfg)
	defer restorer.Shutdown()

	//Запускаем сервер
	server := &http.Server{
		Addr:    cfg.Endpoint,
		Handler: router,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("cannot listen: %s\n", err)
		}
	}()

	logger.Info("Running server: address %s", cfg.Endpoint)

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failed:%s\n", err)
	}
}
