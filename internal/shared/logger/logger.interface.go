// app/logging/logger.go
package logging

import "context"

type Logger interface {
	Info(ctx context.Context, msg string, fields ...map[string]any)
	Warn(ctx context.Context, msg string, fields ...map[string]any)
	Error(ctx context.Context, msg string, err error, fields ...map[string]any)
}
