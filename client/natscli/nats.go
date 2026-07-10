package natscli

import (
	"fmt"
	"github.com/chenparty/gog/zlog"
	"github.com/nats-io/nats.go"
	"strings"
	"time"
)

var nc *nats.Conn

type Options struct {
	// 连接基础配置项
	reconnectWait time.Duration // 每次重连等待时间
	maxReconnects int           // 最大重连次数

	// 认证配置-用户名密码方式
	Username string // 用户名
	Password string // 密码
	// 认证配置-NKey方式
	NKeySeedFile string
	// 认证配置-TOKEN方式
	Token string

	//启用JetStream
	EnableJetStream bool
}

type Option func(*Options)

// MustConnect 连接 NATS（Must 版本，适合服务启动阶段，失败直接 panic）
func MustConnect(clientName string, servers []string, options ...Option) {
	serversStr := strings.Join(servers, ",")
	if err := Connect(clientName, servers, options...); err != nil {
		zlog.Error().Str("servers", serversStr).Err(err).Msg("NATS 连接失败")
		panic(err)
	}
}

// Connect NATS连接
func Connect(clientName string, servers []string, options ...Option) error {
	opts := Options{
		reconnectWait: time.Second * 30,
		maxReconnects: 120,
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	// 基础配置项
	natsOpts := []nats.Option{nats.Name(clientName)}
	natsOpts = append(natsOpts, nats.ReconnectWait(opts.reconnectWait))
	natsOpts = append(natsOpts, nats.MaxReconnects(opts.maxReconnects))
	natsOpts = append(natsOpts, nats.DisconnectErrHandler(func(nc *nats.Conn, err error) {
		zlog.Error().Err(err).Msg("nats.DisconnectErrHandler")
	}))
	natsOpts = append(natsOpts, nats.ReconnectHandler(func(nc *nats.Conn) {
		zlog.Info().Str("url", nc.ConnectedUrl()).Msg("NATS reconnected")
	}))
	natsOpts = append(natsOpts, nats.ClosedHandler(func(nc *nats.Conn) {
		zlog.Info().Str("url", nc.ConnectedUrl()).Msg("NATS closed")
	}))

	// 加密配置
	if opts.Username != "" && opts.Password != "" {
		natsOpts = append(natsOpts, nats.UserInfo(opts.Username, opts.Password))
	} else if opts.NKeySeedFile != "" {
		natsOpt, e := nats.NkeyOptionFromSeed(opts.NKeySeedFile)
		if e != nil {
			return fmt.Errorf("nats NKeyOptionFromSeed 失败 [%s]: %w", opts.NKeySeedFile, e)
		}
		natsOpts = append(natsOpts, natsOpt)
	} else if opts.Token != "" {
		natsOpts = append(natsOpts, nats.Token(opts.Token))
	}

	// 发起连接
	serversStr := strings.Join(servers, ",")
	localNc, err := nats.Connect(serversStr, natsOpts...)
	if err != nil {
		return fmt.Errorf("nats 连接失败 [%s]: %w", serversStr, err)
	}

	// Stream 配置
	if opts.EnableJetStream {
		// 因为 newJetStreamContext() 内部肯定依赖读取全局 nc 变量，所以这里必须先赋值
		nc = localNc
		err = newJetStreamContext()
		if err != nil {
			// 如果 JetStream 初始化失败，必须关闭底层连接并清理全局变量，防止连接泄漏
			nc.Close()
			nc = nil
			return fmt.Errorf("nats 创建 JetStream Context 失败: %w", err)
		}
		zlog.Info().Str("servers", serversStr).Msg("JetStream Context创建成功")
	} else {
		nc = localNc
	}

	zlog.Info().Str("servers", serversStr).Msg("nats连接成功")
	return nil
}

// NewZlogLoggerWithNATS 使用NATS作为日志输出
func NewZlogLoggerWithNATS(level string, subj string) {
	if nc == nil {
		panic("NATS还未创建连接")
	}
	zlog.NewLogLogger("NATS", level, zlog.NATSAttr(nc, subj))
}

// WithUserAndPass 用户名密码认证
func WithUserAndPass(username, password string) Option {
	return func(opts *Options) {
		opts.Username = username
		opts.Password = password
	}
}

// WithNKey NKey认证
func WithNKey(seedFile string) Option {
	return func(options *Options) {
		options.NKeySeedFile = seedFile
	}
}

// WithToken TOKEN认证
func WithToken(token string) Option {
	return func(options *Options) {
		options.Token = token
	}
}

// WithJetStream 是否启用 JetStream
func WithJetStream(enable bool) Option {
	return func(options *Options) {
		options.EnableJetStream = enable
	}
}

// Close 关闭 NATS 连接和 JetStream 上下文
func Close() {
	if nc != nil {
		nc.Close()
		zlog.Info().Msg("NATS 连接已关闭")
	}
}
