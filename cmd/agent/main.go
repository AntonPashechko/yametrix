package main

import (
	"context"
	"os/signal"
	"syscall"
	"time"

	"github.com/AntonPashechko/yametrix/internal/agent/client"
	"github.com/AntonPashechko/yametrix/internal/agent/metrix"
	"github.com/AntonPashechko/yametrix/internal/agent/shaduller"
)

const (
	pollInterval   = 2
	reportInterval = 10

	endpoint = "http://localhost:8080"
)

func main() {

	/*Для реализации Graceful Shutdown*/
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runAgent(ctx)
}

func runAgent(ctx context.Context) {

	runtimeMetrix := metrix.NewRuntimeMetrix()

	updateWorker := metrix.NewUpdateMetrixWorker(runtimeMetrix)

	metrixHTTPClient := client.NewMetrixClient(runtimeMetrix, endpoint)

	sendWorker := client.NewSendMetrixWorker(metrixHTTPClient)

	/*Запуск шадуллера обновления метрик*/
	pollShaduller := shaduller.NewShaduller(pollInterval*time.Second, updateWorker)
	go pollShaduller.Start()

	/*Запуск шадуллера отправки метрик на сервер*/
	reportShaduller := shaduller.NewShaduller(reportInterval*time.Second, sendWorker)
	go reportShaduller.Start()

	<-ctx.Done()

	/*Остановка шадуллера*/
	pollShaduller.Stop()
	reportShaduller.Stop()
}
