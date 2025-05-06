package ginplugin

import (
	"github.com/chenparty/gog/zlog"
	"github.com/gin-gonic/gin"
)

const HeaderRequestID = "Z-Request-ID"

// GinRequestIDForTrace gin middleware for request id
func GinRequestIDForTrace(allowedRequestIDs ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		if len(allowedRequestIDs) == 0 {
			allowedRequestIDs = []string{HeaderRequestID}
		}
		var rid string
		for _, id := range allowedRequestIDs {
			if rid = c.GetHeader(id); len(rid) > 0 {
				continue
			}
		}
		if rid == "" {
			rid = generatorRequestID()
		}
		handleRequest(c, rid)

		c.Header(HeaderRequestID, rid)
		c.Next()
	}
}

func generatorRequestID() string {
	return zlog.NewTraceID()
}

func handleRequest(c *gin.Context, traceID string) {
	ctx := zlog.NewTraceContextWithID(traceID)
	c.Request = c.Request.WithContext(ctx)
}
