package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	cfg    zap.Config
	logger *zap.Logger
}

type Config struct {
	Host string
	Port string
}

type IngestionService struct {
}

func getBaseLogger() (*zap.Logger, error) {
	var cfg = zap.Config{
		Level:            zap.NewAtomicLevel(),
		Encoding:         "json",
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:    "timestamp",
			LevelKey:   "level",
			MessageKey: "msg",
			EncodeTime: zapcore.ISO8601TimeEncoder,
		},
	}

	return cfg.Build()
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
		next.ServeHTTP(w, r)
		logger.Info(
			"Request Ended",
			zap.String("method", method),
			zap.String("path", r.URL.Path),
			zap.Int64("duration_ns", time.Since(startTime).Nanoseconds()),
		)
	})
}

func NewServerHandler(logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		logger,
	)
	var handler http.Handler = mux
	handler = LoggingMiddleware(logger, handler)
	return handler
}

// HandleLogsIngestion returns a handler that processes logs ingestion request of type ingestionService
func HandleLogsIngestion(logger *zap.Logger) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		msg := []byte("hello world")
		msg = append(msg, '\n')
		w.Header().Add("Content-Type", "text/plain")
		w.Write(msg)
	})
}

func addRoutes(mux *http.ServeMux, logger *zap.Logger) {
	mux.Handle("POST /api/v1/logs", HandleLogsIngestion(logger))
}

func GracefulShutDown(parentctx context.Context, srv *http.Server) error {
	log.Printf("gracefully shutting down")

	ctx, cancel := context.WithTimeout(parentctx, 10*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	if err != nil {
		return err
	}
	if err = ctx.Err(); err != nil {
		return err
	}
	return nil
}

func run(config Config) {
	baseLogger := zap.Must(getBaseLogger())
	defer baseLogger.Sync()

	handler := NewServerHandler(
		baseLogger,
	)

	httpServer := http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: handler,
	}

	quit, stop := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	defer stop()

	var wg sync.WaitGroup
	wg.Go(func() {
		<-quit.Done()

		if err := GracefulShutDown(context.Background(), &httpServer); err != nil {
			baseLogger.Error(
				"graceful shutdown failed",
				zap.Error(err),
			)
		}
	})

	go func() {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			baseLogger.Error("error occured while listening", zap.Error(err))
			stop()
		}
	}()

	wg.Wait()
}

func main() {
	config := Config{
		Host: "localhost",
		Port: "3000",
	}

	run(config)
	os.Exit(1)
}
