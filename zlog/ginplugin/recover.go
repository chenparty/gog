package ginplugin

import (
	"fmt"
	"github.com/chenparty/gog/zlog"
	"github.com/gin-gonic/gin"
	"net/http"
	"runtime/debug"
)

// Recovery recover掉项目可能出现的panic，并使用zlog记录相关日志
func Recovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				if stack {
					zlog.Error().Ctx(c.Request.Context()).Msg(fmt.Sprint("Recovery from panic:", err, " stack:", string(debug.Stack())))
				} else {
					zlog.Error().Ctx(c.Request.Context()).Msg(fmt.Sprint("Recovery from panic:", err))
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
