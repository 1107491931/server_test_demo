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
		PrivateKey/PublicKey：
			解释：用于对Token进行RS256非对称加密签名的密钥 (PEM格式).
			PrivateKey: 用于签名 (user-service持有)
			PublicKey: 用于验签 (所有验证Token的服务持有)
	*/
	tokenConfig := &auth.TokenConfig{
		PrivateKey:           GetEnv("JWT_PRIVATE_KEY", ""), // PEM格式的私钥
		PublicKey:            GetEnv("JWT_PUBLIC_KEY", ""),  // PEM格式的公钥
		AccessTokenDuration:  auth.AccessTokenDuration,      // 使用公共常量
		RefreshTokenDuration: auth.RefreshTokenDuration,     // 使用公共常量
		Issuer:               GetEnv("JWT_ISSUER", "we-circle-prod"),
	}

	tokenManager, err := auth.NewTokenManager(tokenConfig, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize TokenManager: %v", err)
	}
	handler.SetTokenManager(tokenManager)
	fmt.Println("TokenManager initialized successfully")

	return redisClient, tokenManager
}
