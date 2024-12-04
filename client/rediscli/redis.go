package rediscli

import (
	"context"
	"github.com/chenparty/gog/zlog"
	"github.com/redis/go-redis/v9"
)

var redisClient *redis.Client

type Options struct {
	Username string
	Password string
	DB       int
}

type Option func(*Options)

// Connect 连接redis
func Connect(addr string, options ...Option) {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	redisClient = redis.NewClient(&redis.Options{Addr: addr,
		Username: opts.Username, Password: opts.Password, DB: opts.DB})
	//检测是否连接成功
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("redis连接失败")
		panic(err)
	}
	zlog.Info().Str("addr", addr).Msg("redis连接成功")
}

// WithUsername 设置用户名
func WithUsername(user string) Option {
	return func(options *Options) {
		options.Username = user
	}
}

// WithPassword 设置密码
func WithPassword(pwd string) Option {
	return func(options *Options) {
		options.Password = pwd
	}
}

// WithDB 设置数据库
func WithDB(db int) Option {
	return func(options *Options) {
		options.DB = db
	}
}
