package mqttcli

import (
	"github.com/chenparty/gog/zlog"
	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/oklog/ulid/v2"
)

var subscribes = map[string]MQTT.MessageHandler{}
var mqttClient MQTT.Client

type Options struct {
	ClientID string // 客户端ID,不设置时会自动随机生成
	Username string // 用户名
	Password string // 密码
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
	clientOptions.SetUsername(opts.Username)
	clientOptions.SetPassword(opts.Password)
	clientOptions.OnConnect = func(client MQTT.Client) {
		zlog.Info().Str("addr", addr).Msg("MQTT连接成功")
		// 连接后自动订阅Topic
		for key, sub := range subscribes {
			client.Subscribe(key, 0, sub)
		}
	}
	clientOptions.OnConnectionLost = func(client MQTT.Client, e error) {
		zlog.Error().Str("addr", addr).Err(e).Msg("MQTT连接断开")
	}
	mqttClient = MQTT.NewClient(clientOptions)
	if token := mqttClient.Connect(); token.Wait() && token.Error() != nil {
		zlog.Error().Str("addr", addr).Err(token.Error()).Msg("MQTT连接失败")
	}
}

// Close 关闭MQTT连接
func Close() {
	mqttClient.Disconnect(200)
}

// WithClientID 设置客户端ID,仅仅作为前缀时会自动拼接随机串
func WithClientID(clientID string, asPrefix bool) Option {
	return func(options *Options) {
		if asPrefix {
			options.ClientID = clientID + "-" + options.ClientID
		} else {
			options.ClientID = clientID
		}
	}
}

// WithUserAndPass 设置用户名和密码
func WithUserAndPass(username, pwd string) Option {
	return func(options *Options) {
		options.Username = username
		options.Password = pwd
	}
}

// Subscribe 订阅主题
func Subscribe(topic string, callback MQTT.MessageHandler) {
	subscribes[topic] = callback
	if mqttClient.IsConnected() {
		if token := mqttClient.Subscribe(topic, 0, callback); token.Wait() && token.Error() != nil {
			zlog.Error().Str("topic", topic).Err(token.Error()).Msg("MQTT订阅失败")
		}
	}
}

// Publish 发布消息
func Publish(topic string, qos byte, payload any) error {
	if token := mqttClient.Publish(topic, qos, false, payload); token.Wait() && token.Error() != nil {
		zlog.Error().Str("topic", topic).Err(token.Error()).Msg("MQTT发布失败")
		return token.Error()
	}
	zlog.Info().Str("topic", topic).Msg("MQTT发布成功")
	return nil
}
