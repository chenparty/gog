package user

import (
	"github.com/chenparty/gog/example/internal/app/api/service/user"
	"github.com/gin-gonic/gin"
)

type Handler interface {
	UserInfo(c *gin.Context)
	AddUser(c *gin.Context)
}

func NewHandler(userService user.Service) Handler {
	return &_defaultHandler{
		userService: userService,
	}
}

type _defaultHandler struct {
	userService user.Service
}
