package logger

import (
	"io"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

var Logger zerolog.Logger

func Init(level string, pretty bool) zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	zerolog.TimestampFunc = func() time.Time { return time.Now().UTC() }

	var output io.Writer = os.Stdout
	if pretty {
		output = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}
	}

	lvl, err := zerolog.ParseLevel(level)
	if err != nil {
		lvl = zerolog.InfoLevel
	}

	logger := zerolog.New(output).Level(lvl).With().Timestamp().Caller().Logger()
	log.Logger = logger
	Logger = logger
	return logger
}

func WithContext(ctx map[string]interface{}) zerolog.Logger {
	return log.With().Fields(ctx).Logger()
}

func WithRequestID(requestID string) zerolog.Logger {
	return log.With().Str("request_id", requestID).Logger()
}

func WithError(err error) zerolog.Logger {
	return log.With().Err(err).Logger()
}
