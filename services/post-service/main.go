package main

import (
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

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API路由组
	v1 := r.Group("/api/v1")
	{
		posts := v1.Group("/posts")
		{
			posts.POST("", handler.CreatePost)
			posts.GET("/:post_id", handler.GetPostByID)
			posts.GET("/user/:user_id", handler.GetPostsByUserID)
			posts.GET("", handler.GetAllPosts)
			posts.POST("/:post_id/like", handler.LikePost)
			posts.POST("/:post_id/forward", handler.ForwardPost)
			posts.POST("/:post_id/favorite", handler.FavoritePost)
		}
	}

	// 8. 启动服务
	fmt.Printf("Post Service is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("failed to run server")
	}
}
