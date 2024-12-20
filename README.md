# gog - Go 微服务开发工具集

`gog` 是一款用于简化 Go 微服务开发的工具包，集成了日志、客户端支持以及其他常用功能，帮助开发者快速构建和维护高效的微服务应用。

## 功能

### 1. **日志（Log）**
- 支持在 Gin 和 GORM 中进行追踪（Trace）。
- 支持多种日志输出方式：
  - **stdout**（标准输出）
  - **file**（文件输出，按日期大小自动分割）
  - **NATS**（消息队列，可通过vector采集转发）

### 2. **客户端（Client）**
- 支持连接以下常见服务：
  - **MySQL**
  - **Redis**
  - **NATS**
  - **MQTT**
  - **Etcd**
  - **Minio**

## 安装

使用以下命令安装 `gog` 工具包：

```shell
go get github.com/chenparty/gog
