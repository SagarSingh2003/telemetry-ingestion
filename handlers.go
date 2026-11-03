package main

import (
	"net/http"

	"go.uber.org/zap"
)

// HandleLogsIngestion returns a handler that processes logs ingestion request of type ingestionService
func HandleLogsIngestion(logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := []byte("hello world")
		msg = append(msg, '\n')
		_, err := w.Write(msg)
		if err != nil {
			logger.Error("error occured while writing", zap.Error(err))
		}
	})
}
