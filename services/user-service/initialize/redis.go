package initialize

import (
	"common/auth"
	"fmt"
	"log"
	"user-service/handler"

	"github.com/redis/go-redis/v9"
)

// InitRedisAndAuth 初始化Redis和认证模块
func InitRedisAndAuth() (*redis.Client, *auth.TokenManager) {
	// 初始化Redis
	redisConfig := &auth.RedisConfig{
		Host:     GetEnv("REDIS_HOST", "localhost"),
		Port:     GetEnvAsInt("REDIS_PORT", 6379),
		Password: GetEnv("REDIS_PASSWORD", ""),
		DB:       GetEnvAsInt("REDIS_DB", 0),
	}

	redisClient, err := auth.NewRedisClient(redisConfig)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Token blacklist feature will be disabled")
	} else {
		fmt.Println("Redis connected successfully")
	}

	// 初始化TokenManager
	/*
	SecretKey：
		解释：用于对Token进行签名和验证的密钥. 必须保密. 不能在代码中写死. 因此通过环境变量传入.
		长度: token加密算法不同，建议的长度不同，当然具体长度倒是没限制。
			SigningMethodHS256: 建议32字符
			SigningMethodHS384: 建议48字符
			SigningMethodHS512: 建议64字符
	*/
	// Issuer：标识Token的签发者, 用于验证Token的来源是否可信. 验证token有效性时，也会校验这个字段.
	tokenConfig := &auth.TokenConfig{
		SecretKey:            GetEnv("JWT_SECRET_KEY", "default-secret-key-change-in-production"),
		AccessTokenDuration:  auth.AccessTokenDuration, // 使用公共常量
		RefreshTokenDuration: auth.RefreshTokenDuration, // 使用公共常量
		Issuer:               GetEnv("JWT_ISSUER", "user-service"),
	}

	tokenManager := auth.NewTokenManager(tokenConfig, redisClient)
	handler.SetTokenManager(tokenManager)
	fmt.Println("TokenManager initialized successfully")

	return redisClient, tokenManager
}
