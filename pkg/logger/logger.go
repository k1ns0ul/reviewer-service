package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
)

const (
	LevelDebug = slog.LevelDebug
	LevelInfo  = slog.LevelInfo
	LevelWarn  = slog.LevelWarn
	LevelError = slog.LevelError
)

type Config struct {
	Level       slog.Level
	Format      string
	Output      io.Writer
	ServiceName string
	Version     string
	Environment string
	AddSource   bool
}

func New(cfg Config) *slog.Logger {
	if cfg.Output == nil {
		cfg.Output = os.Stdout
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "reviewer-service"
	}
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
	}
	if cfg.Environment == "" {
		cfg.Environment = "development"
	}

	baseAttrs := []slog.Attr{
		slog.String("service", cfg.ServiceName),
		slog.String("version", cfg.Version),
		slog.String("env", cfg.Environment),
	}

	handlerOpts := &slog.HandlerOptions{
		Level:     cfg.Level,
		AddSource: cfg.AddSource,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(cfg.Output, handlerOpts)
	} else {
		handler = slog.NewTextHandler(cfg.Output, handlerOpts)
	}

	handler = handler.WithAttrs(baseAttrs)

	return slog.New(handler)
}

func NewDefault() *slog.Logger {
	return New(Config{
		Level:       LevelDebug,
		Format:      "text",
		ServiceName: "reviewer-service",
		Version:     "1.0.0",
		Environment: "development",
		AddSource:   true,
	})
}

func NewProduction() *slog.Logger {
	return New(Config{
		Level:       LevelInfo,
		Format:      "json",
		ServiceName: "reviewer-service",
		Version:     "1.0.0",
		Environment: "production",
		AddSource:   false,
	})
}

type loggerKey struct{}

func WithLogger(ctx context.Context, l *slog.Logger) context.Context {
	return context.WithValue(ctx, loggerKey{}, l)
}

func FromContext(ctx context.Context) *slog.Logger {
	if l, ok := ctx.Value(loggerKey{}).(*slog.Logger); ok {
		return l
	}
	return slog.Default()
}

func WithRequestID(l *slog.Logger, requestID string) *slog.Logger {
	return l.With(slog.String("request_id", requestID))
}

func WithUserID(l *slog.Logger, userID string) *slog.Logger {
	return l.With(slog.String("user_id", userID))
}
