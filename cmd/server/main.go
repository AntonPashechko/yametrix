package main

import (
	"log"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/app"
	"github.com/AntonPashechko/yametrix/internal/server/config"
)

func main() {
	cfg, err := config.LoadAgentConfig()
	if err != nil {
		log.Fatalf("cannot load config: %s\n", err)
	}

	//Инициализируем синглтон логера
	if err := logger.Initialize(cfg.LogLevel); err != nil {
		//Тут я считаю, что без логера можно жить себе спокойно... ну да ладно
		log.Fatalf("cannot load config: %s\n", err)
	}

	//Ошибки пока неоткуда взяться
	app := app.Create(cfg)

	go app.Run()

	logger.Info("Running server: address %s", cfg.Endpoint)

	<-app.ServerDone()

	if err := app.Shutdown(); err != nil {
		log.Fatalf("Server shutdown failed:%s\n", err)
	}

	logger.Info("Server has benn shutdown")
}
