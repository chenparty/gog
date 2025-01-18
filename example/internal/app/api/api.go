package api

import (
	"github.com/chenparty/gog/example/config"
	"github.com/chenparty/gog/example/internal/app/api/handler/user"
	userService "github.com/chenparty/gog/example/internal/app/api/service/user"
	"github.com/chenparty/gog/zlog/ginplugin"
	"github.com/gin-gonic/gin"
)

func Init(release bool) {
	if release {
		gin.SetMode(gin.ReleaseMode)
	}
	g := gin.New()
	g.MaxMultipartMemory = 10 << 20 // 设置请求最大体积为 10 MB, 防止恶意请求
	g.Use(ginplugin.GinRequestIDForTrace())
	g.Use(ginplugin.GinLogger(true))
	g.Use(ginplugin.Recovery(true))
	registryRouter(g)
	err := g.Run(config.Get().Http.Addr)
	if err != nil {
		panic(err)
	}
}

func registryRouter(r *gin.Engine) {
	// v1版本路由
	v1 := r.Group("v1")
	{
		// 用户模块接口
		uh := user.NewHandler(userService.NewService())
		userGroup := v1.Group("user")
		{
			userGroup.GET("user", uh.UserInfo)
			userGroup.POST("user", uh.AddUser)
		}
	}
}
