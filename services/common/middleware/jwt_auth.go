package middleware

import (
	"common/auth"
	"common/errs"
	"common/utils"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// JWTAuth JWT认证中间件
func JWTAuth(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求头获取Token
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			utils.Error(c, http.StatusOK, errs.UNAUTHORIZED, "Missing authorization header")
			c.Abort()
			return
		}

		// 检查Bearer前缀
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			utils.Error(c, http.StatusOK, errs.UNAUTHORIZED, "Invalid authorization header format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// 验证Token
		claims, err := tokenManager.ValidateToken(tokenString)
		if err != nil {
			utils.Error(c, http.StatusOK, errs.TOKEN_REVOKED, "Invalid or expired token")
			c.Abort()
			return
		}

		// 将用户信息存储到Context中
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("email", claims.Email)

		c.Next()
	}
}

// OptionalJWTAuth 可选的JWT认证中间件（Token可以不存在）
func OptionalJWTAuth(tokenManager *auth.TokenManager) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			// 没有Token，继续处理请求
			c.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			// Token格式错误，继续处理请求
			c.Next()
			return
		}

		tokenString := parts[1]
		claims, err := tokenManager.ValidateToken(tokenString)
		if err == nil {
			// Token有效，存储用户信息
			c.Set("user_id", claims.UserID)
			c.Set("username", claims.Username)
			c.Set("email", claims.Email)
		}

		c.Next()
	}
}

// GetUserID 从Context获取用户ID
func GetUserID(c *gin.Context) (uint, bool) {
	userID, exists := c.Get("user_id")
	if !exists {
		return 0, false
	}
	id, ok := userID.(uint)
	return id, ok
}

// GetUsername 从Context获取用户名
func GetUsername(c *gin.Context) (string, bool) {
	username, exists := c.Get("username")
	if !exists {
		return "", false
	}
	name, ok := username.(string)
	return name, ok
}

// GetEmail 从Context获取邮箱
func GetEmail(c *gin.Context) (string, bool) {
	email, exists := c.Get("email")
	if !exists {
		return "", false
	}
	e, ok := email.(string)
	return e, ok
}
