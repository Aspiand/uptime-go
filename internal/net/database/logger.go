package database

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const slowSQLThreshold = 500 * time.Millisecond

type gormLogger struct {
	logLevel      logger.LogLevel
	slowThreshold time.Duration
}

func newGormLogger() logger.Interface {
	return gormLogger{
		logLevel:      logger.Warn,
		slowThreshold: slowSQLThreshold,
	}
}

func (l gormLogger) LogMode(level logger.LogLevel) logger.Interface {
	l.logLevel = level
	return l
}

func (l gormLogger) Info(ctx context.Context, msg string, args ...any) {
	if l.logLevel < logger.Info {
		return
	}
	log.Info().Msgf(msg, args...)
}

func (l gormLogger) Warn(ctx context.Context, msg string, args ...any) {
	if l.logLevel < logger.Warn {
		return
	}
	log.Warn().Msgf(msg, args...)
}

func (l gormLogger) Error(ctx context.Context, msg string, args ...any) {
	if l.logLevel < logger.Error {
		return
	}
	log.Error().Msgf(msg, args...)
}

func (l gormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if l.logLevel == logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	sql, rows := fc()

	switch {
	case err != nil && !errors.Is(err, gorm.ErrRecordNotFound) && l.logLevel >= logger.Error:
		log.Error().
			Err(err).
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("sql error")
	case l.slowThreshold > 0 && elapsed > l.slowThreshold && l.logLevel >= logger.Warn:
		log.Warn().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("slow sql")
	case l.logLevel >= logger.Info:
		log.Info().
			Dur("elapsed", elapsed).
			Int64("rows", rows).
			Str("sql", sql).
			Msg("sql")
	}
}
