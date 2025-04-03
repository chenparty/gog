package rediscli

import (
	"context"
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

// Connect 连接redis
func Connect(addrs []string, options ...Option) {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	uniOpt := &redis.UniversalOptions{
		Addrs:    addrs,
		Username: opts.Username, Password: opts.Password, DB: opts.DB,
		MasterName: opts.MasterName, SentinelUsername: opts.SentinelUsername, SentinelPassword: opts.SentinelPassword,
	}
	redisClient = redis.NewUniversalClient(uniOpt)
	//检测是否连接成功
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		zlog.Error().Str("addr", strings.Join(addrs, ",")).Err(err).Msg("redis连接失败")
		panic(err)
	}
	zlog.Info().Str("addr", strings.Join(addrs, ",")).Msg("redis连接成功")
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
