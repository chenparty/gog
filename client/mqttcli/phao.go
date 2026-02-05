package mqttcli

import (
	"crypto/tls"
	"fmt"
	"sync"
	"time"

	"github.com/chenparty/gog/zlog"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/oklog/ulid/v2"
)

type MsgHandler func(ID uint16, topic string, payload []byte)

var (
	subscribes  = map[string]MsgHandler{}
	subTopicQos = map[string]byte{}
	mu          sync.RWMutex // 保护subscribes和subTopicQos
	mqttClient  MQTT.Client
)

type Options struct {
	ClientID string // 客户端ID,不设置时会自动随机生成
	Username string // 用户名
	Password string // 密码

	tls *tls.Config
}

type Option func(*Options)

// Connect 连接MQTT服务器
func Connect(addr string, options ...Option) {
	opts := Options{
		ClientID: ulid.Make().String(),
	}
	for _, opt := range options {
		if opt != nil {
			opt(&opts)
		}
	}
	clientOptions := MQTT.NewClientOptions()
	clientOptions.AddBroker(addr)
	clientOptions.SetClientID(opts.ClientID)

	if opts.Username != "" {
		clientOptions.SetUsername(opts.Username)
		clientOptions.SetPassword(opts.Password)
	}

	if opts.tls != nil {
		clientOptions.SetTLSConfig(opts.tls)
	}

	clientOptions.SetAutoReconnect(true)
	clientOptions.SetConnectTimeout(30 * time.Second)
	clientOptions.SetWriteTimeout(10 * time.Second)
	clientOptions.SetKeepAlive(30 * time.Second)            // 设置KeepAlive为60秒，根据实际情况调整
	clientOptions.SetPingTimeout(10 * time.Second)          // 设置PingTimeout为10秒，根据实际情况调整
	clientOptions.SetMaxReconnectInterval(30 * time.Second) // 最大重连间隔

	// 添加连接状态监控
	clientOptions.OnConnect = onConnectHandler(addr)
	clientOptions.OnConnectionLost = onConnectionLostHandler(addr)
	clientOptions.OnReconnecting = onReconnectingHandler(addr)

	mqttClient = MQTT.NewClient(clientOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		zlog.Error().Str("addr", addr).Err(token.Error()).Msg("MQTT连接失败")
		panic(token.Error())
	}
}

// 连接成功回调
func onConnectHandler(addr string) func(MQTT.Client) {
	return func(client MQTT.Client) {
		zlog.Info().Str("addr", addr).Msg("MQTT连接成功")

		// 使用互斥锁保护全局变量访问
		mu.RLock()
		defer mu.RUnlock()

		// 连接后重新订阅所有主题
		for topic, handler := range subscribes {
			qos := subTopicQos[topic]
			topicCopy := topic
			handlerCopy := handler
			token := client.Subscribe(topicCopy, qos, func(client MQTT.Client, message MQTT.Message) {
				handlerCopy(message.MessageID(), message.Topic(), message.Payload())
			})

			if token.Wait() && token.Error() != nil {
				zlog.Error().Str("topic", topic).Err(token.Error()).Msg("MQTT重新订阅失败")
			} else {
				zlog.Debug().Str("topic", topic).Msg("MQTT重新订阅成功")
			}
		}
	}
}

// 连接断开回调
func onConnectionLostHandler(addr string) func(MQTT.Client, error) {
	return func(client MQTT.Client, err error) {
		zlog.Error().Str("addr", addr).Err(err).Msg("MQTT连接断开")
	}
}

// 重连中回调
func onReconnectingHandler(addr string) func(MQTT.Client, *MQTT.ClientOptions) {
	return func(client MQTT.Client, opts *MQTT.ClientOptions) {
		zlog.Warn().Str("addr", addr).Msg("MQTT正在尝试重新连接...")
	}
}

// Close 关闭MQTT连接
func Close() {
	if mqttClient != nil && mqttClient.IsConnected() {
		mqttClient.Disconnect(250) // 增加断开等待时间
		zlog.Info().Msg("MQTT连接已关闭")
	}
}

// WithClientID 设置客户端ID,仅仅作为前缀时会自动拼接随机串
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

// AuthWithUser 用户名密码认证
func AuthWithUser(username, pwd string) Option {
	return func(options *Options) {
		options.Username = username
		options.Password = pwd
	}
}

// AuthWithTLS TLS认证
func AuthWithTLS(certFile, keyFile string) Option {
	return func(options *Options) {
		cert, err := tls.LoadX509KeyPair(certFile, keyFile)
		if err != nil {
			// 更好的做法是在 Options 中增加一个 error 字段，在 NewClient 时检查
			zlog.Error().Err(err).Str("cert", certFile).Msg("加载 TLS 证书失败")
			return
		}
		options.tls = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
}

// Subscribe 订阅主题
func Subscribe(topic string, qos byte, callback MsgHandler) {
	mu.Lock()
	subscribes[topic] = callback
	subTopicQos[topic] = qos
	mu.Unlock()
	if mqttClient != nil && mqttClient.IsConnected() {
		token := mqttClient.Subscribe(topic, qos, func(client MQTT.Client, message MQTT.Message) {
			callback(message.MessageID(), message.Topic(), message.Payload())
		})
		if token.Wait() && token.Error() != nil {
			zlog.Error().Str("topic", topic).Err(token.Error()).Msg("MQTT订阅失败")
		} else {
			zlog.Debug().Str("topic", topic).Msg("MQTT订阅成功")
		}
	}
}

// Publish 发布消息
func Publish(topic string, qos byte, payload any) error {
	if mqttClient == nil || !mqttClient.IsConnected() {
		zlog.Error().Str("topic", topic).Msg("MQTT客户端未连接，发布失败")
		return fmt.Errorf("MQTT客户端未连接")
	}

	token := mqttClient.Publish(topic, qos, false, payload)
	if token.Wait() && token.Error() != nil {
		zlog.Error().Str("topic", topic).Err(token.Error()).Msg("MQTT发布失败")
		return token.Error()
	}
	zlog.Info().Str("topic", topic).Msg("MQTT发布成功")
	return nil
}

// IsConnected 连接状态检查
func IsConnected() bool {
	return mqttClient != nil && mqttClient.IsConnected()
}

// GetConnectionStatus 获取客户端状态
func GetConnectionStatus() string {
	if mqttClient == nil {
		return "未初始化"
	}
	if mqttClient.IsConnected() {
		return "已连接"
	}
	return "未连接"
}
