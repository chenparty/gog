package natscli

import (
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
}

type Option func(*Options)

func Connect(clientName string, servers []string, options ...Option) {
	opts := Options{
		reconnectWait: time.Second * 30,
		maxReconnects: 120,
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	var err error
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
			zlog.Error().Err(err).Msg("NkeyOptionFromSeed")
		} else {
			natsOpts = append(natsOpts, natsOpt)
		}
	} else if opts.Token != "" {
		natsOpts = append(natsOpts, nats.Token(opts.Token))
	}
	// 发起连接
	serversStr := strings.Join(servers, ",")
	nc, err = nats.Connect(serversStr, natsOpts...)
	if err != nil {
		zlog.Error().Err(err).Str("servers", serversStr).Msg("nats连接失败")
		panic(err)
	}
	zlog.Info().Str("servers", serversStr).Msg("nats连接成功")
	// Stream配置
	err = newJetStreamContext()
	if err != nil {
		zlog.Error().Err(err).Msg("createJetStreamContext")
		panic(err)
	}
	zlog.Info().Str("servers", serversStr).Msg("Jetstream Context创建成功")
}
