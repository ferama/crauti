package logger

import (
	"os"

	"github.com/ferama/crauti/pkg/conf"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func GetLogger(component string) *zerolog.Logger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if conf.ConfInst.Debug {
		// using console writer is inefficient.
		// use it for debug only
		log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr, TimeFormat: "2006-01-02T15:04:05"})
	}

	logger := log.With().
		Str("src", component).
		Logger()

	return &logger
}
