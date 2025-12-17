package middleware

import (
	"net/http"
	"sync"
	"time"

	"jd-task-platform-go/pkg/response"

	"github.com/gin-gonic/gin"
)

// RateLimiter 限流器结构
type RateLimiter struct {
	mu         sync.Mutex
	visitors   map[string]*visitorInfo
	maxCalls   int           // 最大调用次数
	window     time.Duration // 时间窗口
	cleanupAge time.Duration // 清理过期数据的时间
}

type visitorInfo struct {
	calls     []time.Time // 调用时间记录
	lastVisit time.Time   // 最后访问时间
}

// NewRateLimiter 创建新的限流器
// maxCalls: 时间窗口内的最大调用次数
// window: 时间窗口大小
func NewRateLimiter(maxCalls int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		visitors:   make(map[string]*visitorInfo),
		maxCalls:   maxCalls,
		window:     window,
		cleanupAge: time.Minute * 5, // 5分钟清理一次过期数据
	}
	go rl.cleanup()
	return rl
}

// cleanup 定期清理过期的访客数据
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.cleanupAge)
	defer ticker.Stop()
	for range ticker.C {
		rl.mu.Lock()
		for key, v := range rl.visitors {
			if time.Since(v.lastVisit) > rl.cleanupAge {
				delete(rl.visitors, key)
			}
		}
		rl.mu.Unlock()
	}
}

// IsAllowed 检查是否允许访问
func (rl *RateLimiter) IsAllowed(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists {
		rl.visitors[key] = &visitorInfo{
			calls:     []time.Time{now},
			lastVisit: now,
		}
		return true
	}

	// 清理窗口外的调用记录
	validCalls := make([]time.Time, 0)
	for _, t := range v.calls {
		if now.Sub(t) <= rl.window {
			validCalls = append(validCalls, t)
		}
	}

	// 检查是否超过限制
	if len(validCalls) >= rl.maxCalls {
		v.lastVisit = now
		return false
	}

	// 添加本次调用
	v.calls = append(validCalls, now)
	v.lastVisit = now
	return true
}

// GetRemainingCalls 获取剩余调用次数
func (rl *RateLimiter) GetRemainingCalls(key string) int {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	v, exists := rl.visitors[key]
	if !exists {
		return rl.maxCalls
	}

	// 计算有效调用次数
	validCount := 0
	for _, t := range v.calls {
		if now.Sub(t) <= rl.window {
			validCount++
		}
	}

	remaining := rl.maxCalls - validCount
	if remaining < 0 {
		remaining = 0
	}
	return remaining
}

// 全局API Key限流器 (1秒2次)
var apiKeyRateLimiter = NewRateLimiter(2, time.Second)

// APIKeyRateLimitMiddleware API Key限流中间件
// 基于API Key进行限流，每秒最多2次调用
func APIKeyRateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从context获取API Key（由APIKeyMiddleware设置）
		apiKey, exists := c.Get("api_key")
		if !exists {
			// 如果没有API Key，尝试从header获取
			apiKey = c.GetHeader("X-API-KEY")
		}

		key := ""
		if apiKey != nil {
			key = apiKey.(string)
		}
		if key == "" {
			// 没有API Key，跳过限流
			c.Next()
			return
		}

		// 检查是否允许访问
		if !apiKeyRateLimiter.IsAllowed(key) {
			remaining := apiKeyRateLimiter.GetRemainingCalls(key)
			c.Header("X-RateLimit-Limit", "2")
			c.Header("X-RateLimit-Remaining", string(rune('0'+remaining)))
			c.Header("X-RateLimit-Reset", "1")
			response.Error(c, http.StatusTooManyRequests, "请求过于频繁，请稍后再试（限制：每秒最多2次）")
			c.Abort()
			return
		}

		// 设置限流响应头
		remaining := apiKeyRateLimiter.GetRemainingCalls(key)
		c.Header("X-RateLimit-Limit", "2")
		c.Header("X-RateLimit-Remaining", string(rune('0'+remaining)))
		c.Header("X-RateLimit-Reset", "1")

		c.Next()
	}
}
