// infra/logging/slog_logger.go
package logging

import (
	"context"
	"log/slog"
	"os"
)

type SlogLogger struct {
	l *slog.Logger
}

func NewSlogLogger(env string) *SlogLogger {
	var h slog.Handler

	if env == "local" || env == "dev" {
		// En desarrollo: salida legible en texto
		h = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug})
	} else {
		// En producción (Fargate → CloudWatch): JSON estructurado
		h = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo})
	}

	return &SlogLogger{l: slog.New(h)}
}

// Helpers para añadir IDs del contexto (request_id, trace_id…)
func withCtx(ctx context.Context, attrs []any) []any {
	if v := ctx.Value("request_id"); v != nil {
		attrs = append(attrs, "request_id", v)
	}
	if v := ctx.Value("trace_id"); v != nil {
		attrs = append(attrs, "trace_id", v)
	}
	if v := ctx.Value("user_id"); v != nil {
		attrs = append(attrs, "user_id", v)
	}
	return attrs
}

func (s *SlogLogger) Info(ctx context.Context, msg string, fields ...map[string]any) {
	attrs := []any{}
	for _, f := range fields {
		for k, v := range f {
			attrs = append(attrs, k, v)
		}
	}
	s.l.Info(msg, withCtx(ctx, attrs)...)
}

func (s *SlogLogger) Warn(ctx context.Context, msg string, fields ...map[string]any) {
	attrs := []any{}
	for _, f := range fields {
		for k, v := range f {
			attrs = append(attrs, k, v)
		}
	}
	s.l.Warn(msg, withCtx(ctx, attrs)...)
}

func (s *SlogLogger) Error(ctx context.Context, msg string, err error, fields ...map[string]any) {
	attrs := []any{"error", err}
	for _, f := range fields {
		for k, v := range f {
			attrs = append(attrs, k, v)
		}
	}
	s.l.Error(msg, withCtx(ctx, attrs)...)
}
