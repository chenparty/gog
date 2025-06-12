package main

import (
	"flag"
	"github.com/chenparty/gog/client/mqttcli"
	"github.com/chenparty/gog/client/mysqlcli"
	"github.com/chenparty/gog/example/config"
	"github.com/chenparty/gog/example/internal/app/api"
	"github.com/chenparty/gog/example/internal/app/mq"
	"github.com/chenparty/gog/zlog"
)

func init() {
	var configFile string
	flag.StringVar(&configFile, "conf", "example/config/app.yaml", "specify user config file path")
	flag.Parse()
	// 初始化配置
	config.InitConfig(configFile)
	// 初始化日志
	if config.Get().Release {
		zlog.NewLogLogger("file", "info", zlog.FileAttr("log/mtbar.log", 2, 7, true))
	}
}

func main() {
	cfg := config.Get()
	mysqlcli.Connect(cfg.Mysql.Addr, cfg.Mysql.User, cfg.Mysql.Pwd, cfg.Mysql.DbName)
	mqttcli.Connect(cfg.Mqtt.Addr, mqttcli.AuthWithUser(cfg.Mqtt.User, cfg.Mqtt.Pwd))
	mq.InitSubscription()
	api.Init(cfg.Release)
}
