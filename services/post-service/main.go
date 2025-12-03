package main

import (
	"common/middleware" // 速率限制中间件
	"fmt"
	"log"
	"os"
	"post-service/dao"
	"post-service/handler"
	"post-service/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           Post Service API
// @version         1.0
// @description     动态服务API
// @host            localhost:8082
// @BasePath        /
func main() {
	// 1. 读取环境配置
	env := os.Getenv("ENV")
	if env == "" {
		log.Fatal("ENV is not set")
	}

	// 2. 初始化数据库
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		log.Fatal("DB_DSN is not set")
	}

	// 3. 读取服务端口
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Fatal("SERVER_PORT is not set")
	}

	// 打印启动信息
	fmt.Println("========================================")
	fmt.Printf("Service:     Post Service\n")
	fmt.Printf("Environment: %s\n", env)
	fmt.Printf("Database:    %s\n", dsn)
	fmt.Printf("Port:        %s\n", port)
	fmt.Println("========================================")

	// 4. 连接数据库
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// 5. 自动迁移
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("failed to migrate database")
	}

	// 6. 初始化DAO
	dao.InitDB(db)

	// 7. 初始化 Gin 路由
	r := gin.Default()

	// 健康检查（不限流）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// 使用全局限流器：每秒10个请求，突发20个
	globalLimiter := middleware.GetGlobalRateLimiter(10, 20)

	// API路由组 - 应用全局速率限制
	v1 := r.Group("/api/v1")
	v1.Use(globalLimiter.Middleware())
	{
		posts := v1.Group("/posts")
		{
			// 创建动态 - 更严格的限流：每秒3个请求，突发5个
			createGroup := posts.Group("")
			createGroup.Use(middleware.RateLimit(3, 5))
			{
				createGroup.POST("", handler.CreatePost) // 创建动态
			}

			// 互动接口（点赞、转发、收藏）- 中等限流：每秒5个请求，突发10个
			interactionGroup := posts.Group("")
			interactionGroup.Use(middleware.RateLimit(5, 10))
			{
				interactionGroup.POST("/:post_id/like", handler.LikePost)         // 点赞动态
				interactionGroup.POST("/:post_id/forward", handler.ForwardPost)   // 转发动态
				interactionGroup.POST("/:post_id/favorite", handler.FavoritePost) // 收藏动态
			}

			// 查询接口 - 使用默认的全局限流
			posts.GET("/:post_id", handler.GetPostByID)           // 获取动态信息
			posts.GET("/:post_id/user", handler.GetUserByPostID)  // 根据动态ID获取用户信息
			posts.GET("/user/:user_id", handler.GetPostsByUserID) // 根据用户ID获取动态列表
			posts.GET("", handler.GetAllPosts)                    // 获取所有动态
		}
	}

	// 8. 启动服务
	fmt.Printf("Post Service is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("failed to run server")
	}
}
