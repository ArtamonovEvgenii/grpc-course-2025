package http

import (
	"context"
	"fmt"
	"net"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ArtamonovEvgenii/grpc-course-2025/config"
	grpcv1 "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"
)

func NewGatewayMux(
	ctx context.Context,
	grpcCfg config.GRPCServer,
) (*runtime.ServeMux, error) {
	conn, err := grpc.NewClient(
		net.JoinHostPort(grpcCfg.Host, strconv.Itoa(grpcCfg.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create grpc client: %w", err)
	}

	grpcGWMux := runtime.NewServeMux(
	//runtime.WithIncomingHeaderMatcher(httptransport.CustomIncomingHeaderMatcher),
	//runtime.WithOutgoingHeaderMatcher(httptransport.CustomOutgoingHeaderMatcher),
	)

	err = grpcv1.RegisterNotesAPIHandler(ctx, grpcGWMux, conn)
	if err != nil {
		return nil, fmt.Errorf("register grpc gateway: %w", err)
	}

	return grpcGWMux, nil
}
