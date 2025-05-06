package ginplugin

import (
	"github.com/gin-gonic/gin"
	"github.com/maypok86/otter"
	"golang.org/x/time/rate"
	"net/http"
	"strings"
	"time"
)

type requestInfo struct {
	LastAccessTime time.Time // 上次访问时间
	RequestNum     int       // 请求计数
}

var (
	requestInfoCache     otter.Cache[string, *requestInfo] // IP与请求信息的映射
	defaultMaxRequests   = 50                              // 默认允许的最大请求数
	defaultTimeWindow    = 3 * time.Second                 // 默认时间窗口
	defaultCacheCapacity = 100_000                         // 默认缓存容量
	defaultCacheExpire   = 1 * time.Minute                 // 默认缓存过期时间
)

// IPRateLimit IP限流器-基于内存，建议在面向用户侧服务使用
func IPRateLimit(ipCacheCapacity int, timeWindow time.Duration, maxRequests int) gin.HandlerFunc {
	// 校验参数，判断是否使用默认值
	if ipCacheCapacity <= 0 {
		ipCacheCapacity = defaultCacheCapacity
	}
	if timeWindow <= 0 {
		timeWindow = defaultTimeWindow
	}
	if maxRequests <= 0 {
		maxRequests = defaultMaxRequests
	}
	// 初始化请求限流信息缓存
	var err error
	requestInfoCache, err = otter.MustBuilder[string, *requestInfo](ipCacheCapacity).WithTTL(defaultCacheExpire).Build()
	if err != nil {
		panic(err)
	}
	return func(c *gin.Context) {
		userAgent := c.Request.UserAgent()
		// 如果是微服务内部调用放行
		if strings.HasPrefix(userAgent, "go-resty") && c.Request.Header.Get(HeaderRequestID) != "" {
			c.Next()
			return
		}
		ip := c.ClientIP()
		info, ok := requestInfoCache.Get(ip)
		// 如果IP不存在，初始化并添加到缓存中，并放行
		if !ok {
			requestInfoCache.Set(ip, &requestInfo{LastAccessTime: time.Now(), RequestNum: 1})
			c.Next()
			return
		}
		// 如果超过时间窗口，重置请求计数，并放行
		if time.Since(info.LastAccessTime) > timeWindow {
			info.RequestNum = 1
			info.LastAccessTime = time.Now()
			requestInfoCache.Set(ip, info)
			c.Next()
			return
		}
		// 如果在时间窗口内，增加请求计数
		info.RequestNum++
		// 如果请求计数超过限制，禁止访问
		if info.RequestNum > maxRequests {
			// 如果请求被限制，返回 429 状态码
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试！",
			})
			c.Abort()
			return
		}
		// 更新最后访问时间
		info.LastAccessTime = time.Now()
		c.Next()
	}
}

// RateLimit 全局限流器（令牌桶）-基于内存
func RateLimit(time time.Duration, rps, burst int) gin.HandlerFunc {
	// 创建一个速率限制器（令牌桶），每秒最大 `rps` 次请求，最多突发 `burst` 次请求
	limiter := rate.NewLimiter(rate.Limit(rps), burst)
	return func(c *gin.Context) {
		// 限制请求数量
		if !limiter.Allow() {
			// 如果请求被限制，返回 429 状态码
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error": "请求过于频繁，请稍后再试！",
			})
			c.Abort()
			return
		}
	}
}
