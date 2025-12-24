# JWT + Redis 认证系统集成指南

## 概述

项目已集成JWT认证和Redis黑名单机制，所有依赖统一在`common`模块中管理。

## 功能特性

### 1. Token管理
- ✅ **AccessToken**: 有效期24小时
- ✅ **RefreshToken**: 有效期15天
- ✅ **Token刷新**: 使用RefreshToken获取新的AccessToken
- ✅ **Token撤销**: 退出登录时将Token加入Redis黑名单
- ✅ **黑名单检查**: 每次验证Token时检查是否已被撤销

### 2. 依赖管理
所有三方库版本统一在`services/common/go.mod`中定义：
- `github.com/golang-jwt/jwt/v5 v5.3.0` - JWT库
- `github.com/redis/go-redis/v9 v9.17.2` - Redis客户端

## 核心组件

### 1. Token管理器 (`common/auth/token_manager.go`)

**主要功能：**
- `GenerateTokenPair()` - 生成AccessToken和RefreshToken
- `ValidateToken()` - 验证Token有效性
- `RefreshAccessToken()` - 刷新AccessToken
- `RevokeToken()` - 撤销单个Token
- `Logout()` - 用户登出（撤销所有Token）

### 2. Redis客户端 (`common/auth/redis_client.go`)

**主要功能：**
- `NewRedisClient()` - 创建Redis连接
- `CloseRedisClient()` - 关闭Redis连接

### 3. JWT认证中间件 (`common/middleware/jwt_auth.go`)

**主要功能：**
- `JWTAuth()` - 必需认证中间件
- `OptionalJWTAuth()` - 可选认证中间件
- `GetUserID()` - 获取当前用户ID
- `GetUsername()` - 获取当前用户名
- `GetEmail()` - 获取当前用户邮箱

## 环境变量配置

需要在启动服务时设置以下环境变量：

```bash
# 基础配置
ENV=staging
SERVER_PORT=8081
DB_DSN=dbs/staging/user_staging.db

# JWT配置
# JWT配置 (RS256 非对称加密)
JWT_PRIVATE_KEY="$(cat ../../private.pem)"   # 仅 user-service 需要
JWT_PUBLIC_KEY="$(cat ../../public.pem)"     # 所有服务都需要
JWT_ISSUER=we-circle-prod

# Redis配置
REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```
非对称加密文件生成命令，注意：本地开发时有各位研发本地自己生成，prod环境由运维生成，禁止公开。

生成私钥：
```
openssl genpkey -algorithm RSA -out private.pem -pkeyopt rsa_keygen_bits:2048
```
- genpkey：这是OpenSSL中生成通用私钥的现代命令，支持多种算法。
- algorithm RSA：明确指定算法为RSA。
- pkeyopt rsa_keygen_bits:2048：设置密钥长度为2048位。
- 结果：生成一个 PKCS#8 格式的私钥文件 private.pem。

通过私钥导出公钥：
```
openssl rsa -pubout -in private.pem -out public.pem
```
- rsa：这是OpenSSL中处理RSA密钥的命令。
- -pubout：指定输出公钥。
- -in private.pem：指定输入的私钥文件。
- -out public.pem：指定输出的公钥文件。

## 在服务中集成

### 步骤1: 初始化Redis和TokenManager

在`main.go`中添加：

```go
package main

import (
    "common/auth"
    "common/middleware"
    "log"
    "os"
    "strconv"
    "time"
    
    "github.com/gin-gonic/gin"
)

func main() {
    // ... 其他初始化代码 ...
    
    // 初始化Redis
    redisConfig := &auth.RedisConfig{
        Host:     getEnv("REDIS_HOST", "localhost"),
        Port:     getEnvAsInt("REDIS_PORT", 6379),
        Password: getEnv("REDIS_PASSWORD", ""),
        DB:       getEnvAsInt("REDIS_DB", 0),
    }
    
    redisClient, err := auth.NewRedisClient(redisConfig)
    if err != nil {
        log.Fatalf("Failed to connect to Redis: %v", err)
    }
    defer auth.CloseRedisClient(redisClient)
    
    // 初始化TokenManager
    tokenConfig := &auth.TokenConfig{
        PrivateKey:           getEnv("JWT_PRIVATE_KEY", ""), // PEM格式字符串
        PublicKey:            getEnv("JWT_PUBLIC_KEY", ""),  // PEM格式字符串
        AccessTokenDuration:  24 * time.Hour,
        RefreshTokenDuration: 15 * 24 * time.Hour,
        Issuer:               getEnv("JWT_ISSUER", "user-service"),
    }
    
    tokenManager := auth.NewTokenManager(tokenConfig, redisClient)
    
    // 将tokenManager传递给handler
    handler.SetTokenManager(tokenManager)
    
    // ... 路由配置 ...
}

// 辅助函数
func getEnv(key, defaultValue string) string {
    value := os.Getenv(key)
    if value == "" {
        return defaultValue
    }
    return value
}

func getEnvAsInt(key string, defaultValue int) int {
    valueStr := os.Getenv(key)
    if valueStr == "" {
        return defaultValue
    }
    value, err := strconv.Atoi(valueStr)
    if err != nil {
        return defaultValue
    }
    return value
}
```

### 步骤2: 应用JWT认证中间件

```go
// 需要认证的路由
authenticated := v1.Group("/users")
authenticated.Use(middleware.JWTAuth(tokenManager))
{
    authenticated.POST("/profile", handler.GetProfile)
    authenticated.POST("/update", handler.UpdateProfile)
    authenticated.POST("/logout", handler.Logout)
}

// 不需要认证的路由
public := v1.Group("/users")
{
    public.POST("/register", handler.Register)
    public.POST("/login", handler.Login)
    public.POST("/refresh", handler.RefreshToken)
}
```

### 步骤3: 修改Handler

#### 注册/登录Handler

```go
package handler

import (
    "common/auth"
    "github.com/gin-gonic/gin"
)

var tokenManager *auth.TokenManager

// SetTokenManager 设置TokenManager
func SetTokenManager(tm *auth.TokenManager) {
    tokenManager = tm
}

// Register 用户注册
func Register(c *gin.Context) {
    var req RegisterRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err.Error())
        return
    }
    
    // ... 创建用户逻辑 ...
    
    // 生成Token
    tokenPair, err := tokenManager.GenerateTokenPair(user.ID, user.Username, user.Email)
    if err != nil {
        utils.InternalServerError(c, "Failed to generate token")
        return
    }
    
    utils.SuccessWithMessage(c, "注册成功", gin.H{
        "user":         toUserResponse(&user),
        "accessToken":  tokenPair.AccessToken,
        "refreshToken": tokenPair.RefreshToken,
        "expiresIn":    tokenPair.ExpiresIn,
        "refreshIn":    tokenPair.RefreshIn,
    })
}

// Login 用户登录
func Login(c *gin.Context) {
    var req LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err.Error())
        return
    }
    
    // ... 验证用户逻辑 ...
    
    // 生成Token
    tokenPair, err := tokenManager.GenerateTokenPair(user.ID, user.Username, user.Email)
    if err != nil {
        utils.InternalServerError(c, "Failed to generate token")
        return
    }
    
    utils.SuccessWithMessage(c, "登录成功", gin.H{
        "user":         toUserResponse(user),
        "accessToken":  tokenPair.AccessToken,
        "refreshToken": tokenPair.RefreshToken,
        "expiresIn":    tokenPair.ExpiresIn,
        "refreshIn":    tokenPair.RefreshIn,
    })
}
```

#### 刷新Token Handler

```go
// RefreshTokenRequest 刷新Token请求
type RefreshTokenRequest struct {
    RefreshToken string `json:"refreshToken" binding:"required"`
}

// RefreshToken 刷新AccessToken
func RefreshToken(c *gin.Context) {
    var req RefreshTokenRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err.Error())
        return
    }
    
    // 使用RefreshToken生成新的Token对
    tokenPair, err := tokenManager.RefreshAccessToken(req.RefreshToken)
    if err != nil {
        utils.Unauthorized(c, "Invalid or expired refresh token")
        return
    }
    
    utils.SuccessWithMessage(c, "Token刷新成功", tokenPair)
}
```

#### 登出Handler

```go
// LogoutRequest 登出请求
type LogoutRequest struct {
    AccessToken  string `json:"accessToken" binding:"required"`
    RefreshToken string `json:"refreshToken" binding:"required"`
}

// Logout 用户登出
func Logout(c *gin.Context) {
    var req LogoutRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        utils.BadRequest(c, err.Error())
        return
    }
    
    // 撤销Token
    ctx := c.Request.Context()
    if err := tokenManager.Logout(ctx, req.AccessToken, req.RefreshToken); err != nil {
        utils.InternalServerError(c, "Failed to logout")
        return
    }
    
    utils.SuccessWithMessage(c, "登出成功", nil)
}
```

#### 需要认证的Handler

```go
// GetProfile 获取用户资料
func GetProfile(c *gin.Context) {
    // 从Context获取用户信息
    userID, exists := middleware.GetUserID(c)
    if !exists {
        utils.Unauthorized(c, "User not authenticated")
        return
    }
    
    // 使用userID查询用户信息
    user, err := dao.GetUserByID(userID)
    if err != nil {
        utils.NotFound(c, "User not found")
        return
    }
    
    utils.Success(c, toUserResponse(user))
}
```

## API使用示例

### 1. 注册

```bash
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "张三",
    "email": "zhangsan@example.com",
    "password": "123456"
  }'
```

**响应：**
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user": {
      "userId": 1,
      "username": "张三",
      "email": "zhangsan@example.com"
    },
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expiresIn": 86400,
    "refreshIn": 1296000
  }
}
```

### 2. 登录

```bash
curl -X POST http://localhost:8081/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "zhangsan@example.com",
    "password": "123456"
  }'
```

### 3. 访问需要认证的接口

```bash
curl -X POST http://localhost:8081/api/v1/users/profile \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
```

### 4. 刷新Token

```bash
curl -X POST http://localhost:8081/api/v1/users/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

### 5. 登出

```bash
curl -X POST http://localhost:8081/api/v1/users/logout \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..." \
  -d '{
    "accessToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refreshToken": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }'
```

## Redis安装

### macOS
```bash
brew install redis
brew services start redis
```

### Docker
```bash
docker run -d --name redis -p 6379:6379 redis:latest
```

## 注意事项

1. **JWT密钥安全 (RS256)**
   - 使用 `openssl genrsa` 生成强 RSA 密钥对
   - **Private Key** 必须严格保密，仅用于签发 Token（user-service）
   - **Public Key** 可分发给验证 Token 的服务（post-service 等）
   - 不要将私钥提交到代码仓库
   - 使用环境变量传入 PEM 格式的密钥内容

2. **Token过期时间**
   - AccessToken: 24小时（可根据需求调整）
   - RefreshToken: 15天（可根据需求调整）

3. **Redis连接**
   - 确保Redis服务正常运行
   - 生产环境建议配置Redis密码
   - 考虑使用Redis集群提高可用性

4. **黑名单清理**
   - Redis会自动清理过期的黑名单记录
   - TTL设置为Token的剩余有效时间

5. **中间件顺序**
   - JWT认证中间件应该在限流中间件之后
   - 日志中间件可以在JWT之前或之后

## 完整启动命令

```bash
# 启动Redis
brew services start redis

# 启动user-service (使用 cat 读取密钥文件)
cd services/user-service
JWT_PRIVATE_KEY="$(cat ../../private.pem)" \
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
ENV=staging \
SERVER_PORT=8081 \
DB_DSN=dbs/staging/user_staging.db \
JWT_ISSUER=we-circle-staging \
REDIS_HOST=localhost \
REDIS_PORT=6379 \
REDIS_PASSWORD= \
REDIS_DB=0 \
go run main.go

# 启动post-service (仅需公钥)
cd services/post-service
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
ENV=staging \
SERVER_PORT=8082 \
DB_DSN=dbs/staging/post_staging.db \
JWT_ISSUER=we-circle-staging \
REDIS_HOST=localhost \
REDIS_PORT=6379 \
REDIS_PASSWORD= \
REDIS_DB=0 \
go run main.go
```

## 故障排查

### 问题1: Redis连接失败
```
Failed to connect to Redis: dial tcp 127.0.0.1:6379: connect: connection refused
```
**解决方案**: 确保Redis服务已启动
```bash
brew services start redis
# 或
redis-server
```

### 问题2: Token验证失败
```
Invalid or expired token
```
**可能原因**:
- Token已过期
- Token已被撤销（在黑名单中）
- JWT密钥配置错误 (Private/Public Key 不匹配)
- Token格式错误

### 问题3: 环境变量未设置
```
JWT_PRIVATE_KEY is not set
```
**解决方案**: 确保所有必需的环境变量都已设置
