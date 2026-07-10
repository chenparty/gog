package rediscli

import (
	"context"
	"fmt"
	"github.com/chenparty/gog/zlog"
	"github.com/redis/go-redis/v9"
	"strings"
)

var redisClient redis.UniversalClient

type Options struct {
	Username string
	Password string
	DB       int

	MasterName       string
	SentinelUsername string
	SentinelPassword string
}

type Option func(*Options)

// MustConnect 连接 Redis（Must 版本，适合服务启动阶段，失败直接 panic）
func MustConnect(addrs []string, options ...Option) {
	if err := Connect(addrs, options...); err != nil {
		zlog.Error().Str("addr", strings.Join(addrs, ",")).Err(err).Msg("Redis 连接失败")
		panic(err)
	}
}

// Connect 连接redis
func Connect(addrs []string, options ...Option) error {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	uniOpt := &redis.UniversalOptions{
		Addrs:            addrs,
		Username:         opts.Username,
		Password:         opts.Password,
		DB:               opts.DB,
		MasterName:       opts.MasterName,
		SentinelUsername: opts.SentinelUsername,
		SentinelPassword: opts.SentinelPassword,
		IsClusterMode:    len(addrs) > 1,
	}

	// 先用局部变量接收，不要污染全局变量
	client := redis.NewUniversalClient(uniOpt)

	// 检测是否连接成功
	_, err := client.Ping(context.Background()).Result()
	if err != nil {
		// 失败时直接返回 error，不记录 Error 日志（由调用方决定如何记录）
		return fmt.Errorf("redis连接失败 [%s]: %w", strings.Join(addrs, ","), err)
	}

	// 只有彻底连接成功，才赋值给全局变量
	redisClient = client
	zlog.Info().Str("addr", strings.Join(addrs, ",")).Msg("redis连接成功")
	return nil
}

// WithUserAndPass 设置用户名和密码
func WithUserAndPass(user, pwd string) Option {
	return func(options *Options) {
		options.Username = user
		options.Password = pwd
	}
}

// WithDB 设置数据库
func WithDB(db int) Option {
	return func(options *Options) {
		options.DB = db
	}
}

// WithSentinel 设置哨兵模式
func WithSentinel(masterName, user, pwd string) Option {
	return func(options *Options) {
		options.MasterName = masterName
		options.SentinelUsername = user
		options.SentinelPassword = pwd
	}
}

// Close 关闭 Redis 连接
func Close() {
	if redisClient != nil {
		_ = redisClient.Close()
		zlog.Info().Msg("Redis 连接已关闭")
	}
}
