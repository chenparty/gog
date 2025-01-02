package user

import (
	"github.com/chenparty/gog/example/internal/app/mq/service/user"
)

type Handler interface {
	UserInfo(msgID uint16, topic string, payload []byte)
	AddUser(msgID uint16, topic string, payload []byte)
}

func NewHandler(userService user.Service) Handler {
	return &_defaultHandler{
		userService: userService,
	}
}

type _defaultHandler struct {
	userService user.Service
}
