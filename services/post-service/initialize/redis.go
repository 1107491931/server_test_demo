package initialize

import (
	"common/auth"
	"fmt"
	"log"

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
	// 初始化TokenManager
	// post-service 只需要公钥进行验签
	tokenConfig := &auth.TokenConfig{
		PublicKey:            GetEnv("JWT_PUBLIC_KEY", ""),           // PEM格式的公钥
		AccessTokenDuration:  auth.AccessTokenDuration,               // 使用公共常量
		RefreshTokenDuration: auth.RefreshTokenDuration,              // 使用公共常量
		Issuer:               GetEnv("JWT_ISSUER", "we-circle-prod"), // 注意：Issuer应与user-service保持一致
	}

	tokenManager, err := auth.NewTokenManager(tokenConfig, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize TokenManager: %v", err)
	}
	fmt.Println("TokenManager initialized successfully")

	return redisClient, tokenManager
}
