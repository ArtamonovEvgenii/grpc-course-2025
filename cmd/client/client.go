package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	pb "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"
)

func main() {
	handler := slog.NewJSONHandler(os.Stderr, nil)
	lgr := slog.New(handler)

	lgr.Info("starting client")

	title, err := getNoteTitle("019b92a9-9e74-7efb-84ec-6b6ff051c089")
	if err != nil {
		if errors.Is(err, ErrCommonClientError) {
			lgr.Error("common client error", slog.String("error", err.Error()))
		} else if errors.Is(err, ErrNoteNotFound) {
			lgr.Error("note not found", slog.String("error", err.Error()))
		} else {
			lgr.Error("non typified error", slog.String("error", err.Error()))
		}
		os.Exit(1)
	}

	lgr.Info("request success", slog.String("title", title))
}

func getNoteTitle(noteUUID string) (string, error) {
	grpcClient, err := grpc.NewClient(
		"127.0.0.1:8082",
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return "", fmt.Errorf("create grpc client: %w", err)
	}
	defer func() {
		_ = grpcClient.Close()
	}()

	apiClient := pb.NewNotesAPIClient(grpcClient)

	md := metadata.New(map[string]string{"authorization": "bearer secret"})
	reqCtx := metadata.NewOutgoingContext(context.Background(), md)

	req := &pb.GetNoteRequest{Uuid: noteUUID}
	resp, err := apiClient.GetNote(reqCtx, req)
	if err != nil {
		return "", fmt.Errorf("get note: %w", transportToDomainError(err))
	}

	return resp.Title, nil
}

func transportToDomainError(err error) error {
	grpcStatus, exist := status.FromError(err)
	if !exist {
		return fmt.Errorf("%w: %w", ErrCommonClientError, err)
	}

	for _, s := range grpcStatus.Details() {
		if errorDetails, ok := s.(*pb.ErrorDetails); ok {
			switch errorDetails.Code {
			case pb.ErrorCode_ERROR_CODE_NOTE_NOT_FOUND:
				return fmt.Errorf("%w: %s: grpc code: %s", ErrNoteNotFound, errorDetails.Description, grpcStatus.Code().String())
			}
		}
	}

	return fmt.Errorf("%w: %w", ErrCommonClientError, err)
}
