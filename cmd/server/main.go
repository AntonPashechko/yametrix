package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/handlers/metrix"
	memstorage "github.com/AntonPashechko/yametrix/internal/storage/mem_storage"
	"github.com/go-chi/chi/v5"
)

func main() {
	parseFlags()

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	runServer(ctx)
}

func runServer(ctx context.Context) {

	storage := memstorage.NewMemStorage()

	router := chi.NewRouter()

	metrixHandler := metrix.NewMetrixHandler(storage)
	metrixHandler.Register(router)

	server := &http.Server{
		Addr:    endpoint,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("listen: %s\n", err)
	}

	/*go func() {
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	<-ctx.Done()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer func() {
		cancel()
	}()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server Shutdown Failed:%+v", err)
	}*/
}
