package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

type logger struct {
	instance zerolog.Logger
}

type Logger interface {
	Info(msg string, tags map[string]interface{})
	Warn(msg string, tags map[string]interface{})
	Error(msg string, err error, tags map[string]interface{})
	Debug(msg string, tags map[string]interface{})
	Trace(msg string, tags map[string]interface{})
	SetGlobalValue(key string, value any)
}

func NewLogger(level string) Logger {
	var writer io.Writer = os.Stdout
	lvl := toZerologLevel(level)

	zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack

	if lvl == zerolog.DebugLevel || lvl == zerolog.TraceLevel {
		writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
	}

	instance := zerolog.New(writer).Level(lvl).With().Timestamp().Logger()

	return &logger{
		instance: instance,
	}
}

func (l *logger) Info(msg string, tags map[string]interface{}) {
	l.instance.Info().Fields(tags).Msg(msg)
}

func (l *logger) Warn(msg string, tags map[string]interface{}) {
	l.instance.Warn().Fields(tags).Msg(msg)
}

func (l *logger) Error(msg string, err error, tags map[string]interface{}) {
	l.instance.Error().Fields(tags).Err(err).Stack().Msg(msg)
}

func (l *logger) Debug(msg string, tags map[string]interface{}) {
	l.instance.Debug().Fields(tags).Msg(msg)
}

func (l *logger) Trace(msg string, tags map[string]interface{}) {
	l.instance.Trace().Fields(tags).Msg(msg)
}

func (l *logger) SetGlobalValue(key string, value any) {
	l.instance.UpdateContext(func(c zerolog.Context) zerolog.Context {
		return c.Any(key, value)
	})
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
