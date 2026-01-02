package server

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"golang.org/x/sync/errgroup"
	"google.golang.org/grpc/reflection"

	"github.com/ArtamonovEvgenii/grpc-course-2025/config"
	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/repository/inmemory"
	"github.com/ArtamonovEvgenii/grpc-course-2025/internal/usecase"

	grpcv1 "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"

	grpccontroller "github.com/ArtamonovEvgenii/grpc-course-2025/internal/controller/grpc"
	grpctransport "github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure/transport/grpc"
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

	handler := slog.NewJSONHandler(os.Stderr, nil)
	lgr := slog.New(handler)

	inMemoryStorage := inmemory.NewStorage()

	notesUsecase := usecase.NewNotes(lgr, inMemoryStorage)

	grpcController, err := grpccontroller.NewController(notesUsecase)
	if err != nil {
		lgr.Error("create grpc controller", slog.String("error", err.Error()))
		return errRunCommand
	}

	grpcServerOption := grpctransport.ServerOption(lgr)
	grpcServer, err := grpctransport.NewServer(lgr, cfg.GRPCServer, grpcServerOption)
	if err != nil {
		lgr.Error("create grpc server", slog.String("error", err.Error()))
		return errRunCommand
	}

	grpcv1.RegisterNotesServer(grpcServer.Server(), grpcController)
	reflection.Register(grpcServer.Server())

	lgr.Info("service starting")

	err = runServices(ctx, grpcServer)
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
