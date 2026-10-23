package logger

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type (
	logger struct {
		instance zerolog.Logger
	}

	Logger interface {
		Info(msg string, keysAndValues ...any)
		Warn(msg string, keysAndValues ...any)
		Error(msg string, err error, keysAndValues ...any)
		Debug(msg string, keysAndValues ...any)
		Trace(msg string, keysAndValues ...any)
		With(keysAndValues ...any) Logger
		Instance() zerolog.Logger
	}

	contextKey struct{}
)

const badKey = "!BADKEY"

func NewLogger(level string, prod bool) Logger {
	var writer io.Writer = os.Stdout

	if !prod {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	return newWithWriter(level, writer)
}

func NewWithWriter(level string, w io.Writer) Logger {
	return newWithWriter(level, w)
}

func newWithWriter(level string, w io.Writer) Logger {
	instance := zerolog.New(w).Level(toZerologLevel(level)).With().Timestamp().Logger()

	return &logger{
		instance: instance,
	}
}

func WithCtx(ctx context.Context, l Logger) context.Context {
	return context.WithValue(ctx, contextKey{}, l)
}

func FromCtx(ctx context.Context) Logger {
	if l, ok := ctx.Value(contextKey{}).(Logger); ok && l != nil {
		return l
	}

	return &logger{
		instance: zerolog.Nop(),
	}
}

func FromCtxOr(ctx context.Context, fallback Logger) Logger {
	if l, ok := ctx.Value(contextKey{}).(Logger); ok && l != nil {
		return l
	}

	return fallback
}

func fields(keysAndValues []any) map[string]any {
	out := make(map[string]any, len(keysAndValues)/2)

	for i := 0; i < len(keysAndValues); i += 2 {
		key, ok := keysAndValues[i].(string)
		if !ok {
			out[badKey] = keysAndValues[i]
			continue
		}

		if i+1 >= len(keysAndValues) {
			out[badKey] = key
			break
		}

		out[key] = keysAndValues[i+1]
	}

	return out
}

func (l *logger) Info(msg string, keysAndValues ...any) {
	l.instance.Info().Fields(fields(keysAndValues)).Msg(msg)
}

func (l *logger) Warn(msg string, keysAndValues ...any) {
	l.instance.Warn().Fields(fields(keysAndValues)).Msg(msg)
}

func (l *logger) Error(msg string, err error, keysAndValues ...any) {
	l.instance.Error().Fields(fields(keysAndValues)).Err(err).Msg(msg)
}

func (l *logger) Debug(msg string, keysAndValues ...any) {
	l.instance.Debug().Fields(fields(keysAndValues)).Msg(msg)
}

func (l *logger) Trace(msg string, keysAndValues ...any) {
	l.instance.Trace().Fields(fields(keysAndValues)).Msg(msg)
}

func (l *logger) With(keysAndValues ...any) Logger {
	return &logger{
		instance: l.instance.With().Fields(fields(keysAndValues)).Logger(),
	}
}

func (l *logger) Instance() zerolog.Logger {
	return l.instance
}

func toZerologLevel(level string) zerolog.Level {
	switch level {
	case "info":
		return zerolog.InfoLevel
	case "warn":
		return zerolog.WarnLevel
	case "error":
		return zerolog.ErrorLevel
	case "debug":
		return zerolog.DebugLevel
	case "trace":
		return zerolog.TraceLevel
	default:
		return zerolog.InfoLevel
	}
}
