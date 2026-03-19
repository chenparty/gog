package main

import (
	"github.com/chenparty/gog/client/mqttcli"
	"github.com/chenparty/gog/client/mysqlcli"
	"github.com/chenparty/gog/example/config/app"
	"github.com/chenparty/gog/example/internal/app/api"
	"github.com/chenparty/gog/example/internal/app/mq"
	"github.com/chenparty/gog/zlog"
	"github.com/joho/godotenv"
	"log"
	"path/filepath"
)

func init() {
	envPath := filepath.Join("example/config/app", ".env")
	if err := godotenv.Load(envPath); err != nil {
		log.Printf("No .env file found for service1, using system environment variables")
	}
	// 初始化配置
	app.InitEnv()
	// 初始化日志
	if app.Get().Release {
		zlog.NewLogLogger("file", "info", zlog.FileAttr("log/mtbar.log", 2, 7, true))
	}
}

func main() {
	cfg := app.Get()
	mysqlcli.Connect(cfg.Mysql.Addr, cfg.Mysql.User, cfg.Mysql.Pwd, cfg.Mysql.DbName)
	mqttcli.Connect(cfg.Mqtt.Addr, mqttcli.AuthWithUser(cfg.Mqtt.User, cfg.Mqtt.Pwd))
	mq.InitSubscription()
	api.Init(cfg.Release)
}
