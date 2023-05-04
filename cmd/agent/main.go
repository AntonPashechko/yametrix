package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/agent/sender"
	"github.com/AntonPashechko/yametrix/internal/agent/updater"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
)

func main() {

	/*Для реализации Graceful Shutdown*/
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg := new(config.Config)
	parseFlags(cfg)

	storage := memstorage.NewMemStorage()

	/*Запуск шедулера обновления метрик*/
	updateWorker := updater.NewUpdateMetrixWorker(storage)
	pollScheduler := scheduler.NewScheduler(cfg.PollInterval, updateWorker)
	defer pollScheduler.Stop()
	go pollScheduler.Start()

	/*Запуск шедулера отправки метрик на сервер*/
	sendWorker := sender.NewHTTPSendWorker(storage, cfg.ServerEndpoint)
	reportScheduler := scheduler.NewScheduler(cfg.ReportInterval, sendWorker)
	defer reportScheduler.Stop()
	go reportScheduler.Start()

	<-ctx.Done()
}
