package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"
)

type ServerConfig struct {
	Host string
	Port string
}

func Run(config ServerConfig) error {
	baseLogger := zap.Must(getBaseLogger())
	defer baseLogger.Sync()

	handler := NewServerHandler(
		baseLogger,
	)

	httpServer := http.Server{
		Addr:    net.JoinHostPort(config.Host, config.Port),
		Handler: handler,
	}

	quit, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	errGo, ctx := errgroup.WithContext(context.Background())
	errGo.Go(func() error {
		<-quit.Done()

		if err := GracefulShutDown(ctx, &httpServer); err != nil {
			baseLogger.Error(
				"graceful shutdown failed",
				zap.Error(err),
			)
			return err
		}
		return nil
	})

	errGo.Go(func() error {
		log.Printf("listening on %s\n", httpServer.Addr)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			baseLogger.Error("error occured while listening", zap.Error(err))
			stop()
		}
		return nil
	})

	if err := errGo.Wait(); err != nil {
		return err
	}
	return nil
}

func NewServerHandler(logger *zap.Logger) http.Handler {
	mux := http.NewServeMux()
	addRoutes(
		mux,
		logger,
	)
	var handler http.Handler = mux
	handler = LoggingMiddleware(logger, handler)
	handler = PanicMiddlewaare(logger, handler)
	return handler
}

func GracefulShutDown(parentctx context.Context, srv *http.Server) error {
	log.Printf("gracefully shutting down")
	ctx, cancel := context.WithTimeout(parentctx, 10*time.Second)
	defer cancel()
	err := srv.Shutdown(ctx)
	return err
}
