package rediscli

import (
	"context"
	"github.com/chenparty/gog/zlog"
	"github.com/redis/go-redis/v9"
	"strings"
)

var redisClient *redis.Client
var redisClusterClient *redis.ClusterClient

type Options struct {
	Username string
	Password string
	DB       int

	SentinelUsername string
	SentinelPassword string
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

// ConnectForSentinel 连接redis(哨兵模式)
func ConnectForSentinel(masterName string, sentinelAddrs []string, options ...Option) {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	redisClient = redis.NewFailoverClient(&redis.FailoverOptions{
		MasterName:       masterName,
		SentinelAddrs:    sentinelAddrs,
		SentinelUsername: opts.SentinelUsername,
		SentinelPassword: opts.SentinelPassword,
		Username:         opts.Username, Password: opts.Password, DB: opts.DB})
	//检测是否连接成功
	_, err := redisClient.Ping(context.Background()).Result()
	if err != nil {
		zlog.Error().Str("addr", strings.Join(sentinelAddrs, ",")).Err(err).Msg("redis连接失败")
		panic(err)
	}
	zlog.Info().Str("addr", strings.Join(sentinelAddrs, ",")).Msg("redis连接成功")
}

// ConnectForCluster 连接redis(集群模式)
func ConnectForCluster(addrs []string, options ...Option) {
	opts := Options{}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	redisClusterClient = redis.NewClusterClient(&redis.ClusterOptions{Addrs: addrs,
		Username: opts.Username, Password: opts.Password})
	//检测是否连接成功
	_, err := redisClusterClient.Ping(context.Background()).Result()
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

// WithSentinelUserAndPass 设置哨兵用户名和密码
func WithSentinelUserAndPass(user, pwd string) Option {
	return func(options *Options) {
		options.SentinelUsername = user
		options.SentinelPassword = pwd
	}
}
