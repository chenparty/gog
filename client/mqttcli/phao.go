package mqttcli

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/chenparty/gog/zlog"
	"github.com/eclipse/paho.golang/autopaho"
	"github.com/eclipse/paho.golang/paho"
	"github.com/oklog/ulid/v2"
)

type MsgHandler func(ID uint16, topic string, payload []byte)

var (
	connectionMgr *autopaho.ConnectionManager
	mu            sync.RWMutex

	router    = paho.NewStandardRouter()
	subTopics = make(map[string]byte)
	subMu     sync.RWMutex

	connected int32
	ctx       context.Context
	cancel    context.CancelFunc
)

type Options struct {
	ClientID    string
	Username    string
	Password    string
	tls         *tls.Config
	willTopic   string
	willPayload []byte
	willQos     byte
	willRetain  bool
}

type Option func(*Options)

// MustConnect 建立 MQTT 连接（Must 版本，适合服务启动阶段，失败直接 panic）
func MustConnect(addr string, options ...Option) {
	if err := Connect(addr, options...); err != nil {
		zlog.Error().Str("addr", addr).Err(err).Msg("MQTT 连接失败")
		panic(err)
	}
}

func Connect(addr string, options ...Option) error {
	mu.Lock()
	defer mu.Unlock()

	if connectionMgr != nil {
		return fmt.Errorf("MQTT 客户端已连接，请先调用 Close")
	}

	opts := Options{ClientID: ulid.Make().String()}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}

	// 兼容处理：如果没有协议头，默认加上 tcp://
	if !strings.Contains(addr, "://") {
		addr = "tcp://" + addr
	}
	u, err := url.Parse(addr)
	if err != nil {
		return fmt.Errorf("invalid broker address: %w", err)
	}

	ctx, cancel = context.WithCancel(context.Background())

	cliCfg := autopaho.ClientConfig{
		ServerUrls:                    []*url.URL{u},
		KeepAlive:                     30,
		CleanStartOnInitialConnection: true,
		SessionExpiryInterval:         60,

		OnConnectionUp: func(cm *autopaho.ConnectionManager, connAck *paho.Connack) {
			atomic.StoreInt32(&connected, 1)
			zlog.Info().Str("addr", addr).Msg("MQTT 连接成功/重连成功")
			// 直接传递 ctx，底层派生子 context 是并发安全的
			go resubscribeAll(cm, ctx)
		},

		OnConnectError: func(err error) {
			zlog.Error().Str("addr", addr).Err(err).Msg("MQTT 连接尝试失败")
		},

		ClientConfig: paho.ClientConfig{
			ClientID: opts.ClientID,
			Router:   router,
			OnPublishReceived: []func(paho.PublishReceived) (bool, error){
				func(pr paho.PublishReceived) (bool, error) {
					if zlog.Debug().Enabled() {
						zlog.Debug().Str("topic", pr.Packet.Topic).Msg("收到MQTT消息")
					}
					return false, nil
				},
			},
			OnClientError: func(err error) {
				atomic.StoreInt32(&connected, 0)
				zlog.Error().Err(err).Msg("MQTT 客户端错误")
			},
			OnServerDisconnect: func(d *paho.Disconnect) {
				atomic.StoreInt32(&connected, 0)
				if d.Properties != nil {
					zlog.Warn().Str("reason", d.Properties.ReasonString).Msg("MQTT 服务器请求断开连接")
				} else {
					zlog.Warn().Int("reason_code", int(d.ReasonCode)).Msg("MQTT 服务器请求断开连接")
				}
			},
		},
	}

	cliCfg.ConnectPacketBuilder = func(c *paho.Connect, u *url.URL) (*paho.Connect, error) {
		if opts.Username != "" {
			c.Username = opts.Username
			c.Password = []byte(opts.Password)
			c.UsernameFlag = true
			c.PasswordFlag = true
		}
		if opts.willTopic != "" {
			c.WillMessage = &paho.WillMessage{
				Topic:   opts.willTopic,
				Payload: opts.willPayload,
				QoS:     opts.willQos,
				Retain:  opts.willRetain,
			}
		}
		return c, nil
	}

	if opts.tls != nil {
		cliCfg.TlsCfg = opts.tls
	}

	cm, err := autopaho.NewConnection(ctx, cliCfg)
	if err != nil {
		cancel()
		return fmt.Errorf("failed to create connection: %w", err)
	}
	if err := cm.AwaitConnection(ctx); err != nil {
		cancel()
		return fmt.Errorf("failed to establish initial connection: %w", err)
	}
	connectionMgr = cm
	atomic.StoreInt32(&connected, 1)
	zlog.Info().Str("addr", addr).Msg("MQTT 初始连接已建立")
	return nil
}

func resubscribeAll(cm *autopaho.ConnectionManager, parentCtx context.Context) {
	subMu.RLock()
	subs := make([]paho.SubscribeOptions, 0, len(subTopics))
	for topic, qos := range subTopics {
		subs = append(subs, paho.SubscribeOptions{Topic: topic, QoS: qos})
	}
	subMu.RUnlock()

	if len(subs) == 0 {
		return
	}

	const batchSize = 50
	for i := 0; i < len(subs); i += batchSize {
		if parentCtx.Err() != nil {
			return
		}
		end := i + batchSize
		if end > len(subs) {
			end = len(subs)
		}
		batch := subs[i:end]

		subCtx, cancel := context.WithTimeout(parentCtx, 10*time.Second)
		_, err := cm.Subscribe(subCtx, &paho.Subscribe{Subscriptions: batch})
		cancel()
		if err != nil {
			zlog.Error().Err(err).Int("batch_start", i).Msg("重连后批量重新订阅失败")
		}
	}
	zlog.Debug().Msg("重连后批量重新订阅完成")
}

func Close() {
	mu.Lock()
	if connectionMgr == nil {
		mu.Unlock()
		return
	}
	cm := connectionMgr
	connectionMgr = nil

	if cancel != nil {
		cancel()
	}
	mu.Unlock()

	<-cm.Done()
	atomic.StoreInt32(&connected, 0)
	zlog.Info().Msg("MQTT连接已关闭")
}

func WithClientID(clientID string, asPrefix bool) Option {
	return func(options *Options) {
		if len(clientID) > 0 {
			if asPrefix {
				options.ClientID = clientID + "-" + options.ClientID
			} else {
				options.ClientID = clientID
			}
		}
	}
}

func AuthWithUser(username, pwd string) Option {
	return func(options *Options) {
		options.Username = username
		options.Password = pwd
	}
}

func AuthWithTLS(certFile, keyFile string) Option {
	return func(options *Options) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			zlog.Error().Err(err).Str("cert", certFile).Msg("加载 TLS 证书失败")
			return
		}
		options.tls = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
}

func WithWillMessage(topic string, payload []byte, qos byte, retain bool) Option {
	return func(options *Options) {
		options.willTopic = topic
		options.willPayload = payload
		options.willQos = qos
		options.willRetain = retain
	}
}

// Subscribe 订阅主题
func Subscribe(topic string, qos byte, callback MsgHandler) error {
	mu.RLock()
	cm := connectionMgr
	mu.RUnlock()
	if cm == nil {
		return fmt.Errorf("MQTT 客户端未初始化")
	}

	router.UnregisterHandler(topic)
	router.RegisterHandler(topic, func(p *paho.Publish) {
		defer func() {
			if r := recover(); r != nil {
				zlog.Error().Interface("panic", r).Str("topic", p.Topic).Msg("MQTT 消息处理发生 panic")
			}
		}()
		callback(p.PacketID, p.Topic, p.Payload)
	})

	subMu.Lock()
	subTopics[topic] = qos
	subMu.Unlock()

	if !IsConnected() {
		zlog.Warn().Str("topic", topic).Msg("MQTT 当前未连接，订阅将在重连后自动执行")
		return nil
	}

	// 直接使用全局 ctx，安全且简洁
	subCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := cm.Subscribe(subCtx, &paho.Subscribe{
		Subscriptions: []paho.SubscribeOptions{{Topic: topic, QoS: qos}},
	})
	if err != nil {
		zlog.Error().Str("topic", topic).Err(err).Msg("MQTT 订阅失败")
		return err
	}
	zlog.Debug().Str("topic", topic).Msg("MQTT 订阅成功")
	return nil
}

// Unsubscribe 取消订阅
func Unsubscribe(topic string) error {
	mu.RLock()
	cm := connectionMgr
	mu.RUnlock()
	if cm == nil {
		return fmt.Errorf("MQTT 客户端未初始化")
	}
	if !IsConnected() {
		return fmt.Errorf("MQTT 未连接，请先建立连接后再取消订阅")
	}

	// 直接使用全局 ctx
	unsubCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if _, err := cm.Unsubscribe(unsubCtx, &paho.Unsubscribe{Topics: []string{topic}}); err != nil {
		zlog.Error().Str("topic", topic).Err(err).Msg("MQTT 取消订阅失败")
		return err
	}

	router.UnregisterHandler(topic)
	subMu.Lock()
	delete(subTopics, topic)
	subMu.Unlock()
	zlog.Debug().Str("topic", topic).Msg("MQTT 取消订阅成功")
	return nil
}

// Publish 发布消息
func Publish(topic string, qos byte, payload any) error {
	mu.RLock()
	cm := connectionMgr
	mu.RUnlock()
	if cm == nil {
		return fmt.Errorf("MQTT 客户端未初始化")
	}

	var pld []byte
	switch v := payload.(type) {
	case []byte:
		pld = v
	case string:
		pld = []byte(v)
	default:
		var err error
		pld, err = json.Marshal(payload)
		if err != nil {
			return fmt.Errorf("序列化 payload 失败: %w", err)
		}
	}

	// 直接使用全局 ctx
	pubCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	_, err := cm.Publish(pubCtx, &paho.Publish{Topic: topic, QoS: qos, Payload: pld})
	if err != nil {
		zlog.Error().Str("topic", topic).Err(err).Msg("MQTT 发布失败")
		return err
	}
	zlog.Debug().Str("topic", topic).Msg("MQTT 发布成功")
	return nil
}

// IsConnected 检测当前是否已连接
func IsConnected() bool {
	return atomic.LoadInt32(&connected) == 1
}

// GetConnectionStatus 获取当前连接状态
func GetConnectionStatus() string {
	if IsConnected() {
		return "已连接"
	}
	mu.RLock()
	cm := connectionMgr
	mu.RUnlock()
	if cm != nil {
		return "连接断开中(自动重连)"
	}
	return "未初始化"
}

// WaitForConnection 阻塞等待连接恢复（通常用于断线重连场景）。
// 与 IsConnected() 的非阻塞瞬时判断不同，当检测到断开时，此方法会挂起当前协程，
// 等待底层 autopaho 自动重连成功，或直到超时。
// 适用场景：发送关键消息前发现断线，不想写 for+sleep 盲等，可用此方法优雅阻塞等待。
func WaitForConnection(timeout time.Duration) error {
	mu.RLock()
	cm := connectionMgr
	mu.RUnlock()
	if cm == nil {
		return fmt.Errorf("MQTT 客户端未初始化")
	}

	// 直接使用全局 ctx
	waitCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return cm.AwaitConnection(waitCtx)
}

// GetConnectionManager 获取底层连接管理器实例（逃生舱方法）。
// 仅当本包封装的基础能力无法满足特殊需求，需要直接调用 autopaho 原生高级 API 时才使用。
// 普通业务场景请勿调用，以免破坏内部状态管理。
func GetConnectionManager() *autopaho.ConnectionManager {
	mu.RLock()
	defer mu.RUnlock()
	return connectionMgr
}
