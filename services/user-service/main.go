package main

import (
	"common/auth"
	"common/logger"
	"common/middleware"
	"context"
	"fmt"
	"time"
	_ "user-service/docs" // Swagger 文档
	"user-service/initialize"
)

// @title           User Service API
// @version         2.0
// @description     用户服务API
// @host            localhost:8081
// @BasePath        /
func main() {
	// 1. 加载配置
	config := initialize.LoadConfig()

	// 2. 初始化日志
	err := logger.InitStandardLogger("user-service", config.Env, config.LogLevel, config.LokiURL, config.LokiUserID, config.LokiToken)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	// 统一使用 zap 打印启动信息
	logger.Info("Starting User Service",
		logger.String("env", config.Env),
		logger.String("port", config.ServerPort),
		logger.String("db", config.DBDSN),
	)

	// 3. 初始化数据库
	initialize.InitDB(config.DBDSN)

	// 4. 初始化Redis和认证模块
	redisClient, tokenManager := initialize.InitRedisAndAuth(config)
	if redisClient != nil {
		defer auth.CloseRedisClient(redisClient)
	}

	// 5. 初始化路由
	r := initialize.InitRouter(tokenManager)

	// 5.5 启动指标推送器 (每15秒主动推送一次指标到 Grafana Cloud)
	middleware.StartMetricsPusher(context.Background(), middleware.MetricsPusherConfig{
		Enabled:  config.PromRemoteURL != "",
		URL:      config.PromRemoteURL,
		UserID:   config.PromUserID,
		Token:    config.PromToken,
		Interval: 15 * time.Second,
		Labels:   map[string]string{"service": "user-service", "env": config.Env},
	}, logger.GetLogger())

	// 6. 启动服务
	logger.Info("User Service is running", logger.String("addr", ":"+config.ServerPort))
	if err := r.Run(":" + config.ServerPort); err != nil {
		logger.Fatal("failed to run server", logger.Err(err))
	}
}
