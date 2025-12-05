package auth

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// TokenConfig JWT配置
type TokenConfig struct {
	SecretKey            string        // JWT密钥
	AccessTokenDuration  time.Duration // AccessToken有效期
	RefreshTokenDuration time.Duration // RefreshToken有效期
	Issuer               string        // 签发者
}

// Claims JWT声明
type Claims struct {
	UserID   uint   `json:"userId"`
	Username string `json:"username"`
	Email    string `json:"email"`
	jwt.RegisteredClaims
}

// TokenPair Token对
type TokenPair struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken"`
	ExpiresIn    int64  `json:"expiresIn"` // AccessToken过期时间（秒）
	RefreshIn    int64  `json:"refreshIn"` // RefreshToken过期时间（秒）
}

// TokenManager Token管理器
type TokenManager struct {
	config      *TokenConfig
	redisClient *redis.Client
}

// NewTokenManager 创建Token管理器
func NewTokenManager(config *TokenConfig, redisClient *redis.Client) *TokenManager {
	return &TokenManager{
		config:      config,
		redisClient: redisClient,
	}
}

// GenerateTokenPair 生成Token对（AccessToken和RefreshToken）
func (tm *TokenManager) GenerateTokenPair(userID uint, username, email string) (*TokenPair, error) {
	now := time.Now()

	// 生成AccessToken
	accessClaims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.config.AccessTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    tm.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, accessClaims)
	accessTokenString, err := accessToken.SignedString([]byte(tm.config.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign access token: %w", err)
	}

	// 生成RefreshToken
	refreshClaims := &Claims{
		UserID:   userID,
		Username: username,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(tm.config.RefreshTokenDuration)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    tm.config.Issuer,
			Subject:   fmt.Sprintf("%d", userID),
		},
	}

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString([]byte(tm.config.SecretKey))
	if err != nil {
		return nil, fmt.Errorf("failed to sign refresh token: %w", err)
	}

	return &TokenPair{
		AccessToken:  accessTokenString,
		RefreshToken: refreshTokenString,
		ExpiresIn:    int64(tm.config.AccessTokenDuration.Seconds()),
		RefreshIn:    int64(tm.config.RefreshTokenDuration.Seconds()),
	}, nil
}

// ValidateToken 验证Token
func (tm *TokenManager) ValidateToken(tokenString string) (*Claims, error) {
	// 检查Token是否在黑名单中
	ctx := context.Background()
	isBlacklisted, err := tm.IsTokenBlacklisted(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	// 解析Token
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(tm.config.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 提取Claims
	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

// RefreshAccessToken 使用RefreshToken刷新AccessToken
func (tm *TokenManager) RefreshAccessToken(refreshTokenString string) (*TokenPair, error) {
	// 验证RefreshToken
	claims, err := tm.ValidateToken(refreshTokenString)
	if err != nil {
		return nil, fmt.Errorf("invalid refresh token: %w", err)
	}

	// 生成新的Token对
	return tm.GenerateTokenPair(claims.UserID, claims.Username, claims.Email)
}

// RevokeToken 撤销Token（加入黑名单）
func (tm *TokenManager) RevokeToken(ctx context.Context, tokenString string) error {
	// 解析Token获取过期时间
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return []byte(tm.config.SecretKey), nil
	})

	if err != nil {
		return fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return errors.New("invalid token claims")
	}

	// 计算Token剩余有效时间
	ttl := time.Until(claims.ExpiresAt.Time)
	if ttl <= 0 {
		// Token已过期，无需加入黑名单
		return nil
	}

	// 将Token加入Redis黑名单
	key := fmt.Sprintf("blacklist:token:%s", tokenString)
	err = tm.redisClient.Set(ctx, key, "1", ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsTokenBlacklisted 检查Token是否在黑名单中
func (tm *TokenManager) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	key := fmt.Sprintf("blacklist:token:%s", tokenString)
	exists, err := tm.redisClient.Exists(ctx, key).Result()
	if err != nil {
		return false, err
	}
	return exists > 0, nil
}

// Logout 用户登出（撤销AccessToken和RefreshToken）
func (tm *TokenManager) Logout(ctx context.Context, accessToken, refreshToken string) error {
	// 撤销AccessToken
	if err := tm.RevokeToken(ctx, accessToken); err != nil {
		return fmt.Errorf("failed to revoke access token: %w", err)
	}

	// 撤销RefreshToken
	if err := tm.RevokeToken(ctx, refreshToken); err != nil {
		return fmt.Errorf("failed to revoke refresh token: %w", err)
	}

	return nil
}
