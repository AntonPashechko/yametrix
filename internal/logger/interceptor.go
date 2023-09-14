package logger

import (
	"context"
	"time"

	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {

	start := time.Now()

	//Вызов целевого handler
	resp, err := handler(ctx, req)

	duration := time.Since(start)

	// Логируем данные запроса и результат
	log.Info("",
		zap.String("method", info.FullMethod),
		zap.Duration("duration", duration),
		zap.Error(err),
	)
	return resp, err
}
