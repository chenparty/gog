package app

import (
	"fmt"
	"github.com/caarlos0/env/v11"
)

type AppConfig struct {
	Release bool
	Http    struct {
		Addr string `env:"HTTP_ADDR"`
	}
	Mysql struct {
		Addr   string `env:"MYSQL_ADDR"`
		User   string `env:"MYSQL_USER"`
		Pwd    string `env:"MYSQL_PWD"`
		DbName string `env:"MYSQL_DB_NAME"`
	}
	Mqtt struct {
		Addr string `env:"MQTT_ADDR"`
		User string `env:"MQTT_USER"`
		Pwd  string `env:"MQTT_PWD"`
	}
}

// 用于存储唯一的配置实例
var cfg AppConfig
var initialized bool

// InitEnv 初始化环境变量
func InitEnv() {
	var err error
	cfg, err = env.ParseAs[AppConfig]()
	if err != nil {
		fmt.Printf("错误：环境变量加载出现错误：%v\n", err)
		panic("环境变量加载出现错误")
	}
	initialized = true
}

// Get 获取环境变量
func Get() *AppConfig {
	if !initialized {
		panic("配置尚未初始化，请先调用 InitEnv() 初始化配置")
	}
	return &cfg
}
