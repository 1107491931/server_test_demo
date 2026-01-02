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
	// 使用 gin.New() 而不是 Default()，因为我们要自己注册日志中间件
	r := gin.New()

	// 注册核心中间件, 工作流程见
	r.Use(middleware.RequestId())            // 生成请求 ID
	r.Use(middleware.PrometheusMiddleware()) // 监控指标
	r.Use(middleware.ZapLogger())            // Zap 日志
	r.Use(middleware.ZapRecovery())          // Zap Recovery

	// 健康检查（不限流）
	r.GET("/health", handler.HealthCheck)
	r.GET("/ready", handler.ReadinessCheck)
	r.GET("/live", handler.LivenessCheck)

	// 暴露 Prometheus 指标端点
	r.GET("/metrics", middleware.MetricsHandler())

	// 使用全局限流器：每秒10个请求，突发20个
	globalLimiter := middleware.GetGlobalRateLimiter(10, 20)

	// API路由组
	v1 := r.Group("/api/v1")
	v1.Use(globalLimiter.Middleware()) // 应用全局速率限制

	// 注入请求上下文 Logger (request_id, user_id)
	// 注意：ContextLogger 如果依赖 user_id，理想情况下应该在 JWTAuth 之后，
	// 但如果放在 JWTAuth 之前，则只能获得 request_id。
	// 这里我们先全局应用，以保证所有请求都有 request_id。
	v1.Use(middleware.ContextLogger())

	{
		users := v1.Group("/users")
		{
			// 公开接口 - 不需要认证
			publicGroup := users.Group("")
			publicGroup.Use(middleware.RateLimit(2, 5)) // 更严格的限流
			{
				publicGroup.POST("/register", handler.Register)          // 注册用户
				publicGroup.POST("/login", handler.Login)                // 登录
				publicGroup.POST("/refresh_token", handler.RefreshToken) // 刷新Token
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
