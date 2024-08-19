package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

type logger struct {
	level zerolog.Level
}

type Logger interface {
	Info(msg string, tags map[string]interface{})
	Warn(msg string, tags map[string]interface{})
	Error(msg string, err error, tags map[string]interface{})
	Debug(msg string, tags map[string]interface{})
	Trace(msg string, tags map[string]interface{})
}

func NewLogger(level string) Logger {
	lvl := setup(level)

	return &logger{
		level: lvl,
	}
}

func (l *logger) Info(msg string, tags map[string]interface{}) {
	l.instance().Info().Fields(tags).Msg(msg)
}

func (l *logger) Warn(msg string, tags map[string]interface{}) {
	l.instance().Warn().Fields(tags).Msg(msg)
}

func (l *logger) Error(msg string, err error, tags map[string]interface{}) {
	ev := l.instance().Error().Fields(tags).Err(err)

	if l.level == zerolog.DebugLevel {
		ev = ev.Stack()
	}

	ev.Msg(msg)
}

func (l *logger) Debug(msg string, tags map[string]interface{}) {
	l.instance().Debug().Fields(tags).Msg(msg)
}

func (l *logger) Trace(msg string, tags map[string]interface{}) {
	l.instance().Trace().Fields(tags).Msg(msg)
}

func setup(level string) zerolog.Level {
	lvl := toZerologLevel(level)
	zerolog.SetGlobalLevel(lvl)

	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	})

	return lvl
}

func (l *logger) instance() *zerolog.Logger {
	return &log.Logger
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
