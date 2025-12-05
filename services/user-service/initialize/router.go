package initialize

import (
	"common/auth"
	"common/middleware"
	"user-service/handler"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// InitRouter 初始化路由
func InitRouter(tokenManager *auth.TokenManager) *gin.Engine {
	r := gin.Default()

	// 健康检查（不限流）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 使用全局限流器：每秒10个请求，突发20个
	globalLimiter := middleware.GetGlobalRateLimiter(10, 20)

	// API路由组 - 应用全局速率限制和请求/响应日志
	v1 := r.Group("/api/v1")
	v1.Use(globalLimiter.Middleware())         // 应用全局速率限制
	v1.Use(middleware.RequestResponseLogger()) // 添加请求/响应日志中间件
	{
		users := v1.Group("/users")
		{
			// 公开接口 - 不需要认证
			publicGroup := users.Group("")
			publicGroup.Use(middleware.RateLimit(2, 5)) // 更严格的限流
			{
				publicGroup.POST("/register", handler.Register)    // 注册用户
				publicGroup.POST("/login", handler.Login)          // 登录
				publicGroup.POST("/refresh", handler.RefreshToken) // 刷新Token
			}

			// 需要认证的接口
			authGroup := users.Group("")
			authGroup.Use(middleware.JWTAuth(tokenManager)) // 应用JWT认证
			{
				authGroup.POST("/logout", handler.Logout)                   // 登出
				authGroup.POST("/get_by_id", handler.GetUserByID)           // 获取用户信息
				authGroup.POST("/get_with_posts", handler.GetUserWithPosts) // 根据用户ID获取用户信息及其所有动态
				authGroup.POST("/get_by_email", handler.GetUserByEmail)     // 根据邮箱获取用户信息
				authGroup.POST("/batch", handler.BatchGetUsers)             // 批量获取用户信息
				authGroup.POST("/get_all", handler.GetAllUsers)             // 获取所有用户信息
			}
		}
	}

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	return r
}
