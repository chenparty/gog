package ginplugin

import (
	"github.com/chenparty/gog/zlog"
	"github.com/gin-contrib/requestid"
	"github.com/gin-gonic/gin"
)

// GinRequestIDForTrace gin middleware for request id
func GinRequestIDForTrace() gin.HandlerFunc {
	return requestid.New(requestid.WithGenerator(func() string {
		return zlog.NewTraceID()
	}), requestid.WithHandler(func(c *gin.Context, id string) {
		ctx := zlog.ContextWithValue(c.Request.Context(), requestid.Get(c))
		c.Request = c.Request.WithContext(ctx)
	}))
}
