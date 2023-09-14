package trustedsubnets

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"
)

func Interceptor(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
	//Если инициализирован объект для проверки подсети
	if MetricsSubnetChecker != nil {
		var clientip string
		if md, ok := metadata.FromIncomingContext(ctx); ok {
			values := md.Get("X-Real-IP")
			if len(values) > 0 {
				clientip = values[0]
			}
		}
		if len(clientip) == 0 {
			return nil, status.Error(codes.Unauthenticated, "missing IP")
		}

		//Проверяем, что в диапазоне cidr
		if !MetricsSubnetChecker.checkIp(clientip) {
			return nil, status.Error(codes.Unauthenticated, "client ip is not in CIDR range")
		}
	}

	return handler(ctx, req)
}
