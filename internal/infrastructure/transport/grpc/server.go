package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net"
	"strconv"

	"google.golang.org/grpc"

	"github.com/ArtamonovEvgenii/grpc-course-2025/config"
	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure"
)

type Server struct {
	server   *grpc.Server
	listener net.Listener
	lgr      *slog.Logger
	address  string
}

func NewServer(
	lgr *slog.Logger,
	cfg config.GRPCServer,
	options ...grpc.ServerOption,
) (*Server, error) {
	address := net.JoinHostPort(cfg.Host, strconv.Itoa(cfg.Port))

	lis, err := net.Listen("tcp", address)
	if err != nil {
		return nil, fmt.Errorf("listen on %s: %w", address, err)
	}

	grpcSrv := grpc.NewServer(options...)

	return &Server{
		lgr:      lgr.WithGroup("grpc.server"),
		server:   grpcSrv,
		listener: lis,
		address:  address,
	}, nil
}

func (s *Server) Server() *grpc.Server {
	return s.server
}

func (s *Server) Run(ctx context.Context) error {
	var err error
	errCh := make(chan error)

	go func() {
		s.lgr.Info(fmt.Sprintf("grpc server will be started at %s", s.address))
		errCh <- s.server.Serve(s.listener)
	}()

	select {
	case <-ctx.Done():
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), infrastructure.ServiceShutdownTimeout)
		defer shutdownCancel()

		err = s.shutdown(shutdownCtx)
		if err != nil {
			err = fmt.Errorf("grpc server shutdown error: %w", err)
		}
		s.lgr.Info("grpc server stopped, reason: context canceled")

	case err = <-errCh:
		err = fmt.Errorf("grpc server: %w", err)
		s.lgr.Warn("grpc server stopped, reason: error from channel")
	}

	return err
}

func (s *Server) shutdown(ctx context.Context) error {
	ok := make(chan struct{})
	// NOTE: GracefulStop will return instantly
	// when Stop it called, preventing this
	// goroutine from leaking.
	go func() {
		s.server.GracefulStop()
		close(ok)
	}()

	select {
	case <-ok:
		return nil
	case <-ctx.Done():
		s.server.Stop()
		return ctx.Err()
	}
}
