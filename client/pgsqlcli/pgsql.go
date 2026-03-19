package pgsqlcli

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"gorm.io/gorm/logger"

	"github.com/chenparty/gog/zlog"
	"github.com/chenparty/gog/zlog/gormplugin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	db *gorm.DB
)

const (
	DefaultSlowThreshold = time.Second
	DefaultSingularTable = true
)

type Options struct {
	TablePrefix   string // 表前缀
	SingularTable bool   // 使用单数表名

	// Logger
	Silent                    bool          // 是否打印 sql 语句
	ParameterizedQueries      bool          // 使用参数化查询
	IgnoreRecordNotFoundError bool          // 忽略记录不存在错误
	SlowThreshold             time.Duration // 慢查询阈值

	// Connection
	TimeZone    string // 时区配置，默认 "UTC"
	SSLMode     string // SSL 模式，默认 "disable"
	ConnTimeout int    // 连接超时时间（秒），默认 10
}

type Option func(*Options)

// Connect 连接数据库
func Connect(addr, user, pwd, dbName string, options ...Option) {
	opts := Options{
		SingularTable: DefaultSingularTable,
		SlowThreshold: DefaultSlowThreshold,
		TimeZone:      "Asia/Shanghai", // 默认使用 Asia/Shanghai 时区
		SSLMode:       "disable",       // 默认禁用 SSL
		ConnTimeout:   10,              // 默认 10 秒超时
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	dsn, err := buildDSN(addr, user, pwd, dbName, opts)
	if err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("pgsql 连接失败")
		panic(err)
	}

	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   opts.TablePrefix,
			SingularTable: opts.SingularTable,
		},
		Logger: newGORMLogger(opts),
	})
	if err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("pgsql 连接失败")
		panic(err)
	}
	// 验证数据库连接
	sqlDB, err := db.DB()
	if err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("pgsql 获取底层连接失败")
		panic(err)
	}
	if err = sqlDB.Ping(); err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("pgsql 连接测试失败")
		panic(err)
	}
	zlog.Info().Str("addr", addr).Msg("pgsql 连接成功")
}

func buildDSN(addr, user, pwd, dbName string, opts Options) (string, error) {
	hostPort := strings.Split(addr, ":")
	if len(hostPort) == 0 || hostPort[0] == "" {
		return "", fmt.Errorf("invalid address format: %s", addr)
	}

	host := hostPort[0]
	port := ""
	if len(hostPort) >= 2 {
		port = hostPort[1]
	}

	dsnParts := []string{
		fmt.Sprintf("host=%s", host),
		fmt.Sprintf("dbname=%s", dbName),
		fmt.Sprintf("user=%s", user),
		fmt.Sprintf("password=%s", pwd),
		fmt.Sprintf("sslmode=%s", opts.SSLMode),
		fmt.Sprintf("timezone=%s", opts.TimeZone),
		fmt.Sprintf("connect_timeout=%d", opts.ConnTimeout),
	}
	if port != "" {
		dsnParts = append(dsnParts, fmt.Sprintf("port=%s", port))
	}

	return strings.Join(dsnParts, " "), nil
}

func newGORMLogger(opts Options) logger.Interface {
	return gormplugin.NewLogger(gormplugin.Config{
		Silent:                    opts.Silent,
		SlowThreshold:             opts.SlowThreshold,
		ParameterizedQueries:      opts.ParameterizedQueries,
		IgnoreRecordNotFoundError: opts.IgnoreRecordNotFoundError,
	})
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

// WithTimeZone 设置时区（例如："UTC", "Asia/Shanghai", "America/New_York"）
func WithTimeZone(timeZone string) Option {
	return func(options *Options) {
		if timeZone != "" {
			options.TimeZone = timeZone
		}
	}
}

// WithSSLMode 设置 SSL 模式（"disable", "require", "verify-ca", "verify-full"）
func WithSSLMode(sslMode string) Option {
	return func(options *Options) {
		if sslMode != "" {
			options.SSLMode = sslMode
		}
	}
}

// WithConnTimeout 设置连接超时时间（秒）
func WithConnTimeout(timeout int) Option {
	return func(options *Options) {
		if timeout > 0 {
			options.ConnTimeout = timeout
		}
	}
}

// DB 获取数据库连接
func DB(ctx context.Context) *gorm.DB {
	if db == nil {
		panic("database not initialized")
	}
	return db.WithContext(ctx)
}

type TransactionFunc func(tx *gorm.DB) error

// StartTransaction 开启事务
func StartTransaction(ctx context.Context, trans TransactionFunc) error {
	if db == nil {
		return errors.New("database not initialized")
	}
	return db.WithContext(ctx).Transaction(trans)
}

// IsRecordNotFoundErr 判断是否记录不存在错误
func IsRecordNotFoundErr(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// Close 关闭数据库连接
func Close() {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
			zlog.Info().Msg("PostgreSQL 连接已关闭")
		}
	}
}
