package log

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger

func InitLogger() {
	env := os.Getenv("APP_ENV")

	if env == "" {
		env = "development"
	}

	if env == "production" {
		Logger = zerolog.New(os.Stdout).
			With().
			Timestamp().
			Logger()
	} else {
		Logger = zerolog.New(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}).
			With().
			Timestamp().
			Logger()
	}
}
