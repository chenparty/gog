package pgsqlcli

import (
	"context"
	"errors"
	"fmt"
	"gorm.io/gorm/logger"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/chenparty/gog/zlog"
	"github.com/chenparty/gog/zlog/gormplugin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
)

var (
	db   *gorm.DB
	once sync.Once
)

const (
	DefaultSlowThreshold = time.Second
	DefaultSingularTable = true
)

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
	once.Do(func() {
		opts := Options{
			SingularTable: DefaultSingularTable,
			SlowThreshold: DefaultSlowThreshold,
		}
		for _, opt := range options {
			if opt != nil {
				opt(&opts)
			}
		}

		dsn, err := buildDSN(addr, user, pwd, dbName)
		if err != nil {
			zlog.Error().Str("addr", addr).Err(err).Msg("pgsql连接失败")
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
	})
}

func buildDSN(addr, user, pwd, dbName string) (string, error) {
	hostPort := strings.Split(addr, ":")
	if len(hostPort) == 0 || hostPort[0] == "" {
		return "", fmt.Errorf("invalid address format: %s", addr)
	}

	// 使用 url.Values 安全编码参数，避免特殊字符问题
	params := url.Values{}
	params.Set("dbname", dbName)
	params.Set("user", user)
	params.Set("password", pwd)
	params.Set("sslmode", "disable") // 默认禁用 SSL，如需启用可通过选项配置

	dsn := fmt.Sprintf("host=%s %s", hostPort[0], params.Encode())
	if len(hostPort) >= 2 {
		dsn = fmt.Sprintf("%s port=%s", dsn, hostPort[1])
	}
	return dsn, nil
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
