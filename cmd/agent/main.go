package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/agent/client"
	"github.com/AntonPashechko/yametrix/internal/agent/metrix"
	"github.com/AntonPashechko/yametrix/internal/agent/scheduler"
)

func main() {

	/*Для реализации Graceful Shutdown*/
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runAgent(ctx)
}

func runAgent(ctx context.Context) {

	parseFlags()

	runtimeMetrix := metrix.NewRuntimeMetrix()

	updateWorker := metrix.NewUpdateMetrixWorker(runtimeMetrix)

	metrixHTTPClient := client.NewMetrixClient(runtimeMetrix, options.serverEndpoint)

	sendWorker := client.NewSendMetrixWorker(metrixHTTPClient)

	/*Запуск шадуллера обновления метрик*/
	pollScheduler := scheduler.NewScheduler(options.pollInterval, updateWorker)
	defer pollScheduler.Stop()
	go pollScheduler.Start()

	/*Запуск шадуллера отправки метрик на сервер*/
	reportScheduler := scheduler.NewScheduler(options.reportInterval, sendWorker)
	defer reportScheduler.Stop()
	go reportScheduler.Start()

	<-ctx.Done()
}
