package main

import (
	"fmt"
	"log"
	"server_test_demo/db_handler"
	"server_test_demo/handler"
	"server_test_demo/model"

	"os"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	_ "server_test_demo/docs" // 导入生成的 docs 包
)

// @title           Test Demo API
// @version         1.0
// @description     This is a sample server for Test Demo.
// @host            localhost:8081
// @BasePath        /
func main() {
	/*
		环境变量可以通过两种设置, 从而在main.go中通过os.Getenv()获取
		1. 运行Docker镜像时：
			docker run -d \
					--name test_demo_staging \
					-p 8082:8081 \
					-v $(pwd)/dbs/staging:/app/dbs \
					--env-file config/.env.staging \
					test_demo_1.0.0
		2. 终端运行项目时
			ENV=staging SERVER_PORT=8082 DB_DSN=dbs/staging/test_staging.db go run main.go
	*/
	// 1. 读取环境配置
	env := os.Getenv("ENV")
	if env == "" {
		// 直接报错并退出
		log.Fatal("ENV is not set")
	}

	// 2. 初始化数据库，优先从环境变量获取数据库路径
	dsn := os.Getenv("DB_DSN")
	if dsn == "" {
		// 直接报错并退出
		log.Fatal("DB_DSN is not set")
	}

	// 3. 读取服务端口
	// 如果环境变量未设置 SERVER_PORT，则默认使用 8081
	// 在 Docker 部署中，通常不需要设置此变量，容器内部统一使用 8081
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		// 直接报错并退出
		log.Fatal("SERVER_PORT is not set")
		// port = "8081"
	}

	// 打印启动信息
	fmt.Println("========================================")
	fmt.Printf("Environment: %s\n", env)
	fmt.Printf("Database:    %s\n", dsn)
	fmt.Printf("Port:        %s\n", port)
	fmt.Println("========================================")

	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatal("failed to connect database")
	}

	// 4. 自动迁移
	if err := model.AutoMigrate(db); err != nil {
		log.Fatal("failed to migrate database")
	}

	// 5. 初始化工具类数据库连接
	db_handler.InitDB(db)

	// 6. 初始化 Gin 路由
	r := gin.Default()

	// Swagger 文档
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// 7. 注册路由
	// 登录注册
	r.POST("/register", handler.Register)
	r.POST("/login", handler.Login)

	// 用户信息
	r.GET("/users/:phone", handler.GetUserByPhone)
	r.GET("/users", handler.GetAllUsers)

	// 8. 启动服务
	fmt.Printf("Server is running on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("failed to run server")
	}
}
