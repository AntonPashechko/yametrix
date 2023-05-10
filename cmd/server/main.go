package main

import (
	"log"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/app"
	"github.com/AntonPashechko/yametrix/internal/server/config"
)

func main() {
	//Инициализируем синглтон логера
	if err := logger.Initialize("info"); err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	cfg, err := config.LoadServerConfig()
	if err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	//Ошибки пока неоткуда взяться
	app := app.Create(cfg)

	go app.Run()

	logger.Info("Running server: address %s", cfg.Endpoint)

	<-app.ServerDone()

	if err := app.Shutdown(); err != nil {
		logger.Error("Server shutdown failed: %s", err)
	}

	logger.Info("Server has benn shutdown")
}
