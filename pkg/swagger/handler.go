package swagger

import (
	"embed"
	"net/http"

	"github.com/ArtamonovEvgenii/grpc-course-2025/third_party/swagger"
)

func Handler(swaggerSpecs embed.FS) http.Handler {
	mux := http.NewServeMux()

	swaggerStaticsHandler := http.StripPrefix("/swagger", http.FileServer(http.FS(swagger.Content)))
	mux.Handle("GET /swagger/", swaggerStaticsHandler)

	swaggerSpecsHandler := http.StripPrefix("/swagger/specs", http.FileServer(http.FS(swaggerSpecs)))
	mux.Handle("GET /swagger/specs/", swaggerSpecsHandler)

	return mux
}
