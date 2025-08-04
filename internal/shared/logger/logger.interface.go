package logger

import (
	"context"
	"log/slog"
	"os"

	"neuro.app.jordi/internal/shared/config"
)

type Logger interface {
	Error(ctx context.Context, msg string)
	Debug(ctx context.Context, msg string)
	Info(ctx context.Context, msg string)
}

type baseLogger struct {
	logger *slog.Logger
}

// --- Local Logger (human-readable) ---
type LocalLogger struct {
	*baseLogger
}

func NewLocalLogger() *LocalLogger {
	handler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelDebug, // Show debug in local
	})
	return &LocalLogger{&baseLogger{slog.New(handler)}}
}

// --- Production Logger (JSON) ---
type ProdLogger struct {
	*baseLogger
}

func NewProdLogger() *ProdLogger {
	handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo, // Debug logs off by default
	})
	return &ProdLogger{&baseLogger{slog.New(handler)}}
}

// --- BaseLogger methods ---
func (b *baseLogger) Error(ctx context.Context, msg string) {
	b.logger.ErrorContext(ctx, msg)
}

func (b *baseLogger) Debug(ctx context.Context, msg string) {
	b.logger.DebugContext(ctx, msg)
}

func (b *baseLogger) Info(ctx context.Context, msg string) {
	b.logger.InfoContext(ctx, msg)
}

// --- Factory function ---
func NewLogger() Logger {
	env := os.Getenv("environment") // set ENV=production in AWS Fargate
	if env == config.EnvProd {
		return NewProdLogger()
	}
	return NewLocalLogger()
}
