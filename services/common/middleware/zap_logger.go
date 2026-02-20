package middleware

import (
	"bytes"
	"fmt"
	"time"

	"common/logger"

	"github.com/gin-gonic/gin"
)

/*
文件的作用：Zap 日志中间件，用于记录请求和响应日志.
*/

// bodyLogWriter 用于拦截并缓存响应体内容
type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// RequestId 确保每个请求都有一个唯一的 ID，并初始化 Context Logger
func RequestId() gin.HandlerFunc {
	return func(c *gin.Context) {
		rid := c.GetHeader("X-Request-ID") // 从请求头中获取请求ID
		if rid == "" { // 如果请求ID为空，则生成一个唯一的请求ID
			rid = fmt.Sprintf("%d-%d", time.Now().UnixNano(), time.Now().Unix()%1000)
		}
		c.Set("request_id", rid) // 将请求ID存储到上下文中
		c.Header("X-Request-ID", rid) // 将请求ID添加到响应头中

		ctxLogger := logger.GetLogger().With( // 创建一个上下文Logger
			logger.String("request_id", rid),
		)
		c.Set("logger", ctxLogger) // 将上下文Logger存储到上下文中

		c.Next() // 继续处理请求
	}
}

// ZapLogger 返回全量审计版访问日志中间件（包含响应内容）
// 记录请求的详细信息（Method, Path, Status, Latency 等）和响应内容
func ZapLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now() // 记录请求开始时间
		path := c.Request.URL.Path // 记录请求路径
		query := c.Request.URL.RawQuery // 记录请求查询参数

		// 使用自定义 Writer 拦截响应体
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer} // 创建一个自定义响应写入器
		c.Writer = blw // 将自定义响应写入器设置为响应写入器

		c.Next() // 继续处理请求

		latency := time.Since(start)
		l := logger.FromContext(c) // 从上下文中获取Logger

		userID := "" // 用户ID
		if uid, exists := c.Get("user_id"); exists { // 从上下文中获取用户ID	
			userID = fmt.Sprintf("%v", uid) // 将用户ID转换为字符串
		}

		// 限制记录响应体的大小（防止大文件下载导致内存溢出，这里限制 2KB）
		respBody := blw.body.String() // 获取响应体
		if len(respBody) > 2048 {
			respBody = respBody[:2048] + "...(truncated)" // 如果响应体大于2KB，则截断
		}

		fields := []logger.Field{
			logger.Int("status", c.Writer.Status()), // 记录响应状态码
			logger.String("method", c.Request.Method), // 记录请求方法
			logger.String("path", path), // 记录请求路径
			logger.String("latency", latency.String()), // 记录请求延迟
			logger.String("ip", c.ClientIP()), // 记录请求IP
			logger.String("resp", respBody), // 记录响应内容
		}
		// 记录请求查询参数
		if query != "" {
			fields = append(fields, logger.String("query", query))
		}
		// 记录用户ID
		if userID != "" {
			fields = append(fields, logger.String("user_id", userID))
		}

		l.Info(path, fields...) // 记录请求信息
	}
}

// ZapRecovery 返回增强版的恢复中间件，用于记录panic信息，并返回500状态码
/**
 “防止服务因意外错误导致彻底崩溃（Crash）”，并提供 “结构化故障现场记录”。
 在 Go 语言中，如果代码运行到一半发生了 panic（比如：点进了一个空指针、数组下标越界等严重错误），如果不进行拦截，整个微服务进程都会直接退出。 
ZapRecovery相当于给每一个请求套了一个“安全气囊”：当 panic 发生时，它会立刻“捕获”这个错误。它保证了即使这个请求挂了，其他的请求依然能正常处理，服务不会停机。
并给前端返回500状态码，避免前端无法感知到错误。同时记录panic信息，方便后续排查问题。
*/
func ZapRecovery() gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil { // 如果发生panic，则记录panic信息
				l := logger.FromContext(c) // 从上下文中获取Logger
				l.Error("Recovery from panic", logger.Any("error", err), logger.String("request", c.Request.RequestURI)) // 记录panic信息
				c.AbortWithStatus(500) // 返回500状态码
			}
		}()
		c.Next() // 继续处理请求
	}
}

// ContextLogger 已弃用
func ContextLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()
	}
}
