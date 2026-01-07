package interceptor

import (
	"context"

	"github.com/grpc-ecosystem/go-grpc-middleware/v2/interceptors/auth"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func newAuthFunc(token string) func(ctx context.Context) (context.Context, error) {
	return func(ctx context.Context) (context.Context, error) {
		reqToken, err := auth.AuthFromMD(ctx, "bearer")
		if err != nil {
			return nil, err
		}

		if token != reqToken {
			return nil, status.Errorf(codes.Unauthenticated, "invalid auth token")
		}

		return ctx, nil
	}
}

func NewAuthUnaryInterceptor(token string) grpc.UnaryServerInterceptor {
	authFunc := newAuthFunc(token)

	return auth.UnaryServerInterceptor(authFunc)
}
