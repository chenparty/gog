package config

import (
	"fmt"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
	"os"
	"strings"
	"sync"
)

type AppConfig struct {
	Release bool
	Http    struct {
		Addr string `toml:"addr" yaml:"addr"`
	}
	Mysql struct {
		Addr   string `toml:"addr" yaml:"addr"`
		User   string `toml:"user" yaml:"user"`
		Pwd    string `toml:"pwd" yaml:"pwd"`
		DbName string `toml:"db_name" yaml:"db_name"`
	}
	Mqtt struct {
		Addr string `toml:"addr" yaml:"addr"`
		User string `toml:"user" yaml:"user"`
		Pwd  string `toml:"pwd" yaml:"pwd"`
	}
}

// 用于存储唯一的配置实例
var configInstance *AppConfig
var once sync.Once

// InitConfig 初始化配置
func InitConfig(filepath string) {
	if filepath == "" {
		panic("配置文件路径不能为空")
	}

	once.Do(func() {
		file, err := os.Open(filepath)
		if err != nil {
			panic(fmt.Sprintf("无法打开配置文件: %v", err))
		}
		defer file.Close()

		var config AppConfig
		switch {
		case strings.HasSuffix(filepath, ".toml"):
			if err = toml.NewDecoder(file).Decode(&config); err != nil {
				panic(fmt.Sprintf("解析 TOML 配置文件失败: %v", err))
			}
		case strings.HasSuffix(filepath, ".yaml"), strings.HasSuffix(filepath, ".yml"):
			if err = yaml.NewDecoder(file).Decode(&config); err != nil {
				panic(fmt.Sprintf("解析 YAML 配置文件失败: %v", err))
			}
		default:
			panic("不支持的配置文件格式")
		}
		configInstance = &config
	})
}

// Get 获取配置，返回配置实例和错误
func Get() *AppConfig {
	if configInstance == nil {
		panic("配置尚未初始化，请先调用 InitConfig() 初始化配置")
	}
	return configInstance
}
