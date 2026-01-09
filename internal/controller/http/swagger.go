package http

import (
	"net/http"

	swaggerAPI "github.com/ArtamonovEvgenii/grpc-course-2025/docs/api/notes/v1"
	"github.com/ArtamonovEvgenii/grpc-course-2025/third_party/swagger"
)

func ServeSwagger(mux *http.ServeMux) {
	swaggerStaticsHandler := http.StripPrefix("/swagger", http.FileServer(http.FS(swagger.Content)))
	mux.Handle("GET /swagger/", swaggerStaticsHandler)

	swaggerSpecsHandler := http.StripPrefix("/swagger/specs", http.FileServer(http.FS(swaggerAPI.Content)))
	mux.Handle("GET /swagger/specs/", swaggerSpecsHandler)
}
