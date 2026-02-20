package initialize

import (
	"common/auth"
	"common/middleware"
	"post-service/handler"

	"github.com/gin-gonic/gin"
)

// InitRouter 初始化路由
func InitRouter(tokenManager *auth.TokenManager) *gin.Engine {
	// 使用 gin.New() 而不是 Default()
	r := gin.New()

	// 注册核心中间件
	r.Use(middleware.RequestId())            // 生成请求 ID
	r.Use(middleware.PrometheusMiddleware()) // 监控指标
	r.Use(middleware.ZapLogger())            // Zap 日志
	r.Use(middleware.ZapRecovery())          // Zap Recovery
	r.Use(middleware.CORS())                 // CORS 中间件

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

	// 注入请求上下文 Logger
	v1.Use(middleware.ContextLogger())

	{
		posts := v1.Group("/posts")
		// 应用JWT认证 - 所有动态相关接口都需要认证
		posts.Use(middleware.JWTAuth(tokenManager))
		{
			// 创建动态 - 更严格的限流：每秒3个请求，突发5个
			createGroup := posts.Group("")
			createGroup.Use(middleware.RateLimit(3, 5))
			{
				createGroup.POST("/create", handler.CreatePost) // 创建动态
			}

			// 互动接口（点赞、转发、收藏、分享）- 中等限流：每秒5个请求，突发10个
			interactionGroup := posts.Group("")
			interactionGroup.Use(middleware.RateLimit(5, 10))
			{
				interactionGroup.POST("/like", handler.LikePost)         // 点赞动态
				interactionGroup.POST("/dislike", handler.DislikePost)   // 踩动态
				interactionGroup.POST("/favorite", handler.FavoritePost) // 收藏动态
				interactionGroup.POST("/share", handler.SharePost)       // 分享动态
			}

			// 查询接口 - 使用默认的全局限流
			posts.POST("/get_by_id", handler.GetPostByID)               // 获取动态信息
			posts.POST("/get_user_by_post_id", handler.GetUserByPostID) // 根据动态ID获取用户信息
			posts.POST("/get_by_user_id", handler.GetPostsByUserID)     // 根据用户ID获取动态列表
			posts.POST("/get_all", handler.GetAllPosts)                 // 获取所有动态
		}
	}

	return r
}
