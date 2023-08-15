package main

import (
	"context"
	"fmt"
	"log"
	"os/signal"
	"sync"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/agent/config"
	"github.com/AntonPashechko/yametrix/internal/agent/sender"
	"github.com/AntonPashechko/yametrix/internal/agent/updater"
	"github.com/AntonPashechko/yametrix/internal/models"
	"github.com/AntonPashechko/yametrix/internal/sign"
)

var (
	buildVersion string = "N/A"
	buildDate    string = "N/A"
	buildCommit  string = "N/A"
)

func main() {
	fmt.Printf("Build version: %s\n", buildVersion)
	fmt.Printf("Build date: %s\n", buildDate)
	fmt.Printf("Build commit: %s\n", buildCommit)

	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	//Инициализируем подписанта, если задан key
	if cfg.SignKey != `` {
		sign.Initialize([]byte(cfg.SignKey))
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runtime := updater.NewRuntimeMetricsProducer(cfg)
	another := updater.NewAnotherMetricsProducer(cfg)
	consumer := sender.NewMetricsConsumer(cfg)

	metricCh := make(chan models.MetricDTO)

	var wg sync.WaitGroup
	wg.Add(3)

	go runtime.Work(ctx, &wg, metricCh)
	go another.Work(ctx, &wg, metricCh)
	go consumer.Work(ctx, &wg, metricCh)

	wg.Wait()
}
