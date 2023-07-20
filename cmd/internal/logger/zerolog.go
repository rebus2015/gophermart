package logger

import (
	"context"
	"os"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)
type config interface{
	GetLogLevel() (string)
	GetLogMode() (string)
}

func InitZerolog(ctx context.Context, cfg config) context.Context {
	zerolog.SetGlobalLevel(convertToZerologLevel(cfg.GetLogLevel()))

	switch cfg.GetLogMode() {
	case ModeProduction:
		log.Logger = log.Output(os.Stdout)
	case ModeDevelopment:
		fallthrough
	default:
		log.Logger = log.Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		})
	}
	return log.Logger.WithContext(ctx)
}

func (m *message) Fatalf(format string, v ...interface{}) {
	getZerologLogger(m.ctx).Fatal().Fields(m.fields).Msgf(format, v...)
}

func (m *message) Errorf(format string, v ...interface{}) {
	getZerologLogger(m.ctx).Error().Fields(m.fields).Msgf(format, v...)
}

func (m *message) Infof(format string, v ...interface{}) {
	getZerologLogger(m.ctx).Info().Fields(m.fields).Msgf(format, v...)
}

func (m *message) Debugf(format string, v ...interface{}) {
	getZerologLogger(m.ctx).Debug().Fields(m.fields).Msgf(format, v...)
}

func getZerologLogger(ctx context.Context) *zerolog.Logger {
	logger := log.Ctx(ctx)
	if logger.GetLevel() == zerolog.Disabled {
		return &log.Logger
	}
	return logger
}

func convertToZerologLevel(level string) zerolog.Level {
	switch level {
	case LevelCritical:
		return zerolog.FatalLevel
	case LevelError:
		return zerolog.ErrorLevel
	case LevelDebug:
		return zerolog.DebugLevel
	case LevelInfo:
		fallthrough
	default:
		return zerolog.InfoLevel
	}
}