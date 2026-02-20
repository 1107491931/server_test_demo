package handler

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

// HealthCheck 处理健康检查请求
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "healthy",
		"service": "post-service",
		"version": "1.0.0",
	})
}

// ReadinessCheck 处理就绪检查请求
func ReadinessCheck(c *gin.Context) {
	// 这里可以添加其他就绪检查逻辑
	c.JSON(http.StatusOK, gin.H{
		"status":  "ready",
		"service": "post-service",
	})
}

// LivenessCheck 处理存活检查请求
func LivenessCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status":  "alive",
		"service": "post-service",
	})
}
