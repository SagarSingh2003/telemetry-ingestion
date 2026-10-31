package main

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

type ResponseWrapper struct {
	http.ResponseWriter
	code int
}

func (r *ResponseWrapper) WriteHeader(code int) {
	r.code = code
	r.ResponseWriter.WriteHeader(code)
}

func PanicMiddlewaare(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverPanic(logger, w)
		next.ServeHTTP(w, r)
	})
}

func TracingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
	})
}

func LoggingMiddleware(logger *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		method := r.Method
		if method == "" {
			method = "GET"
		}
		startTime := time.Now()
		rw := &ResponseWrapper{ResponseWriter: w, code: http.StatusOK}
		next.ServeHTTP(rw, r)

		logger.Info(
			"Request Ended",
			zap.String("method", method),
			zap.Int("status_code", rw.code),
			zap.String("path", r.URL.Path),
			zap.Int64("duration_ns", time.Since(startTime).Nanoseconds()),
		)
	})
}
