package resp

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

type Output struct {
	Code string `json:"code"`
	Msg  string `json:"msg"`
	Data any    `json:"data"`
}

func (o *Output) WithData(data any) *Output {
	o.Data = data
	return o
}

func (o *Output) WithMsg(msg string) *Output {
	o.Msg = msg
	return o
}

func (o *Output) Json(c *gin.Context) {
	switch State(o.Code) {
	case OK:
		c.JSON(http.StatusOK, o)
	case InvalidErr:
		c.JSON(http.StatusBadRequest, o)
	case UnauthorizedErr:
		c.JSON(http.StatusUnauthorized, o)
	default:
		c.JSON(http.StatusInternalServerError, o)
	}
}
