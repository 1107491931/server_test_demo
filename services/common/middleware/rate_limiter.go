package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

type client struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// RateLimiter 管理不同客户端 (IP) 的速率限制器
type RateLimiter struct {
	clients     map[string]*client
	mu          sync.RWMutex // 使用读写锁提高并发性能
	r           rate.Limit
	b           int
	cleanupTime time.Duration // 清理过期客户端的时间阈值
	stopChan    chan struct{} // 用于停止清理 goroutine
}

// RateLimiterOption 配置选项函数
type RateLimiterOption func(*RateLimiter)

// NewRateLimiter 创建一个新的 RateLimiter
func NewRateLimiter(r rate.Limit, b int, opts ...RateLimiterOption) *RateLimiter {
	rl := &RateLimiter{
		clients:     make(map[string]*client),
		r:           r,
		b:           b,
		cleanupTime: 3 * time.Minute, // 默认 3 分钟
		stopChan:    make(chan struct{}),
	}

	// 应用配置选项
	for _, opt := range opts {
		opt(rl)
	}

	// 启动后台清理协程
	go rl.cleanup()

	return rl
}

// WithCleanupTime 设置清理时间阈值
func WithCleanupTime(d time.Duration) RateLimiterOption {
	return func(rl *RateLimiter) {
		rl.cleanupTime = d
	}
}

// getLimiter 返回指定 IP 的速率限制器（使用读写锁优化）
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	// 先尝试读锁（大部分情况命中缓存）
	rl.mu.RLock()
	c, exists := rl.clients[ip]
	if exists {
		c.lastSeen = time.Now()
		limiter := c.limiter
		rl.mu.RUnlock()
		return limiter
	}
	rl.mu.RUnlock()

	// 缓存未命中，需要创建时再加写锁
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// 双重检查，防止并发创建
	if c, exists := rl.clients[ip]; exists {
		c.lastSeen = time.Now()
		return c.limiter
	}

	// 创建新的限流器
	limiter := rate.NewLimiter(rl.r, rl.b)
	rl.clients[ip] = &client{
		limiter:  limiter,
		lastSeen: time.Now(),
	}

	return limiter
}

// cleanup 安全的清理机制，支持优雅停止
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			rl.cleanupStaleClients()
		case <-rl.stopChan:
			return
		}
	}
}

// cleanupStaleClients 清理过期客户端
func (rl *RateLimiter) cleanupStaleClients() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	for ip, client := range rl.clients {
		if time.Since(client.lastSeen) > rl.cleanupTime {
			delete(rl.clients, ip)
		}
	}
}

// Stop 停止清理 goroutine，防止资源泄漏
func (rl *RateLimiter) Stop() {
	close(rl.stopChan)
}

// Allow 检查是否允许请求
func (rl *RateLimiter) Allow(ip string) bool {
	return rl.getLimiter(ip).Allow()
}

// Middleware 创建 Gin 中间件
func (rl *RateLimiter) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()

		if !rl.Allow(ip) {
			// 添加 RateLimit 相关的 HTTP 头部
			c.Header("X-RateLimit-Limit", strconv.Itoa(int(rl.r)))
			c.Header("X-RateLimit-Remaining", "0")
			c.Header("Retry-After", "1")

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"code":        http.StatusTooManyRequests,
				"message":     "请求过于频繁，请稍后再试",
				"retry_after": 1,
			})
			return
		}

		c.Next()
	}
}

// RateLimitMiddleware 返回一个基于 IP 限制请求的 Gin 中间件（兼容旧版本）
// limit: 每秒请求数 (QPS)
// burst: 最大突发大小
func RateLimitMiddleware(limit rate.Limit, burst int) gin.HandlerFunc {
	rl := NewRateLimiter(limit, burst)
	return rl.Middleware()
}

// RateLimit 是一个便捷函数，返回基于 IP 的速率限制中间件
// qps: 每秒允许的请求数
// burst: 最大突发请求数
func RateLimit(qps float64, burst int) gin.HandlerFunc {
	return RateLimitMiddleware(rate.Limit(qps), burst)
}

// GlobalRateLimiter 全局限流器单例
var (
	globalRateLimiter *RateLimiter
	once              sync.Once
)

// GetGlobalRateLimiter 获取全局限流器单例
// qps: 每秒请求数
// burst: 最大突发请求数
func GetGlobalRateLimiter(qps float64, burst int) *RateLimiter {
	once.Do(func() {
		globalRateLimiter = NewRateLimiter(rate.Limit(qps), burst)
	})
	return globalRateLimiter
}
