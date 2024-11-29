package ginplugin

import (
	"context"
	"github.com/chenparty/gog/zlog"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

func GinRequestIDForTrace() gin.HandlerFunc {
	return requestid.New(requestid.WithHandler(func(c *gin.Context, id string) {
		ctx := context.WithValue(c.Request.Context(), zlog.CtxTraceIDKey, requestid.Get(c))
		c.Request = c.Request.WithContext(ctx)
	}))
}
