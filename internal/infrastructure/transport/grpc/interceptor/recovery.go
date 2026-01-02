package interceptor

import (
	"log/slog"
	"runtime/debug"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/recovery"
)

func NewRecoveryUnaryInterceptor(lgr *slog.Logger) grpc.UnaryServerInterceptor {
	recoveryInterceptorFunc := func(p any) (err error) {
		lgr.Error("recovery from panic",
			slog.Any("panic", p),
			slog.Any("stack", string(debug.Stack())),
		)

		return status.Error(codes.Internal, "panic")
	}

	opts := []recovery.Option{
		recovery.WithRecoveryHandler(recoveryInterceptorFunc),
	}

	return recovery.UnaryServerInterceptor(opts...)
}
