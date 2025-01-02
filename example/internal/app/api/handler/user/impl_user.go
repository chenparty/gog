package user

import (
	"github.com/chenparty/gog/example/internal/app/api/resp"
	"github.com/chenparty/gog/zlog"
	"github.com/gin-gonic/gin"
)

func (h *_defaultHandler) UserInfo(c *gin.Context) {
	p := new(userInfoParam)
	if err := c.ShouldBindQuery(p); err != nil {
		zlog.Error().Ctx(c).Err(err).Msg("参数解析失败")
		resp.InvalidErr.Output().Json(c)
		return
	}
	zlog.Info().Ctx(c).Msgf("参数解析成功:%+v", p)
	// TODO 参数详细校验

	h.userService.UserInfo(c, p.ID).Json(c)
}

func (h *_defaultHandler) AddUser(c *gin.Context) {
	p := new(addUserParam)
	if err := c.ShouldBindJSON(p); err != nil {
		zlog.Error().Ctx(c).Err(err).Msg("参数解析失败")
		resp.InvalidErr.Output().Json(c)
		return
	}
	zlog.Info().Ctx(c).Msgf("参数解析成功:%+v", p)
	// TODO 参数详细校验

	h.userService.AddUser(c, p.Name).Json(c)
}
