package main

import (
	"fmt"
	"log"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/app"
	"github.com/AntonPashechko/yametrix/internal/server/config"
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

	//Инициализируем синглтон логера
	if err := logger.Initialize("info"); err != nil {
		log.Fatalf("cannot initialize logger: %s\n", err)
	}

	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	app, err := app.Create(cfg)
	if err != nil {
		logger.Error("cannot create app: %s", err)
		return
	}

	go app.Run()

	logger.Info("Running server: address %s", cfg.Endpoint)

	<-app.ServerDone()

	if err := app.Shutdown(); err != nil {
		logger.Error("Server shutdown failed: %s", err)
	}

	logger.Info("Server has been shutdown")
}
