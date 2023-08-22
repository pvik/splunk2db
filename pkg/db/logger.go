package db

import (
	"context"
	"errors"
	"fmt"
	"time"

	log "log/slog"

	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
)

// GormLogger struct
type GormLogger struct {
	LogLevel                  logger.LogLevel
	IgnoreRecordNotFoundError bool
	SlowThreshold             time.Duration
	LogTrace                  bool
}

// Print - Log Formatter
func (*GormLogger) Print(v ...interface{}) {
	switch v[0] {
	case "sql":
		log.Debug("gorm sql",
			"msg", v[3],
			"type", "sql",
			"rows_returned", v[5],
			"src", v[1],
			"values", v[4],
			"duration", v[2],
		)
	case "log":
		log.Debug("gorm log", "msg", v[2])
	}
}

// LogMode log mode
func (l *GormLogger) LogMode(level logger.LogLevel) logger.Interface {
	return &GormLogger{}
}

// Info print info
func (l GormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Info {
		log.Debug(fmt.Sprintf(msg, data...))
	}
}

// Warn print warn messages
func (l GormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Warn {
		log.Warn(fmt.Sprintf(msg, data...))
		//log.Warn(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Error print error messages
func (l GormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	if l.LogLevel >= logger.Error {
		log.Error(fmt.Sprintf(msg, data...))
		//log.Error(msg, append([]interface{}{utils.FileWithLineNum()}, data...)...)
	}
}

// Trace print sql message
func (l GormLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if !l.LogTrace {
		return
	}

	if l.LogLevel <= logger.Silent {
		return
	}

	elapsed := time.Since(begin)
	switch {
	case err != nil && l.LogLevel >= logger.Error && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError):
		sql, rows := fc()
		if rows == -1 {
			log.Debug("DB-TRACE-ERR", "src", utils.FileWithLineNum(), "err", err, "elapsed", float64(elapsed.Nanoseconds())/1e6, "rows", "-", "sql", sql)
		} else {
			log.Debug("DB-TRACE-ERR", "src", utils.FileWithLineNum(), "err", err, "elapsed", float64(elapsed.Nanoseconds())/1e6, "rows", rows, "sql", sql)
		}
	case elapsed > l.SlowThreshold && l.SlowThreshold != 0 && l.LogLevel >= logger.Warn:
		sql, rows := fc()
		slowLog := fmt.Sprintf("SLOW SQL >= %v", l.SlowThreshold)
		if rows == -1 {
			log.Debug("DB-TRACE-WARN", "src", utils.FileWithLineNum(), "slow", slowLog, "elapsed", float64(elapsed.Nanoseconds())/1e6, "rows", "-", "sql", sql)
		} else {
			log.Warn("DB-TRACE-WARN", "src", utils.FileWithLineNum(), "slow", slowLog, "elapsed", float64(elapsed.Nanoseconds())/1e6, "rows", rows, "sql", sql)
		}
	case l.LogLevel == logger.Info:
		sql, rows := fc()
		if rows == -1 {
			log.Debug("DB-TRACE", "src", utils.FileWithLineNum(), "elapsed", float64(elapsed.Nanoseconds())/1e6, "rows", "-", "sql", sql)
		} else {
			log.Debug("DB-TRACE", "src", utils.FileWithLineNum(), "elapsed", float64(elapsed.Nanoseconds())/1e6, "rows", rows, "sql", sql)
		}
	}
}
