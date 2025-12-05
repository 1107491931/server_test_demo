package middleware

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

// responseWriter 自定义响应写入器，用于捕获响应内容
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

// Write 重写Write方法，同时写入原始ResponseWriter和缓冲区
func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

// WriteString 重写WriteString方法
func (w *responseWriter) WriteString(s string) (int, error) {
	w.body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

// generateRequestID 生成唯一的请求ID
func generateRequestID() string {
	b := make([]byte, 16)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// RequestResponseLogger 请求和响应日志中间件
// 记录所有请求的详细信息和响应内容
func RequestResponseLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 开始时间
		startTime := time.Now()

		// 读取请求体
		var requestBody []byte
		if c.Request.Body != nil {
			requestBody, _ = io.ReadAll(c.Request.Body)
			// 重新设置请求体，因为读取后会被消耗
			c.Request.Body = io.NopCloser(bytes.NewBuffer(requestBody))
		}

		// 创建自定义响应写入器
		responseBodyWriter := &responseWriter{
			ResponseWriter: c.Writer,
			body:           bytes.NewBufferString(""),
		}
		c.Writer = responseBodyWriter

		// 生成或获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		// 将请求ID存储到context中，方便后续使用
		c.Set("request_id", requestID)
		// 将请求ID添加到响应头
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 打印请求信息
		printRequest(c, requestBody, requestID)

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 打印响应信息
		printResponse(c, responseBodyWriter.body.Bytes(), duration, requestID)
	}
}

// printRequest 打印请求信息
func printRequest(c *gin.Context, body []byte, requestID string) {
	fmt.Println("\n" + strings.Repeat("=", 80))
	fmt.Printf("📥 [REQUEST] %s %s | Request ID: %s\n", c.Request.Method, c.Request.URL.Path, requestID)
	fmt.Println(strings.Repeat("-", 80))

	// 打印请求头（只打印重要的）
	fmt.Println("Headers:")
	importantHeaders := []string{"Content-Type", "Authorization", "User-Agent", "X-Request-ID"}
	for _, header := range importantHeaders {
		if value := c.GetHeader(header); value != "" {
			fmt.Printf("  %s: %s\n", header, value)
		}
	}

	// 打印查询参数
	if len(c.Request.URL.RawQuery) > 0 {
		fmt.Printf("Query: %s\n", c.Request.URL.RawQuery)
	}

	// 打印请求体（格式化JSON）
	if len(body) > 0 {
		fmt.Println("Body:")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "  ", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Printf("  %s\n", string(body))
		}
	}
}

// printResponse 打印响应信息
func printResponse(c *gin.Context, body []byte, duration time.Duration, requestID string) {
	fmt.Println(strings.Repeat("-", 80))
	fmt.Printf("📤 [RESPONSE] Status: %d | Duration: %v | Request ID: %s\n", c.Writer.Status(), duration, requestID)

	// 打印响应头（只打印重要的）
	fmt.Println("Headers:")
	importantHeaders := []string{"Content-Type", "Content-Length"}
	for _, header := range importantHeaders {
		if value := c.Writer.Header().Get(header); value != "" {
			fmt.Printf("  %s: %s\n", header, value)
		}
	}

	// 打印响应体（格式化JSON）
	if len(body) > 0 && len(body) < 10000 { // 限制大小，避免打印过大的响应
		fmt.Println("Body:")
		var prettyJSON bytes.Buffer
		if err := json.Indent(&prettyJSON, body, "  ", "  "); err == nil {
			fmt.Println(prettyJSON.String())
		} else {
			fmt.Printf("  %s\n", string(body))
		}
	} else if len(body) >= 10000 {
		fmt.Printf("Body: [Large response: %d bytes]\n", len(body))
	}

	fmt.Println(strings.Repeat("=", 80) + "\n")
}

// SimpleRequestLogger 简化版请求日志中间件
// 只记录基本的请求信息，不记录请求体和响应体
func SimpleRequestLogger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 生成或获取请求ID
		requestID := c.GetHeader("X-Request-ID")
		if requestID == "" {
			requestID = generateRequestID()
		}
		c.Set("request_id", requestID)
		c.Writer.Header().Set("X-Request-ID", requestID)

		// 处理请求
		c.Next()

		// 计算耗时
		duration := time.Since(startTime)

		// 打印简单日志
		fmt.Printf("[%s] %s %s | Status: %d | Duration: %v | IP: %s | Request ID: %s\n",
			startTime.Format("2006-01-02 15:04:05"),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			duration,
			c.ClientIP(),
			requestID,
		)
	}
}
