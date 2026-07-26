package logger

import (
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init configures the global zerolog logger based on the environment.
// In development: human-readable console output with colors.
// In production: structured JSON output for log aggregation.
func Init(environment string) {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	var output zerolog.ConsoleWriter
	isDev := environment == "development" || environment == ""

	if isDev {
		output = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.Kitchen,
			NoColor:    false,
		}
	}

	var logger zerolog.Logger
	if isDev {
		logger = zerolog.New(output).
			Level(zerolog.DebugLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	} else {
		logger = zerolog.New(os.Stdout).
			Level(zerolog.InfoLevel).
			With().
			Timestamp().
			Caller().
			Logger()
	}

	log.Logger = logger
}
