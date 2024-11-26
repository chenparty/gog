package middleware

import (
	"github.com/gin-gonic/gin"
)

const RequestParam = "request_param"

// ParseJson json传参校验
func ParseJson[T any](c *gin.Context) {
	body := new(T)
	if err := c.ShouldBindJSON(&body); err != nil {
		//zlog.Error("ValidateJsonBody", "ShouldBindJSON").Msg(fmt.Sprint("["+c.HandlerName()+"]", err.Error()))
		//api_resp.InvalidErr.Output().Json(c)
		c.Abort()
		return
	}
	//zlog.Info("ValidateJsonBody", c.HandlerName()).Any("data", body).Send()
	c.Set(RequestParam, body)
	c.Next()
	return
}

// ParseQuery query传参校验
func ParseQuery[T any](c *gin.Context) {
	body := new(T)
	if err := c.ShouldBindQuery(&body); err != nil {
		//zlog.Error("ValidateQueryData", "ShouldBindQuery").Msg(fmt.Sprint("["+c.HandlerName()+"]", err.Error()))
		//api_resp.InvalidErr.Output().Json(c)
		c.Abort()
		return
	}
	//zlog.Info("ValidateQueryData", c.HandlerName()).Any("data", body).Send()
	c.Set(RequestParam, body)
	c.Next()
	return
}

func GetRequestParam[BodyType any](c *gin.Context) *BodyType {
	return c.MustGet(RequestParam).(*BodyType)
}
