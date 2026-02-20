package initialize

import (
	"common/auth"
	"fmt"
	"log"

	"github.com/redis/go-redis/v9"
)

// InitRedisAndAuth 初始化Redis和认证模块
func InitRedisAndAuth(cfg *Config) (*redis.Client, *auth.TokenManager) {
	// 初始化Redis
	redisConfig := &auth.RedisConfig{
		Host:     cfg.RedisHost,
		Port:     cfg.RedisPort,
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	}

	redisClient, err := auth.NewRedisClient(redisConfig)
	if err != nil {
		log.Printf("Warning: Failed to connect to Redis: %v", err)
		log.Println("Token blacklist feature will be disabled")
	} else {
		fmt.Println("Redis connected successfully")
	}

	// 初始化TokenManager
	// post-service 重点是验签
	tokenConfig := &auth.TokenConfig{
		PublicKey:            cfg.JWTPublicKey,
		AccessTokenDuration:  auth.AccessTokenDuration,
		RefreshTokenDuration: auth.RefreshTokenDuration,
		Issuer:               cfg.JWTIssuer,
	}

	tokenManager, err := auth.NewTokenManager(tokenConfig, redisClient)
	if err != nil {
		log.Fatalf("Failed to initialize TokenManager: %v", err)
	}
	fmt.Println("TokenManager initialized successfully")

	return redisClient, tokenManager
}
