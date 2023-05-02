package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/server/handlers"
	"github.com/AntonPashechko/yametrix/internal/server/restorer/filerestorer"
	memstorage "github.com/AntonPashechko/yametrix/internal/storage/memstorage"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
)

func main() {
	parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runServer(ctx)
}

func runServer(ctx context.Context) {

	//Инициализируем синглтон логера
	logger.Initialize(options.logLevel)
	//Наш роутер
	router := chi.NewRouter()
	//Хранилище метрик
	storage := memstorage.NewMemStorage()
	//Цепляем rest обработчики
	metrixHandler := handlers.NewMetrixHandler(storage)
	metrixHandler.Register(router)

	//Работа по синхронизированию данных
	restorer := filerestorer.NewFileRestorer(storage, options.storePath)

	//При штатном завершении сервера все накопленные данные должны сохраняться
	defer func() {
		if err := restorer.Store(); err != nil {
			logger.Log.Error("shutdown cannot save metrics", zap.Error(err))
		}
	}()

	//делаем restore если просят
	if options.restore {
		if err := restorer.Restore(); err != nil {
			logger.Log.Error("cannot restore metrics", zap.String("file", options.storePath), zap.Error(err))
		}
	}

	var storeScheduler scheduler.Scheduler
	defer func() {
		//Стопаем если вообще был запущен
		if storeScheduler != (scheduler.Scheduler{}) {
			storeScheduler.Stop()
		}
	}()
	/*Если периодичность сохранения задана - запускаем шедулер*/
	if options.storeInterval != 0 {
		storeScheduler = scheduler.NewScheduler(int64(options.storeInterval), restorer)
		go storeScheduler.Start()
	} else {
		//Синхронизацию при storeInterval = 0 будем проводить через Middleware синхронно
		//Передадим store handler`у,
		metrixHandler.SetRestorer(restorer)
	}

	//Запускаем сервер
	server := &http.Server{
		Addr:    options.endpoint,
		Handler: router,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	logger.Log.Info("Running server", zap.String("address", options.endpoint))

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}
}
