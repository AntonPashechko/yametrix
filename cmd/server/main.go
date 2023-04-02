package main

import (
	"context"
	"fmt"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/handlers/metrix"
	memstorage "github.com/AntonPashechko/yametrix/internal/storage/mem_storage"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runServer(ctx)
}

func runServer(ctx context.Context) {

	storage := memstorage.NewMemStorage()

	router := http.NewServeMux()

	metrixHandler := metrix.Handler{Storage: storage}
	metrixHandler.Register(router)

	server := &http.Server{
		Addr:    "localhost:8080",
		Handler: router,
	}

	go func() {
		err := server.ListenAndServe()
		if err != nil {
			panic(err)
		}
	}()

	<-ctx.Done()

	if err := server.Shutdown(context.Background()); err != nil {
		fmt.Printf("shutdown: %s", err.Error())
	}
}
