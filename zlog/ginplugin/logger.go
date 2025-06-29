package ginplugin

import (
	"bytes"
	"github.com/chenparty/gog/zlog"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"io"
	"time"
)

// GinLogger 日志
func GinLogger(autoCopyRequestBody bool, maxLogBodySize int) gin.HandlerFunc {
	return func(c *gin.Context) {
		//先打印请求头信息
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if raw != "" {
			path = path + "?" + raw
		}

		reqEvent := zlog.Info().Ctx(c.Request.Context()).
			Str("method", c.Request.Method).
			Str("path", path).
			Str("client_ip", c.ClientIP()).
			Str("user_agent", c.Request.UserAgent()).
			Str("content_type", c.ContentType())
		if autoCopyRequestBody && c.Request.Method == "POST" {
			b, _ := c.Copy().GetRawData()
			reqEvent.Str("req_body", string(b))
			// 重置请求体
			c.Request.Body = io.NopCloser(bytes.NewReader(b))
		}
		start := time.Now()
		reqEvent.Msg("GinRequest")

		// 创建自定义的 ResponseWriter
		customWriter := &CustomResponseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBuffer([]byte{}),
		}
		// 替换原始的 ResponseWriter
		c.Writer = customWriter
		// 接着交给具体接口处理请求
		c.Next()
		// 等接口处理完后，拿到请求体和响应体打印
		duration := time.Now().Sub(start)
		var respEvent *zerolog.Event
		if c.Writer.Status() >= 400 && c.Writer.Status() < 500 {
			respEvent = zlog.Warn()
		} else if c.Writer.Status() >= 500 {
			respEvent = zlog.Error()
		} else {
			respEvent = zlog.Info()
		}
		val, isExist := getRequestBody(c)
		if isExist {
			requestBody, ok := val.([]byte)
			if ok {
				respEvent = respEvent.Str("req_body", string(requestBody))
			} else {
				respEvent = respEvent.Any("req_body", val)
			}
		}
		// 获取响应的字节内容,太大的也不打印
		if maxLogBodySize <= 0 {
			maxLogBodySize = 2000
		}
		if c.Writer.Size() > 0 && c.Writer.Size() < maxLogBodySize {
			responseBody := customWriter.body
			respEvent.Str("resp_body", string(responseBody.Bytes()))
		}
		// 打印
		respEvent.Ctx(c.Request.Context()).
			Str("path", path).
			Int("status", c.Writer.Status()).
			Dur("duration", duration).
			Int("body_size", c.Writer.Size()).
			Msg("GinRequest")
	}
}

// CustomResponseWriter 自定义的 ResponseWriter
type CustomResponseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (c *CustomResponseWriter) Write(p []byte) (n int, err error) {
	// 捕获写入的内容
	c.body.Write(p)
	// 继续写入原始的 ResponseWriter
	return c.ResponseWriter.Write(p)
}

// RequestBodyKey 用于日志输出请求体，避免二次解包
// 调用时可以在请求Handler里解析完参数后再defer SetRequestBody即可
const RequestBodyKey = "RequestBodyKey"

func getRequestBody(c *gin.Context) (any, bool) {
	return c.Get(RequestBodyKey)
}

func LogRequestBody(c *gin.Context, body any) {
	c.Set(RequestBodyKey, body)
}
