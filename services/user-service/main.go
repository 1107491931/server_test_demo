package main

import (
	"common/middleware" // 速率限制中间件
	"fmt"
	"log"
	"os"
	"user-service/dao"
	_ "user-service/docs" // Swagger 文档
	"user-service/handler"
	"user-service/model"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           User Service API
// @version         2.0
// @description     用户服务API
// @host            localhost:8081
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

	// 3. 读取服务端口8081
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		log.Fatal("SERVER_PORT is not set")
	}

	// 打印启动信息
	fmt.Println("========================================")
	fmt.Printf("Service:     User Service\n")
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
		users := v1.Group("/users")
		{
			// 注册和登录接口 - 更严格的限流：每秒2个请求，突发5个
			authGroup := users.Group("")
			authGroup.Use(middleware.RateLimit(2, 5))
			{
				authGroup.POST("/register", handler.Register) // 注册用户
				authGroup.POST("/login", handler.Login)       // 登录
			}

			// 查询接口 - 使用默认的全局限流
			users.GET("/:user_id", handler.GetUserByID)            // 获取用户信息
			users.GET("/:user_id/posts", handler.GetUserWithPosts) // 根据用户ID获取用户信息及其所有动态
			users.GET("/phone/:phone", handler.GetUserByPhone)     // 根据手机号获取用户信息
			users.POST("/batch", handler.BatchGetUsers)            // 批量获取用户信息
			users.GET("", handler.GetAllUsers)                     // 获取所有用户信息
		}
	}

	// Swagger 文档, 需要在启动服务之前
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 8. 启动服务
	fmt.Printf("User Service is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("failed to run server")
	}
}
