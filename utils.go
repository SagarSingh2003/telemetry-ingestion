package main

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

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

func recoverPanic(logger *zap.Logger, w http.ResponseWriter) {
	if r := recover(); r != nil {
		logger.Error(
			"caught panic",
			zap.String("panic_info", fmt.Sprintf("%v", r)),
			zap.String("stack_trace", string(debug.Stack())),
		)
		w.WriteHeader(500)
	}
}
