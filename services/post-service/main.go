package main

import (
	"common/auth"
	"common/logger"
	"common/middleware"
	"context"
	"fmt"
	"post-service/initialize"
	"time"
)

// @title           Post Service API
// @version         1.0
// @description     动态服务API
// @host            localhost:8082
// @BasePath        /
func main() {
	ctx := context.Background()
	// 1. 加载配置
	config := initialize.LoadConfig()

	// 2. 初始化日志
	err := logger.InitStandardLogger("post-service", config.Env, config.LogLevel, config.LokiURL, config.LokiUserID, config.LokiToken)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize logger: %v", err))
	}

	// 统一使用 zap 打印启动信息
	logger.Info("Starting Post Service",
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

	// 5. 启动 Prometheus Metrics Pusher
	middleware.StartMetricsPusher(ctx, middleware.MetricsPusherConfig{
		Enabled:  true,
		URL:      config.PromRemoteURL,
		UserID:   config.PromUserID,
		Token:    config.PromToken,
		Interval: 15 * time.Second,
		Labels:   map[string]string{"service": "post-service", "env": config.Env},
	}, logger.GetLogger())

	// 6. 初始化路由
	r := initialize.InitRouter(tokenManager)

	// 7. 启动服务
	logger.Info("Post Service is running", logger.String("addr", ":"+config.ServerPort))
	if err := r.Run(":" + config.ServerPort); err != nil {
		logger.Fatal("failed to run server", logger.Err(err))
	}
}
