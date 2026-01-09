package server

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"

	"github.com/tmc/grpc-websocket-proxy/wsproxy"

	"github.com/ArtamonovEvgenii/grpc-course-2025/config"
	grpccontroller "github.com/ArtamonovEvgenii/grpc-course-2025/internal/controller/grpc"
	httpcontroller "github.com/ArtamonovEvgenii/grpc-course-2025/internal/controller/http"
	grpctransport "github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure/transport/grpc"
	httptransport "github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure/transport/http"
	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/repository/inmemory"
	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/usecase"
	grpcv1 "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"
)

var errRunCommand = fmt.Errorf("run command error")

type service interface {
	Run(ctx context.Context) error
}

func Run(ctx context.Context) error {
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("load config: %s\n", err)
		return errRunCommand
	}

	handler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelDebug})
	lgr := slog.New(handler)

	inMemoryStorage := inmemory.NewStorage()

	notesUsecase := usecase.NewNotes(lgr, inMemoryStorage)

	grpcController, err := grpccontroller.NewController(lgr, notesUsecase)
	if err != nil {
		lgr.Error("create grpc controller", slog.String("error", err.Error()))
		return errRunCommand
	}

	grpcServerOptions := grpctransport.ServerOptions(lgr, cfg.GRPCServer, cfg.Auth.Token)
	grpcServer, err := grpctransport.NewServer(lgr, cfg.GRPCServer, grpcServerOptions...)
	if err != nil {
		lgr.Error("create grpc server", slog.String("error", err.Error()))
		return errRunCommand
	}

	grpcv1.RegisterNotesAPIServer(grpcServer.Server(), grpcController)
	reflection.Register(grpcServer.Server())

	grpcGWHandler, err := httpcontroller.GRPCGatewayHandler(ctx, cfg.GRPCServer)
	if err != nil {
		lgr.Error("create grpc gateway", slog.String("error", err.Error()))
		return errRunCommand
	}

	swaggerHandler := httpcontroller.SwaggerHandler()

	mux := http.NewServeMux()
	mux.Handle("/", grpcGWHandler)
	mux.Handle("/swagger/", swaggerHandler)
	httpHandler := wsproxy.WebsocketProxy(mux)

	httpServer, err := httptransport.NewServer(lgr, cfg.HTTPServer, httpHandler)
	if err != nil {
		lgr.Error("create http server", slog.String("error", err.Error()))
		return errRunCommand
	}

	lgr.Info("service starting")

	err = runServices(ctx, grpcServer, httpServer)
	if err != nil {
		lgr.Error("service runtime", slog.String("error", err.Error()))
		return errRunCommand
	}

	return nil
}

func runServices(ctx context.Context, services ...service) error {
	g, egCtx := errgroup.WithContext(ctx)
	for _, s := range services {
		g.Go(func() error {
			return s.Run(egCtx)
		})
	}

	return g.Wait()
}
