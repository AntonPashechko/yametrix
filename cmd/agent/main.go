package main

import (
	"context"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/agent/client"
	"github.com/AntonPashechko/yametrix/internal/agent/metrix"
	"github.com/AntonPashechko/yametrix/internal/agent/shaduller"
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
	pollShaduller := shaduller.NewShaduller(options.pollInterval, updateWorker)
	go pollShaduller.Start()

	/*Запуск шадуллера отправки метрик на сервер*/
	reportShaduller := shaduller.NewShaduller(options.reportInterval, sendWorker)
	go reportShaduller.Start()

	<-ctx.Done()

	/*Остановка шадуллера*/
	pollShaduller.Stop()
	reportShaduller.Stop()
}
