package mysqlcli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/chenparty/gog/zlog"
	"github.com/chenparty/gog/zlog/gormplugin"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/schema"
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

// MustConnect 连接 MySQL（Must 版本，适合服务启动阶段，失败直接 panic）
func MustConnect(addr, user, pwd, dbName string, options ...Option) {
	if err := Connect(addr, user, pwd, dbName, options...); err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("MySQL 连接失败")
		panic(err)
	}
}

// Connect 连接数据库
func Connect(addr, user, pwd, dbName string, options ...Option) error {
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

	// 使用局部变量接收，避免初始化失败时污染全局变量
	gormDB, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
		NamingStrategy: schema.NamingStrategy{
			TablePrefix:   opts.TablePrefix,
			SingularTable: opts.SingularTable,
		},
		Logger: gormplugin.NewLogger(gormplugin.Config{
			Silent:                    opts.Silent,
			SlowThreshold:             opts.SlowThreshold,
			ParameterizedQueries:      opts.ParameterizedQueries,
			IgnoreRecordNotFoundError: opts.IgnoreRecordNotFoundError,
		}),
	})
	if err != nil {
		return fmt.Errorf("mysql gorm.Open 失败 [%s]: %w", addr, err)
	}

	// 验证数据库连接
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("mysql 获取底层连接失败 [%s]: %w", addr, err)
	}
	if err = sqlDB.Ping(); err != nil {
		return fmt.Errorf("mysql 连接测试失败 [%s]: %w", addr, err)
	}

	// 全部成功，赋值给全局变量
	db = gormDB
	zlog.Info().Str("addr", addr).Msg("mysql 连接成功")
	return nil
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

// IsRecordNotFoundErr 判断是否记录不存在错误
func IsRecordNotFoundErr(err error) bool {
	return errors.Is(err, gorm.ErrRecordNotFound)
}

// TxOperation 事务操作
type TxOperation func(tx *gorm.DB) error

// StartTransaction 开启事务
func StartTransaction(ctx context.Context, trans TxOperation) error {
	return db.WithContext(ctx).Transaction(trans)
}

// ExecuteInTx 批量执行事务
func ExecuteInTx(ctx context.Context, ops ...TxOperation) error {
	return StartTransaction(ctx, func(tx *gorm.DB) error {
		for _, op := range ops {
			if err := op(tx); err != nil {
				return err
			}
		}
		return nil
	})
}

// Close 关闭数据库连接
func Close() {
	if db != nil {
		sqlDB, err := db.DB()
		if err == nil {
			_ = sqlDB.Close()
			zlog.Info().Msg("MySQL 连接已关闭")
		}
	}
}
