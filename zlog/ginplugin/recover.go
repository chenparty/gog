package ginplugin

import (
	"errors"
	"fmt"
	"github.com/chenparty/gog/zlog"
	"github.com/gin-gonic/gin"
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
)

// Recovery recover掉项目可能出现的panic，并使用zlog记录相关日志
func Recovery(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.,
				var brokenPipe bool
				var ne *net.OpError
				if errors.As(err.(*net.OpError), &ne) {
					var se *os.SyscallError
					if errors.As(ne.Err, &se) {
						if strings.Contains(strings.ToLower(se.Error()), "broken pipe") || strings.Contains(strings.ToLower(se.Error()), "connection reset by peer") {
							brokenPipe = true
						}
					}
				}

				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					zlog.Error().Ctx(c.Request.Context()).Msg(fmt.Sprint("request:", string(httpRequest), err))
					// If the connection is dead, we can't write a status to it.
					_ = c.Error(err.(error))
					c.Abort()
					return
				}

				if stack {
					zlog.Error().Ctx(c.Request.Context()).Msg(fmt.Sprint("Recovery from panic:", " stack:", string(debug.Stack()), err))
				} else {
					zlog.Error().Ctx(c.Request.Context()).Msg(fmt.Sprint("Recovery from panic:", err))
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
