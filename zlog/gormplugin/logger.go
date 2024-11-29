package gormplugin

import (
	"context"
	"errors"
	"gog/zlog"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/utils"
	"time"
)

// New initialize logger
func New(config Config) logger.Interface {
	return &zlogGormLogger{
		Config: config,
	}
}

// Config logger config
type Config struct {
	Silent                    bool
	SlowThreshold             time.Duration
	ParameterizedQueries      bool
	IgnoreRecordNotFoundError bool
}

type zlogGormLogger struct {
	Config
}

func (l *zlogGormLogger) LogMode(_ logger.LogLevel) logger.Interface {
	// 返回一个新的 zlogGormLogger，设置日志级别
	return l
}

func (l *zlogGormLogger) Info(ctx context.Context, msg string, data ...interface{}) {
	zlog.Info().Ctx(ctx).Str("source", utils.FileWithLineNum()).Msgf(msg, data)
}

func (l *zlogGormLogger) Warn(ctx context.Context, msg string, data ...interface{}) {
	zlog.Warn().Ctx(ctx).Str("source", utils.FileWithLineNum()).Msgf(msg, data)
}

func (l *zlogGormLogger) Error(ctx context.Context, msg string, data ...interface{}) {
	zlog.Error().Ctx(ctx).Str("source", utils.FileWithLineNum()).Msgf(msg, data)
}

func (l *zlogGormLogger) Trace(ctx context.Context, start time.Time, fc func() (string, int64), err error) {
	if l.Silent {
		return
	}
	// 获取 SQL 查询的详细信息
	sql, rowsAffected := fc()
	duration := time.Since(start)

	// 记录 SQL 日志信息，包括执行时间、影响的行数等
	event := zlog.Info().Ctx(ctx).
		Dur("duration", duration).
		Int64("rows", rowsAffected).
		Str("sql", sql).
		Str("source", utils.FileWithLineNum())
	if err != nil && (!errors.Is(err, logger.ErrRecordNotFound) || !l.IgnoreRecordNotFoundError) {
		event.Err(err).Msg("SQL exec failed")
	} else if duration > l.SlowThreshold && l.SlowThreshold > 0 {
		event.Dur("slow_threshold", l.SlowThreshold).Msg("Slow SQL")
	} else {
		event.Msg("SQL query executed")
	}
}

func (l *zlogGormLogger) ParamsFilter(_ context.Context, sql string, params ...interface{}) (string, []interface{}) {
	if l.Config.ParameterizedQueries {
		return sql, nil
	}
	return sql, params
}
