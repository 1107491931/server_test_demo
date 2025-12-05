package main

import (
	"common/auth"
	"fmt"
	"log"
	"post-service/initialize"
)

// @title           Post Service API
// @version         1.0
// @description     动态服务API
// @host            localhost:8082
// @BasePath        /
func main() {
	// 1. 加载配置
	config := initialize.LoadConfig()

	// 打印启动信息
	fmt.Println("========================================")
	fmt.Printf("Service:     Post Service\n")
	fmt.Printf("Environment: %s\n", config.Env)
	fmt.Printf("Database:    %s\n", config.DBDSN)
	fmt.Printf("Port:        %s\n", config.ServerPort)
	fmt.Println("========================================")

	// 2. 初始化数据库
	initialize.InitDB(config.DBDSN)

	// 3. 初始化Redis和认证模块
	redisClient, tokenManager := initialize.InitRedisAndAuth()
	if redisClient != nil {
		defer auth.CloseRedisClient(redisClient)
	}

	// 4. 初始化路由
	r := initialize.InitRouter(tokenManager)

	// 5. 启动服务
	fmt.Printf("Post Service is running on :%s\n", config.ServerPort)
	if err := r.Run(":" + config.ServerPort); err != nil {
		log.Fatal("failed to run server")
	}
}
