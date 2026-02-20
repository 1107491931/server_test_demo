package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// HealthCheck 处理健康检查请求
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "user-service",
		"version": "1.0.0",
	})
}

// ReadinessCheck 处理就绪检查请求
func ReadinessCheck(c *gin.Context) {
	// 检查TokenManager是否初始化
	if tokenManager == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"status":  "not ready",
			"message": "TokenManager not initialized",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "user-service",
	})
}

// LivenessCheck 处理存活检查请求
func LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"service": "user-service",
	})
}
