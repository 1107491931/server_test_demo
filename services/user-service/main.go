package main

import (
	"fmt"
	"log"
	"os"
	"user-service/dao"
	"user-service/handler"
	"user-service/model"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// @title           User Service API
// @version         1.0
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

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由组
	v1 := r.Group("/api/v1")
	{
		users := v1.Group("/users")
		{
			users.POST("/register", handler.Register) // 注册用户
			users.POST("/login", handler.Login) // 登录
			users.GET("/:user_id", handler.GetUserByID) // 获取用户信息
			users.GET("/:user_id/posts", handler.GetUserWithPosts) // 根据用户ID获取用户信息及其所有动态
			users.GET("/phone/:phone", handler.GetUserByPhone) // 根据手机号获取用户信息
			users.POST("/batch", handler.BatchGetUsers) // 批量获取用户信息
			users.GET("", handler.GetAllUsers) // 获取所有用户信息
		}
	}

	// 8. 启动服务
	fmt.Printf("User Service is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("failed to run server")
	}
}
