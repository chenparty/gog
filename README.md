# gog - Go 微服务开发工具集

`gog` 是一款用于简化 Go 微服务开发的工具包，集成了日志、客户端支持、HTTP 请求追踪及其他常用功能，帮助开发者快速构建和维护高效的微服务应用。

## 功能

### 1. 日志（Log）

支持结构化日志输出，提供统一的日志管理能力：

| 输出方式 | 说明 |
|---------|------|
| **Stdout** | 标准输出，适合开发环境 |
| **File** | 文件输出，基于 lumberjack 支持按大小和日期自动切割 |
| **NATS** | 消息队列输出，可通过 Vector 等工具采集转发 |

特性：
- 内置 Trace ID 支持，方便链路追踪
- 支持 Gin 中间件集成
- 支持 GORM SQL 日志插件
- 可自定义日志级别

### 2. 客户端（Client）

开箱即用的常见服务客户端：

| 客户端 | 说明 |
|-------|------|
| **MySQL** | 基于 GORM，支持连接池、慢查询日志、参数化查询 |
| **PostgreSQL** | 基于 GORM，支持时区配置、SSL 模式设置 |
| **Redis** | 支持单机/集群/哨兵模式 |
| **NATS** | 支持 JetStream 消息流 |
| **MQTT** | 基于 Paho，支持自动重连 |
| **Etcd** | 基于 clientv3 |
| **MinIO** | 对象存储客户端 |
| **HTTP** | 基于 Resty，支持请求追踪 |

### 3. Gin 中间件

| 中间件 | 说明 |
|-------|------|
| **GinRequestIDForTrace** | 请求 ID 生成与传递 |
| **GinLogger** | 请求/响应日志记录 |
| **Recovery** | Panic 恢复 |
| **IPRateLimit** | 基于 IP 的限流 |
| **RateLimit** | 全局令牌桶限流 |
| **IPWhitelist** | IP 白名单 |

## 安装

```shell
go get github.com/chenparty/gog
```

## 快速开始

### 初始化日志

```go
import "github.com/chenparty/gog/zlog"

// 使用文件输出
zlog.NewLogLogger("file", "info", zlog.FileAttr("log/app.log", 10, 7, true))

// 使用标准输出
zlog.NewLogLogger("stdout", "debug")
```

### 连接数据库

```go
import (
    "github.com/chenparty/gog/client/mysqlcli"
    "github.com/chenparty/gog/client/pgsqlcli"
    "github.com/chenparty/gog/client/rediscli"
)

// MySQL
mysqlcli.Connect(addr, user, pwd, dbName,
    mysqlcli.WithSilent(true),
    mysqlcli.WithSlowThreshold(time.Second),
)

// PostgreSQL
pgsqlcli.Connect(addr, user, pwd, dbName,
    pgsqlcli.WithTimeZone("Asia/Shanghai"),
    pgsqlcli.WithSSLMode("disable"),
)

// Redis
rediscli.Connect([]string{addr},
    rediscli.WithDB(0),
    rediscli.WithUserAndPass(user, pwd),
)
```

### Gin 中间件使用

```go
import (
    "github.com/chenparty/gog/zlog/ginplugin"
    "github.com/gin-gonic/gin"
)

func main() {
    r := gin.Default()

    // 链路追踪
    r.Use(ginplugin.GinRequestIDForTrace())

    // 请求日志
    r.Use(ginplugin.GinLogger(true, 2000))

    // Panic 恢复
    r.Use(ginplugin.Recovery(true))

    // IP 限流（面向用户侧服务）
    r.Use(ginplugin.IPRateLimit(100000, 3*time.Second, 50))

    // 全局限流（令牌桶）
    r.Use(ginplugin.RateLimit(time.Second, 100, 200))

    // IP 白名单
    r.Use(ginplugin.IPWhitelist([]string{"192.168.1.0/24", "10.0.0.1"}))

    r.Run(":8080")
}
```

### HTTP 客户端

```go
import "github.com/chenparty/gog/client/httpcli"

// POST 请求
statusCode, body, err := httpcli.PostJson(ctx, url, headers, requestBody)

// GET 请求
statusCode, body, err := httpcli.Get(ctx, url, headers, queryParams)
```

### 连接 MQTT

```go
import "github.com/chenparty/gog/client/mqttcli"

mqttcli.Connect(addr,
    mqttcli.WithClientID("my-client", false),
    mqttcli.AuthWithUser("username", "password"),
)

// 订阅主题
mqttcli.Subscribe("topic", 0, func(id uint16, topic string, payload []byte) {
    // 处理消息
})
```

### 连接 NATS

```go
import "github.com/chenparty/gog/client/natscli"

natscli.Connect("my-client", []string{"nats://localhost:4222"},
    natscli.WithUserAndPass("user", "pass"),
    natscli.WithJetStream(true),
)

// 发布消息
natscli.Publish("subject", []byte("hello"))

// 订阅消息
natscli.Subscribe("subject", func(msg *nats.Msg) {
    // 处理消息
})
```

## 项目结构

```
gog/
├── client/           # 客户端
│   ├── mysqlcli/     # MySQL 客户端
│   ├── pgsqlcli/    # PostgreSQL 客户端
│   ├── rediscli/    # Redis 客户端
│   ├── natscli/     # NATS 客户端
│   ├── mqttcli/     # MQTT 客户端
│   ├── etcdcli/     # Etcd 客户端
│   ├── miniocli/    # MinIO 客户端
│   └── httpcli/     # HTTP 客户端
├── zlog/             # 日志组件
│   ├── ginplugin/   # Gin 中间件
│   ├── gormplugin/  # GORM 插件
│   └── zwriter/     # 日志输出器
└── example/          # 使用示例
```

## 注意事项

- MySQL/PostgreSQL 连接失败时会 panic，建议在 K8s 环境中配置重启策略
- IP 限流基于内存缓存，适合面向用户的单实例服务
- 日志默认使用结构化 JSON 格式输出

## License

MIT
