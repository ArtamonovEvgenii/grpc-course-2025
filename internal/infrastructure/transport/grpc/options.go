package grpc

import (
	"log/slog"

	grpcmiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	"google.golang.org/grpc"

	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure/transport/grpc/interceptor"
)

func ServerOption(
	lgr *slog.Logger,
) grpc.ServerOption {
	grpcServerOption := grpc.UnaryInterceptor(
		grpcmiddleware.ChainUnaryServer(
			interceptor.NewRecoveryUnaryInterceptor(lgr),
			interceptor.NewLoggingUnaryInterceptor(lgr),
		))

	return grpcServerOption
}
