package logger

import (
	"context"
	"os"

	"github.com/rs/zerolog"
)

var AppLogger zerolog.Logger
var AccessLogger zerolog.Logger

const timeFormat = "02-Jan-2006 15:04:05.000"

func Init(level string) {
	zerolog.SetGlobalLevel(parseLevel(level))
	zerolog.TimeFieldFormat = timeFormat

	AppLogger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat}).
		With().
		Timestamp().
		Caller().
		Str("log_type", "app").
		Logger()
}

func InitAccess(enabled bool) {
	if !enabled {
		AccessLogger = zerolog.Nop()
		return
	}

	AccessLogger = zerolog.New(os.Stdout).With().Str("log_type", "access").Logger()
}

type ctxKey struct{}

// setter
func WithRequestLogger(ctx context.Context, l zerolog.Logger) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// getter
func FromContext(ctx context.Context) zerolog.Logger {
	if l, ok := ctx.Value(ctxKey{}).(zerolog.Logger); ok {
		return l
	}
	return AppLogger
}

func parseLevel(level string) zerolog.Level {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return lvl
}
