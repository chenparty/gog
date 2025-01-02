package user

import (
	"encoding/json"
	"github.com/chenparty/gog/client/mqttcli"
	"github.com/chenparty/gog/example/internal/app/mq/resp"
	"github.com/chenparty/gog/zlog"
)

func (h *_defaultHandler) UserInfo(msgID uint16, topic string, payload []byte) {
	ctx := zlog.NewTraceContext()
	p := new(userInfoParam)
	if err := json.Unmarshal(payload, p); err != nil {
		zlog.Error().Ctx(ctx).Err(err).Msg("参数解析失败")
		_ = mqttcli.Publish(topic+"/reply", 0, resp.InvalidErr.Output())
		return
	}
	zlog.Info().Ctx(ctx).Msgf("参数解析成功:%+v", p)
	// TODO 参数详细校验

	data := h.userService.UserInfo(ctx, p.ID)
	_ = mqttcli.Publish(topic+"/reply", 0, resp.OK.Output().WithData(data))
}

func (h *_defaultHandler) AddUser(msgID uint16, topic string, payload []byte) {
	ctx := zlog.NewTraceContext()
	p := new(addUserParam)
	if err := json.Unmarshal(payload, p); err != nil {
		zlog.Error().Ctx(ctx).Err(err).Msg("参数解析失败")
		_ = mqttcli.Publish(topic+"/reply", 0, resp.InvalidErr.Output())
		return
	}
	zlog.Info().Ctx(ctx).Msgf("参数解析成功:%+v", p)
	// TODO 参数详细校验

	h.userService.AddUser(ctx, p.Name)
	_ = mqttcli.Publish(topic+"/reply", 0, resp.OK.Output())
}
