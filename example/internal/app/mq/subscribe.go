package mq

import (
	"github.com/chenparty/gog/client/mqttcli"
	"github.com/chenparty/gog/example/internal/app/mq/handler/user"
	userService "github.com/chenparty/gog/example/internal/app/mq/service/user"
)

func InitSubscription() {
	// 用户模块接口
	uh := user.NewHandler(userService.NewService())
	mqttcli.Subscribe("user/info", 0, uh.UserInfo)
	mqttcli.Subscribe("user/add", 0, uh.AddUser)
}
