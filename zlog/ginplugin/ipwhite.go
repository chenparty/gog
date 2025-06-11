package ginplugin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"net"
	"strings"
)

// IPWhitelist 创建IP白名单中间件
func IPWhitelist(whitelist []string) gin.HandlerFunc {
	// 预处理白名单：解析为IPNet对象
	ipNets := make([]*net.IPNet, 0, len(whitelist))
	for _, item := range whitelist {
		// 处理单个IP（如 "192.168.1.1"）
		if !strings.Contains(item, "/") {
			item += "/32" // 单个IPv4转换为CIDR
		}

		// 处理IPv6单个地址（如 "::1"）
		if strings.Count(item, ":") >= 2 && !strings.Contains(item, "/") {
			item += "/128"
		}

		_, ipNet, err := net.ParseCIDR(item)
		if err != nil {
			// 解析失败直接panic（建议在服务启动时检查）
			panic(fmt.Sprintf("invalid CIDR %s: %v", item, err))
		}
		ipNets = append(ipNets, ipNet)
	}

	return func(c *gin.Context) {
		// 获取客户端真实IP
		clientIP := net.ParseIP(c.ClientIP())
		if clientIP == nil {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		// 检查IP是否在白名单中
		allowed := false
		for _, ipNet := range ipNets {
			if ipNet.Contains(clientIP) {
				allowed = true
				break
			}
		}

		if !allowed {
			c.AbortWithStatusJSON(403, gin.H{"error": "Forbidden"})
			return
		}

		c.Next()
	}
}
