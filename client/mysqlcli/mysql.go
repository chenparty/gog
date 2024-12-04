package mysqlcli

import (
	"context"
	"fmt"
	"github.com/chenparty/gog/zlog"
	"github.com/chenparty/gog/zlog/gormplugin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
	"time"
)

var db *gorm.DB

type Options struct {
	TablePrefix   string // 表名前缀
	SingularTable bool   // 使用单数表名

	// Logger
	Silent                    bool          // 是否打印sql语句
	ParameterizedQueries      bool          // 使用参数化查询
	IgnoreRecordNotFoundError bool          // 忽略记录不存在错误
	SlowThreshold             time.Duration // 慢查询阈值
}

type Option func(*Options)

// Connect 连接数据库
func Connect(addr, user, pwd, dbName string, options ...Option) {
	opts := Options{
		SingularTable: true,
		SlowThreshold: time.Second,
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", user, pwd, addr, dbName)
	var err error
	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   opts.TablePrefix,   // 表名前缀，`User` 的表名应该是 `t_users`
			SingularTable: opts.SingularTable, // 使用单数表名，启用该选项，此时，`User` 的表名应该是 `t_user`
		},
		Logger: gormplugin.NewLogger(gormplugin.Config{
			Silent:                    opts.Silent,
			SlowThreshold:             opts.SlowThreshold,
			ParameterizedQueries:      opts.ParameterizedQueries,
			IgnoreRecordNotFoundError: opts.IgnoreRecordNotFoundError,
		}),
	})
	if err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("mysql连接失败")
		panic(err)
	}
	zlog.Info().Str("addr", addr).Msg("mysql连接成功")
}

// WithSilent 设置是否打印sql语句
func WithSilent(silent bool) Option {
	return func(options *Options) {
		options.Silent = silent
	}
}

// WithParameterizedQueries 使用参数化查询
func WithParameterizedQueries(parameterizedQueries bool) Option {
	return func(options *Options) {
		options.ParameterizedQueries = parameterizedQueries
	}
}

// WithIgnoreRecordNotFoundError 忽略记录不存在错误
func WithIgnoreRecordNotFoundError(ignoreRecordNotFoundError bool) Option {
	return func(options *Options) {
		options.IgnoreRecordNotFoundError = ignoreRecordNotFoundError
	}
}

// WithSlowThreshold 设置慢查询阈值
func WithSlowThreshold(slowThreshold time.Duration) Option {
	return func(options *Options) {
		options.SlowThreshold = slowThreshold
	}
}

// DB 获取数据库连接
func DB(ctx context.Context) *gorm.DB {
	return db.WithContext(ctx)
}

type TransactionFunc func(tx *gorm.DB) error

// StartTransaction 开启事务
func StartTransaction(ctx context.Context, trans TransactionFunc) error {
	return db.WithContext(ctx).Transaction(trans)
}
