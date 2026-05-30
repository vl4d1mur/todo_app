package log

import (
	"os"
	"time"
	"todo/internal/config"

	"github.com/rs/zerolog"
)

var Logger zerolog.Logger


func InitLogger() {
	if config.AppEnv == "production" {
		Logger = zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger()
	} else {
		Logger = zerolog.New(zerolog.ConsoleWriter{
			Out: os.Stdout,
			TimeFormat: time.RFC3339,
		}).
			With().
			Timestamp().
			Logger()
	}
}

