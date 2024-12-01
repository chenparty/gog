package natscli

import (
	"errors"
	"github.com/nats-io/nats.go"
)

var jsc nats.JetStreamContext

func newJetStreamContext() (err error) {
	jsc, err = nc.JetStream(nats.PublishAsyncMaxPending(256))
	return
}

// AddStream 添加流（不会重复创建，也不会删除再创建）
func AddStream(streamName string, subjects []string) error {
	// 按名称查找流信息
	_, err := jsc.StreamInfo(streamName)
	if err != nil {
		// 流不存在
		if errors.Is(err, nats.ErrStreamNotFound) {
			// 创建流
			_, err = jsc.AddStream(&nats.StreamConfig{
				Name:      streamName,
				Subjects:  subjects,
				NoAck:     false, // false:自动ack | true:手动ack
				Retention: nats.WorkQueuePolicy,
			})
			return err
		}
		// 其他错误
		return err
	}
	// 流已存在
	return nil
}

// DelStream 删除流
func DelStream(streamName string) (err error) {
	err = jsc.DeleteStream(streamName)
	return
}

// AddConsumer 添加流消费者
func AddConsumer(streamName string, config *nats.ConsumerConfig) (err error) {
	_, err = jsc.AddConsumer(streamName, config)
	return
}

// DelConsumer 删除流消费者
func DelConsumer(stream, consumer string) (err error) {
	err = jsc.DeleteConsumer(stream, consumer)
	return
}

// JsPub 发布流消息
func JsPub(subj string, data []byte) (err error) {
	_, err = jsc.Publish(subj, data)
	return
}

// JsSub 订阅流消息
func JsSub(subj string, handler nats.MsgHandler) (err error) {
	_, err = jsc.Subscribe(subj, handler)
	if err != nil {
		return
	}
	return
}

// JsQueueSubscribe 队列方式订阅流消息(分布式场景使用)
func JsQueueSubscribe(subject, queueName string, handler nats.MsgHandler) (err error) {
	_, err = jsc.QueueSubscribe(subject, queueName, handler)
	return
}
