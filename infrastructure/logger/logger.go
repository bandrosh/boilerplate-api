package logger

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/rs/zerolog"
)

var (
	log zerolog.Logger

	// ServerError ...
	ServerError = "%v type=server.error"

	// ServerInfo ...
	ServerInfo = "%v type=server.info"

	// FatalError ...
	FatalError = "%v type=fatal.error"

	// ConfigError ...
	ConfigError = "%v type=config.error"

	// HTTPError ...
	HTTPError = "%v type=http.error"

	// HTTPWarn ...
	HTTPWarn = "%v type=http.warn"

	// HTTPInfo ...
	HTTPInfo = "%v type=http.info"
)

func init() {
	buildInfo, _ := debug.ReadBuildInfo()

	fmt.Println("Initialize Log")

	log = zerolog.New(zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.RFC3339}).
		Level(zerolog.TraceLevel).
		With().
		Timestamp().
		Int("pid", os.Getpid()).
		Str("go_version", buildInfo.GoVersion).
		Logger()
}

func Info(format string, v ...interface{}) {
	log.Info().Msgf(format, v...)
}

func Warn(format string, v ...interface{}) {
	log.Warn().Msgf(format, v...)
}

func Error(format string, v ...interface{}) {
	log.Error().Msgf(format, v...)
}

func Fatal(format string, v ...interface{}) {
	log.Fatal().Msgf(format, v...)
}
