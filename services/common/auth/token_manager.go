package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/redis/go-redis/v9"
)

// Token expiration constants
const (
	AccessTokenDuration = 24 * time.Hour // AccessToken有效期：24小时
	// AccessTokenDuration  = 30 * time.Second    // AccessToken有效期：30秒. 用于测试
	RefreshTokenDuration = 15 * 24 * time.Hour // RefreshToken有效期：15天
)

// TokenConfig JWT配置
type TokenConfig struct {
	PrivateKey           string        // RSA私钥 (PEM格式)
	PublicKey            string        // RSA公钥 (PEM格式)
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
	privateKey  *rsa.PrivateKey
	publicKey   *rsa.PublicKey
}

// NewTokenManager 创建Token管理器
func NewTokenManager(config *TokenConfig, redisClient *redis.Client) (*TokenManager, error) {
	tm := &TokenManager{
		config:      config,
		redisClient: redisClient,
	}

	// 辅助函数：解析密钥内容（支持 Base64 和转义换行符）
	parseKeyContent := func(keyStr string) []byte {
		// 如果不包含 PEM 头，尝试 Base64 解码
		if !strings.Contains(keyStr, "-----BEGIN") {
			decoded, err := base64.StdEncoding.DecodeString(keyStr)
			if err == nil {
				return decoded
			}
		}
		// 处理转义换行符
		return []byte(strings.ReplaceAll(keyStr, "\\n", "\n"))
	}

	// 解析私钥 (如果提供)
	if config.PrivateKey != "" {
		privateKey, err := jwt.ParseRSAPrivateKeyFromPEM(parseKeyContent(config.PrivateKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		tm.privateKey = privateKey
	}

	// 解析公钥
	if config.PublicKey != "" {
		publicKey, err := jwt.ParseRSAPublicKeyFromPEM(parseKeyContent(config.PublicKey))
		if err != nil {
			return nil, fmt.Errorf("failed to parse public key: %w", err)
		}
		tm.publicKey = publicKey
	}

	return tm, nil
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
			IssuedAt:  jwt.NewNumericDate(now),   // 签名时间
			NotBefore: jwt.NewNumericDate(now),   // 生效时间
			Issuer:    tm.config.Issuer,          // 标识Token的签发者, 用于验证Token的来源是否可信
			Subject:   fmt.Sprintf("%d", userID), // 标识Token的主题（这个Token是为谁创建的）,通常是用户ID、邮箱或其他唯一标识
		},
	}

	// 使用RS256签名算法
	accessToken := jwt.NewWithClaims(jwt.SigningMethodRS256, accessClaims)
	if tm.privateKey == nil {
		return nil, errors.New("private key is not configured for signing")
	}
	accessTokenString, err := accessToken.SignedString(tm.privateKey)
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

	refreshToken := jwt.NewWithClaims(jwt.SigningMethodRS256, refreshClaims)
	refreshTokenString, err := refreshToken.SignedString(tm.privateKey)
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
	// 1. 首先解析和验证Token

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodRSA); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		if tm.publicKey == nil {
			return nil, errors.New("public key is not configured for verification")
		}
		return tm.publicKey, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	// 2. 验证Token有效性
	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	// 3. 提取Claims并验证
	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	// 4. 验证标准声明
	now := time.Now()

	// 验证过期时间（token.Valid已经检查过，这里再显式检查）
	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(now) {
		return nil, errors.New("token expired")
	}

	// 验证生效时间（如果有设置）
	if claims.NotBefore != nil && claims.NotBefore.Time.After(now) {
		return nil, errors.New("token not yet valid")
	}

	// 验证签发者
	if claims.Issuer != tm.config.Issuer {
		return nil, errors.New("invalid token issuer")
	}

	// 验证主题（如果业务需要）
	if claims.Subject == "" {
		return nil, errors.New("token subject is empty")
	}

	// 解析用户ID
	userID, err := strconv.ParseUint(claims.Subject, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("invalid user id in subject: %w", err)
	}

	// 确保自定义的UserID和Subject一致
	if claims.UserID != uint(userID) {
		return nil, errors.New("user id mismatch")
	}

	// 5. 检查Token是否在黑名单中
	ctx := context.Background()
	isBlacklisted, err := tm.IsTokenBlacklisted(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to check token blacklist: %w", err)
	}
	if isBlacklisted {
		return nil, errors.New("token has been revoked")
	}

	return claims, nil
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
	// 如果Redis客户端不可用，返回错误
	if tm.redisClient == nil {
		return errors.New("redis client not available")
	}

	// 解析Token获取过期时间
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if tm.publicKey == nil {
			return nil, errors.New("public key is not configured for verification")
		}
		return tm.publicKey, nil
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
	err = tm.redisClient.Set(ctx, key, "1", ttl).Err() // 过期后会删除
	if err != nil {
		return fmt.Errorf("failed to add token to blacklist: %w", err)
	}

	return nil
}

// IsTokenBlacklisted 检查Token是否在黑名单中
func (tm *TokenManager) IsTokenBlacklisted(ctx context.Context, tokenString string) (bool, error) {
	// 如果Redis客户端不可用，返回错误
	if tm.redisClient == nil {
		return false, errors.New("redis client not available")
	}

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
	if accessToken != "" {
		if err := tm.RevokeToken(ctx, accessToken); err != nil {
			return fmt.Errorf("failed to revoke access token: %w", err)
		}
	}

	// 撤销RefreshToken
	if refreshToken != "" {
		if err := tm.RevokeToken(ctx, refreshToken); err != nil {
			return fmt.Errorf("failed to revoke refresh token: %w", err)
		}
	}

	return nil
}
