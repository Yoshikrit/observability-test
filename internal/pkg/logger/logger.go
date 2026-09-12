package logger

import (
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// AccessLogger is a dedicated logger for HTTP access logs, separate from the
// application logger so access log volume/retention can be managed independently.
var AccessLogger zerolog.Logger

// timeFormat is used for every timestamp this app writes (app log, and
// request_time/response_time in the access log). Kept in UTC - see InitAccess
// for why local time shouldn't be baked into stored log timestamps.
const timeFormat = "02-Jan-2006 15:04:05.000"

// Init configures the global zerolog logger used for application logs.
// Deliberately plain/human-readable (not JSON) in every environment: app logs
// are meant for a human scanning `docker logs`/tracing an issue by eye, not
// for structured querying - access logs (JSON, see InitAccess) carry that
// role instead. A collector will still capture these lines into Loki as raw
// text (full-text searchable), just without structured-field filtering.
func Init(level string) {
	zerolog.SetGlobalLevel(parseLevel(level))
	zerolog.TimeFieldFormat = timeFormat

	log.Logger = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: timeFormat}).
		With().
		Timestamp().
		Caller().
		Str("log_type", "app").
		Logger()
}

// InitAccess sets up the access logger. When disabled it becomes a no-op logger.
// It writes structured JSON to stdout (not a file) so a container log collector
// (e.g. Grafana Alloy) can capture it without needing a shared volume/file path.
//
// Every line carries log_type=access - a stable field a log collector's
// pipeline can promote to a Loki label to separate access logs from app logs,
// instead of matching on the "request" message text.
func InitAccess(enabled bool) {
	if !enabled {
		AccessLogger = zerolog.Nop()
		return
	}

	AccessLogger = zerolog.New(os.Stdout).With().Str("log_type", "access").Logger()
}

func parseLevel(level string) zerolog.Level {
	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		return zerolog.InfoLevel
	}
	return lvl
}
