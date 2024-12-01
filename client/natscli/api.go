package natscli

import (
	"context"
	"encoding/json"
	"github.com/chenparty/gog/zlog"
	"github.com/nats-io/nats.go"
	"time"
)

func Pub(ctx context.Context, subj string, data []byte) {
	err := nc.Publish(subj, data)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("subj", subj).Msg("nc.Publish")
	}
}

func PubGo(ctx context.Context, subj string, data any) {
	bs, err := json.Marshal(data)
	if err != nil {
		zlog.Error().Ctx(ctx).Err(err).Str("subj", subj).Msg("json.Marshal")
		return
	}
	Pub(ctx, subj, bs)
}
func Request(subj string, data []byte, timeout time.Duration) (*nats.Msg, error) {
	return nc.Request(subj, data, timeout)
}

func Sub(subj string, handler nats.MsgHandler) (err error) {
	_, err = nc.Subscribe(subj, handler)
	return
}

func QueueSub(subj, queue string, handler nats.MsgHandler) (err error) {
	_, err = nc.QueueSubscribe(subj, queue, handler)
	return
}

func QueueSubSyncWithChan(subject, queueName string, handler chan *nats.Msg) (sub *nats.Subscription, err error) {
	sub, err = nc.QueueSubscribeSyncWithChan(subject, queueName, handler)
	return
}
