package grpc

import (
	"log/slog"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"

	"github.com/ArtamonovEvgenii/grpc-course-2025/config"
	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure/transport/grpc/interceptor"
)

func ServerOptions(
	lgr *slog.Logger,
	cfg config.GRPCServer,
	token string,
) []grpc.ServerOption {
	options := []grpc.ServerOption{
		grpc.Creds(insecure.NewCredentials()),
		grpc.KeepaliveParams(
			keepalive.ServerParameters{
				Time:    cfg.KeepAliveTime,
				Timeout: cfg.KeepAliveTimeout,
			},
		),
		grpc.MaxConcurrentStreams(uint32(cfg.MaxConcurrentStreams)),
		grpc.UnaryInterceptor(
			grpcmiddleware.ChainUnaryServer(
				interceptor.NewRecoveryUnaryInterceptor(lgr),
				interceptor.NewAuthUnaryInterceptor(token),
				interceptor.NewLoggingUnaryInterceptor(lgr),
			)),
	}

	return options
}
