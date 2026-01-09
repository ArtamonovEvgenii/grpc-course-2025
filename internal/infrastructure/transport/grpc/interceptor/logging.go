package interceptor

import (
	"context"
	"log/slog"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/logging"
	"google.golang.org/grpc"
)

// newInterceptorLogger adapts slog logger to interceptor logger.
// This code is simple enough to be copied and not imported.
// source: https://github.com/grpc-ecosystem/go-grpc-middleware/blob/ab2131d954af9580c1b49a3d9475f6adbe5de9d3/interceptors/logging/examples/slog/example_test.go#L17
func newInterceptorLogger(l *slog.Logger) logging.Logger {
	return logging.LoggerFunc(func(ctx context.Context, lvl logging.Level, msg string, fields ...any) {
		l.Log(ctx, slog.Level(lvl), msg, fields...)
	})
}

func NewLoggingUnaryInterceptor(lgr *slog.Logger) grpc.UnaryServerInterceptor {
	opts := []logging.Option{
		logging.WithLogOnEvents(logging.FinishCall),
	}

	return logging.UnaryServerInterceptor(newInterceptorLogger(lgr), opts...)
}

func NewLoggingStreamInterceptor(lgr *slog.Logger) grpc.StreamServerInterceptor {
	lgr = lgr.With(slog.String("component", "grpc_stream_interceptor"))

	return func(srv any, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			lgr:          lgr,
		}

		return handler(srv, wrappedStream)
	}
}

type wrappedServerStream struct {
	grpc.ServerStream
	lgr *slog.Logger
}

func (w *wrappedServerStream) SendMsg(m interface{}) error {
	w.lgr.Debug("server send", slog.Any("data", m))

	// Call the original Send method
	return w.ServerStream.SendMsg(m)
}

func (w *wrappedServerStream) RecvMsg(m interface{}) error {
	err := w.ServerStream.RecvMsg(m)

	w.lgr.Debug("server got", slog.Any("data", m))

	return err
}
