package http

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/ArtamonovEvgenii/grpc-course-2025/config"
	httpMiddleware "github.com/ArtamonovEvgenii/grpc-course-2025/internal/infrastructure/transport/http/middleware"
	grpcv1 "github.com/ArtamonovEvgenii/grpc-course-2025/pkg/api/notes/v1"
)

func GRPCGatewayHandler(
	ctx context.Context,
	grpcCfg config.GRPCServer,
) (http.Handler, error) {
	conn, err := grpc.NewClient(
		net.JoinHostPort(grpcCfg.Host, strconv.Itoa(grpcCfg.Port)),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("create grpc client: %w", err)
	}

	gatewayMux := runtime.NewServeMux(
	//runtime.WithIncomingHeaderMatcher(httptransport.CustomIncomingHeaderMatcher),
	//runtime.WithOutgoingHeaderMatcher(httptransport.CustomOutgoingHeaderMatcher),
	)

	err = grpcv1.RegisterNotesAPIHandler(ctx, gatewayMux, conn)
	if err != nil {
		return nil, fmt.Errorf("register grpc gateway: %w", err)
	}

	mux := httpMiddleware.CORSMiddleware(gatewayMux)

	return mux, nil
}
