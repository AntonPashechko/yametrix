package restorer

import (
	"context"

	"google.golang.org/grpc"
)

func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	//Вызов целевого handler
	resp, err := handler(ctx, req)

	//Синхронизируем
	if instance != nil {
		instance.store()
	}

	return resp, err
}
