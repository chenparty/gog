package ginplugin

import (
	"github.com/chenparty/gog/zlog"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func GinRequestIDForTrace() gin.HandlerFunc {
	return requestid.New(requestid.WithHandler(func(c *gin.Context, id string) {
		ctx := zlog.ContextWithValue(c.Request.Context(), requestid.Get(c))
		c.Request = c.Request.WithContext(ctx)
	}))
}
