package main

import (
	"net/http"

	"go.uber.org/zap"
)

func addRoutes(mux *http.ServeMux, logger *zap.Logger) {
	mux.Handle("POST /api/v1/logs", HandleLogsIngestion(logger))
}
