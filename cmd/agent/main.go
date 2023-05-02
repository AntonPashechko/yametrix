package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/agent/sender"
	"github.com/AntonPashechko/yametrix/internal/agent/updater"
	"github.com/AntonPashechko/yametrix/internal/scheduler"
	"github.com/AntonPashechko/yametrix/internal/storage/memstorage"
)

func main() {

	/*Для реализации Graceful Shutdown*/
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runAgent(ctx)
}

func runAgent(ctx context.Context) {

	parseFlags()

	storage := memstorage.NewMemStorage()

	/*Запуск шадуллера обновления метрик*/
	updateWorker := updater.NewUpdateMetrixWorker(storage)
	pollScheduler := scheduler.NewScheduler(options.pollInterval, updateWorker)
	defer pollScheduler.Stop()
	go pollScheduler.Start()

	/*Запуск шадуллера отправки метрик на сервер*/
	sendWorker := sender.NewHTTPSendWorker(storage, options.serverEndpoint)
	reportScheduler := scheduler.NewScheduler(options.reportInterval, sendWorker)
	defer reportScheduler.Stop()
	go reportScheduler.Start()

	<-ctx.Done()
}
