package initialize

import (
	"common/auth"
	"common/middleware"
	"post-service/handler"

	"github.com/gin-gonic/gin"
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
