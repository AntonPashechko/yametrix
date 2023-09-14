package app

import (
	"context"
	"log"
	"net"
	"os/signal"
	"syscall"

	"github.com/AntonPashechko/yametrix/internal/logger"
	"github.com/AntonPashechko/yametrix/internal/server/config"
	"github.com/AntonPashechko/yametrix/internal/server/metricsgrpc"
	"github.com/AntonPashechko/yametrix/internal/server/restorer"
	"github.com/AntonPashechko/yametrix/internal/storage"
	"github.com/AntonPashechko/yametrix/internal/trustedsubnets"
	"google.golang.org/grpc"
)

// GRPCApp управляет жизненным циклом GRPCApp сервера.
type GRPCApp struct {
	server     *grpc.Server
	storage    storage.MetricsStorage // хранилище метрик
	notifyStop context.CancelFunc     // cancel функция для вызова stop сигнала
	endpoint   string
}

// Run запускает сервис в работу.
func (m *GRPCApp) Run() {
	lis, err := net.Listen("tcp", m.endpoint)
	if err != nil {
		log.Fatalf("failed to listen: %s", err)
	}

	if err := m.server.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %s", err)
	}
}

// ServerDone возвращает канал по которому определяется признак завершения работы.
func (m *GRPCApp) ServerDone() <-chan struct{} {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	m.notifyStop = stop
	return ctx.Done()
}

// Shutdown корректно останавливает сервис.
func (m *GRPCApp) Shutdown() error {
	m.server.GracefulStop()
	m.notifyStop()

	return nil
}

// CreateGRPCApp создает экземпляр GRPCApp.
func CreateGRPCApp(storage storage.MetricsStorage, cfg *config.Config) (*GRPCApp, error) {

	//Создаем grpc сервис, регистрируем перехватчики
	grpcServer := grpc.NewServer(grpc.ChainUnaryInterceptor(logger.Interceptor, trustedsubnets.Interceptor, restorer.Interceptor))

	service := metricsgrpc.NewService(storage)
	metricsgrpc.RegisterMetricsServiceServer(grpcServer, &service)

	return &GRPCApp{
		server:   grpcServer,
		endpoint: cfg.Endpoint,
		storage:  storage,
	}, nil
}
