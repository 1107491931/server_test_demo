# 📋 **Zeabur 详细部署指南 - Go 多服务项目**

## 🏗 **1. 项目结构调整**

### **当前结构分析**
```
├── zeabur.yaml                    # Zeabur 部署配置（新增）
├── docker-compose.zeabur.yml      # Zeabur 环境docker-compose（新增）
├── scripts/
│   ├── deploy-zeabur.sh          # 部署脚本（新增）
│   └── test-deployment.sh        # 测试脚本（新增）
└── services/
    ├── gateway/                   # API网关（新增）
    ├── user-service/
    └── post-service/
```

## 🔧 **2. Zeabur 配置文件详解**

### **`zeabur.yaml` - 完整配置**
```yaml
version: '1.0'
name: social-app

# ============ 全局配置 ============
environment: production
region: auto  # 自动选择可用区域

# ============ 环境变量 ============
variables:
  # 基础配置
  ENVIRONMENT: production
  LOG_LEVEL: info
  TZ: Asia/Shanghai
  
  # JWT 配置
  JWT_ALGORITHM: RS256
  JWT_EXPIRE_HOURS: "24"
  
  # 服务配置
  USER_SERVICE_PORT: "8081"
  POST_SERVICE_PORT: "8082"
  GATEWAY_PORT: "8080"
  
  # Redis 配置（Zeabur 会自动注入）
  REDIS_ENABLED: "true"
  
  # 数据库配置
  DB_DRIVER: "sqlite"  # 或 mysql
  DB_AUTO_MIGRATE: "true"

# ============ 密钥管理 ============
secrets:
  - name: JWT_PUBLIC_KEY
    value: ${JWT_PUBLIC_KEY}
  - name: JWT_PRIVATE_KEY
    value: ${JWT_PRIVATE_KEY}
  - name: DATABASE_ENCRYPT_KEY
    value: ${DATABASE_ENCRYPT_KEY}

# ============ 服务定义 ============
services:
  # ===== API 网关 =====
  gateway:
    name: api-gateway
    build:
      context: ./services/gateway
      dockerfile: Dockerfile
    port: 8080
    expose: true  # 对外暴露
    
    # 环境变量
    environment:
      - ENVIRONMENT=${ENVIRONMENT}
      - LOG_LEVEL=${LOG_LEVEL}
      - USER_SERVICE_URL=http://user-service:8081
      - POST_SERVICE_URL=http://post-service:8082
      - GATEWAY_PORT=${GATEWAY_PORT}
      - CORS_ALLOW_ORIGINS=*
      - RATE_LIMIT_ENABLED=true
      - RATE_LIMIT_REQUESTS=100
      - RATE_LIMIT_WINDOW=60
    
    # 健康检查
    health_check:
      path: /health
      port: 8080
      interval: 30s
      timeout: 5s
      retries: 3
    
    # 资源限制
    resources:
      memory: 128Mi
      cpu: 100m
    
    # 域名配置
    domains:
      - api.yourdomain.com
      - ${ZEABUR_PROJECT_ID}-gateway.zeabur.app
    
    # 路由配置
    routes:
      - path: /api/users/*
        service: user-service
        strip_prefix: true
      - path: /api/posts/*
        service: post-service
        strip_prefix: true
      - path: /*
        service: gateway
        action: serve

  # ===== 用户服务 =====
  user-service:
    build:
      context: ./services/user-service
      dockerfile: Dockerfile
    port: 8081
    expose: false  # 仅内部访问
    
    # 环境变量
    environment:
      - ENVIRONMENT=${ENVIRONMENT}
      - LOG_LEVEL=${LOG_LEVEL}
      - SERVICE_NAME=user-service
      - SERVICE_PORT=${USER_SERVICE_PORT}
      - DB_PATH=/data/user.db
      - DB_DRIVER=${DB_DRIVER}
      - DB_AUTO_MIGRATE=${DB_AUTO_MIGRATE}
      - JWT_PUBLIC_KEY=${JWT_PUBLIC_KEY}
      - JWT_PRIVATE_KEY=${JWT_PRIVATE_KEY}
      - JWT_EXPIRE_HOURS=${JWT_EXPIRE_HOURS}
      - POST_SERVICE_URL=http://post-service:8082
      - REDIS_ENABLED=${REDIS_ENABLED}
    
    # 健康检查
    health_check:
      path: /health
      port: 8081
      interval: 30s
    
    # 资源限制
    resources:
      memory: 256Mi
      cpu: 200m
    
    # 数据持久化
    volumes:
      - name: user-data
        path: /data
        size: 1Gi
        class: ssd
    
    # 依赖服务
    depends_on:
      - post-service

  # ===== 动态服务 =====
  post-service:
    build:
      context: ./services/post-service
      dockerfile: Dockerfile
    port: 8082
    expose: false
    
    environment:
      - ENVIRONMENT=${ENVIRONMENT}
      - LOG_LEVEL=${LOG_LEVEL}
      - SERVICE_NAME=post-service
      - SERVICE_PORT=${POST_SERVICE_PORT}
      - DB_PATH=/data/post.db
      - DB_DRIVER=${DB_DRIVER}
      - DB_AUTO_MIGRATE=${DB_AUTO_MIGRATE}
      - JWT_PUBLIC_KEY=${JWT_PUBLIC_KEY}
      - USER_SERVICE_URL=http://user-service:8081
      - REDIS_ENABLED=${REDIS_ENABLED}
    
    health_check:
      path: /health
      port: 8082
      interval: 30s
    
    resources:
      memory: 256Mi
      cpu: 200m
    
    volumes:
      - name: post-data
        path: /data
        size: 1Gi
        class: ssd

# ============ 附加服务 ============
addons:
  # Redis 缓存
  redis-cache:
    type: redis
    plan: free
    version: 7
    name: main-redis
    configuration:
      maxmemory: 256mb
      maxmemory-policy: allkeys-lru
    # Zeabur 会自动注入环境变量：
    # REDIS_HOST, REDIS_PORT, REDIS_PASSWORD, REDIS_URL
  
  # 如果需要 MySQL（可选）
  mysql-db:
    type: mysql
    plan: free
    version: 8.0
    name: main-mysql
    databases:
      - name: user_db
      - name: post_db
    # Zeabur 会自动注入：
    # MYSQL_HOST, MYSQL_PORT, MYSQL_USER, MYSQL_PASSWORD

# ============ 网络配置 ============
network:
  # 内部网络配置
  internal_domain_suffix: svc.cluster.local
  service_mesh: true
  
  # 外部访问
  ingress:
    class: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    tls:
      - hosts:
          - api.yourdomain.com
        secretName: api-tls-cert

# ============ 构建配置 ============
build:
  # 构建缓存
  cache_from:
    - golang:1.21-alpine
    - alpine:latest
  
  # 构建参数
  args:
    - GO_VERSION=1.21
    - GO_ENV=production
  
  # 构建优化
  optimization:
    minimize: true
    compress: true
```

## 🔐 **3. 环境变量详细配置**

### **环境变量配置文件**
```env
# .env.zeabur
# ===== 系统配置 =====
ENVIRONMENT=production
TZ=Asia/Shanghai
LOG_LEVEL=info
DEBUG=false

# ===== 服务配置 =====
USER_SERVICE_PORT=8081
POST_SERVICE_PORT=8082
GATEWAY_PORT=8080
USER_SERVICE_URL=http://user-service:8081
POST_SERVICE_URL=http://post-service:8082

# ===== JWT 配置 =====
JWT_ALGORITHM=RS256
JWT_EXPIRE_HOURS=24
JWT_ISSUER=social-app
JWT_AUDIENCE=social-app-users

# ===== 数据库配置 =====
DB_DRIVER=sqlite
DB_AUTO_MIGRATE=true
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME=3600

# ===== Redis 配置 =====
REDIS_ENABLED=true
REDIS_POOL_SIZE=10
REDIS_MIN_IDLE_CONNS=5
REDIS_MAX_RETRIES=3
REDIS_DIAL_TIMEOUT=5
REDIS_READ_TIMEOUT=3
REDIS_WRITE_TIMEOUT=3
REDIS_CACHE_TTL=300

# ===== 安全配置 =====
CORS_ALLOW_ORIGINS=*
CORS_ALLOW_CREDENTIALS=true
CORS_MAX_AGE=7200
RATE_LIMIT_ENABLED=true
RATE_LIMIT_REQUESTS=100
RATE_LIMIT_WINDOW=60

# ===== 文件存储 =====
UPLOAD_MAX_SIZE=10485760  # 10MB
UPLOAD_ALLOW_TYPES=image/jpeg,image/png,image/gif
```

### **Go 代码中读取环境变量**
```go
// services/common/config/config.go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Environment string
    ServicePort int
    
    // 数据库
    DBDriver        string
    DBPath          string
    DBAutoMigrate   bool
    
    // Redis
    RedisEnabled    bool
    RedisURL        string
    RedisPassword   string
    RedisDB         int
    
    // JWT
    JWTAlgorithm    string
    JWTPublicKey    string
    JWTPrivateKey   string
    JWTExpireHours  int
    
    // 服务发现
    UserServiceURL  string
    PostServiceURL  string
}

func LoadConfig() *Config {
    // 读取环境变量，提供默认值
    port, _ := strconv.Atoi(getEnv("SERVICE_PORT", "8080"))
    redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))
    jwtExpire, _ := strconv.Atoi(getEnv("JWT_EXPIRE_HOURS", "24"))
    
    return &Config{
        Environment:    getEnv("ENVIRONMENT", "development"),
        ServicePort:    port,
        
        DBDriver:       getEnv("DB_DRIVER", "sqlite"),
        DBPath:         getEnv("DB_PATH", "./data.db"),
        DBAutoMigrate:  getEnv("DB_AUTO_MIGRATE", "true") == "true",
        
        RedisEnabled:   getEnv("REDIS_ENABLED", "false") == "true",
        RedisURL:       getEnv("REDIS_URL", "redis://localhost:6379"),
        RedisPassword:  getEnv("REDIS_PASSWORD", ""),
        RedisDB:        redisDB,
        
        JWTAlgorithm:   getEnv("JWT_ALGORITHM", "RS256"),
        JWTPublicKey:   getEnv("JWT_PUBLIC_KEY", ""),
        JWTPrivateKey:  getEnv("JWT_PRIVATE_KEY", ""),
        JWTExpireHours: jwtExpire,
        
        UserServiceURL: getEnv("USER_SERVICE_URL", "http://localhost:8081"),
        PostServiceURL: getEnv("POST_SERVICE_URL", "http://localhost:8082"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}
```

## 🗄 **4. Redis 配置详细**

### **Redis 客户端配置**
```go
// services/common/redis/redis_client.go
package redis

import (
    "context"
    "fmt"
    "time"
    
    "github.com/go-redis/redis/v8"
)

var (
    client *redis.Client
    ctx    = context.Background()
)

func InitRedis(config *config.Config) (*redis.Client, error) {
    if !config.RedisEnabled {
        return nil, nil
    }
    
    opt := &redis.Options{
        Addr:     config.RedisURL,
        Password: config.RedisPassword,
        DB:       config.RedisDB,
        
        // 连接池配置
        PoolSize:     10,
        MinIdleConns: 5,
        
        // 超时设置
        DialTimeout:  5 * time.Second,
        ReadTimeout:  3 * time.Second,
        WriteTimeout: 3 * time.Second,
        PoolTimeout:  4 * time.Second,
        
        // 重试
        MaxRetries:      3,
        MinRetryBackoff: 8 * time.Millisecond,
        MaxRetryBackoff: 512 * time.Millisecond,
    }
    
    client = redis.NewClient(opt)
    
    // 测试连接
    if err := client.Ping(ctx).Err(); err != nil {
        return nil, fmt.Errorf("redis连接失败: %v", err)
    }
    
    return client, nil
}

// 缓存封装
func SetCache(key string, value interface{}, expiration time.Duration) error {
    return client.Set(ctx, key, value, expiration).Err()
}

func GetCache(key string) (string, error) {
    return client.Get(ctx, key).Result()
}

func DeleteCache(key string) error {
    return client.Del(ctx, key).Err()
}

// 分布式锁
func AcquireLock(key string, expiration time.Duration) bool {
    return client.SetNX(ctx, key, "locked", expiration).Val()
}

func ReleaseLock(key string) error {
    return client.Del(ctx, key).Err()
}
```

### **Redis 使用示例**
```go
// services/user-service/handler/auth_handler.go
package handler

import (
    "github.com/gin-gonic/gin"
    "time"
)

func (h *UserHandler) Login(c *gin.Context) {
    // ... 验证逻辑
    
    // 1. 缓存用户信息
    cacheKey := fmt.Sprintf("user:%d", user.ID)
    userJSON, _ := json.Marshal(user)
    redis.SetCache(cacheKey, userJSON, 30*time.Minute)
    
    // 2. 缓存 Token
    tokenKey := fmt.Sprintf("token:%s", token)
    redis.SetCache(tokenKey, user.ID, 24*time.Hour)
    
    // 3. 记录登录日志
    logKey := fmt.Sprintf("login:user:%d", user.ID)
    redis.LPush(ctx, logKey, time.Now().String())
    redis.LTrim(ctx, logKey, 0, 9) // 只保留最近10条
}
```

## 🗃 **5. 数据库配置详细**

### **多环境数据库配置**
```go
// services/user-service/initialize/db.go
package initialize

import (
    "database/sql"
    "fmt"
    "log"
    "os"
    "path/filepath"
    
    "gorm.io/driver/mysql"
    "gorm.io/driver/sqlite"
    "gorm.io/gorm"
    "gorm.io/gorm/logger"
)

func InitDB(config *config.Config) (*gorm.DB, error) {
    var dialector gorm.Dialector
    
    switch config.DBDriver {
    case "mysql":
        // Zeabur 提供的 MySQL
        dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
            os.Getenv("MYSQL_USER"),
            os.Getenv("MYSQL_PASSWORD"),
            os.Getenv("MYSQL_HOST"),
            os.Getenv("MYSQL_PORT"),
            "user_db",
        )
        dialector = mysql.Open(dsn)
        
    case "sqlite":
        // Zeabur 持久化存储
        dbPath := config.DBPath
        if dbPath == "" {
            dbPath = "/data/user.db"
        }
        
        // 确保目录存在
        if err := os.MkdirAll(filepath.Dir(dbPath), 0755); err != nil {
            return nil, err
        }
        
        dialector = sqlite.Open(dbPath)
        
    default:
        return nil, fmt.Errorf("不支持的数据库驱动: %s", config.DBDriver)
    }
    
    // GORM 配置
    gormConfig := &gorm.Config{
        Logger: logger.Default.LogMode(logger.Info),
        PrepareStmt: true,
        SkipDefaultTransaction: true,
    }
    
    // 连接数据库
    db, err := gorm.Open(dialector, gormConfig)
    if err != nil {
        return nil, err
    }
    
    // 获取通用数据库对象
    sqlDB, err := db.DB()
    if err != nil {
        return nil, err
    }
    
    // 连接池配置
    sqlDB.SetMaxIdleConns(10)
    sqlDB.SetMaxOpenConns(100)
    sqlDB.SetConnMaxLifetime(time.Hour)
    
    // 自动迁移
    if config.DBAutoMigrate {
        if err := db.AutoMigrate(&model.User{}); err != nil {
            log.Printf("自动迁移失败: %v", err)
        }
    }
    
    return db, nil
}
```

### **数据库初始化脚本**
```bash
#!/bin/bash
# scripts/init-db.sh

echo "初始化数据库..."

# 创建数据目录
mkdir -p /data

# 根据环境初始化数据库
if [ "$ENVIRONMENT" = "production" ]; then
    echo "生产环境初始化..."
    # 从备份恢复或初始化
    if [ ! -f /data/user.db ]; then
        sqlite3 /data/user.db ".read init/user_schema.sql"
    fi
    if [ ! -f /data/post.db ]; then
        sqlite3 /data/post.db ".read init/post_schema.sql"
    fi
else
    echo "开发环境初始化..."
    # 测试数据
    sqlite3 /data/user.db ".read test/test_data.sql"
fi

echo "数据库初始化完成"
```

## 🔗 **6. 服务间通信详细配置**

### **HTTP 客户端封装**
```go
// services/common/client/service_client.go
package client

import (
    "bytes"
    "context"
    "encoding/json"
    "fmt"
    "io"
    "net/http"
    "time"
    
    "github.com/sony/gobreaker"
)

type ServiceClient struct {
    baseURL    string
    httpClient *http.Client
    circuitBreaker *gobreaker.CircuitBreaker
    cache      *redis.Client
}

func NewServiceClient(serviceName, baseURL string) *ServiceClient {
    // 断路器配置
    cb := gobreaker.NewCircuitBreaker(gobreaker.Settings{
        Name:        serviceName,
        MaxRequests: 5,
        Interval:    30 * time.Second,
        Timeout:     60 * time.Second,
        ReadyToTrip: func(counts gobreaker.Counts) bool {
            failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
            return counts.Requests >= 3 && failureRatio >= 0.6
        },
        OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
            log.Printf("断路器 %s 状态从 %s 变为 %s", name, from, to)
        },
    })
    
    return &ServiceClient{
        baseURL: baseURL,
        httpClient: &http.Client{
            Timeout: 10 * time.Second,
            Transport: &http.Transport{
                MaxIdleConns:        100,
                MaxIdleConnsPerHost: 10,
                IdleConnTimeout:     30 * time.Second,
            },
        },
        circuitBreaker: cb,
    }
}

// 带缓存的 GET 请求
func (c *ServiceClient) GetWithCache(ctx context.Context, path, cacheKey string, result interface{}) error {
    // 1. 尝试从缓存获取
    if cached, err := c.cache.Get(ctx, cacheKey).Result(); err == nil {
        if err := json.Unmarshal([]byte(cached), result); err == nil {
            return nil
        }
    }
    
    // 2. 调用服务
    resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
        req, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+path, nil)
        if err != nil {
            return nil, err
        }
        
        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, err
        }
        
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }
        
        if resp.StatusCode >= 400 {
            return nil, fmt.Errorf("服务调用失败: %s", body)
        }
        
        return body, nil
    })
    
    if err != nil {
        return err
    }
    
    body := resp.([]byte)
    
    // 3. 解析响应
    if err := json.Unmarshal(body, result); err != nil {
        return err
    }
    
    // 4. 缓存结果（5分钟）
    c.cache.Set(ctx, cacheKey, body, 5*time.Minute)
    
    return nil
}

// POST 请求
func (c *ServiceClient) Post(ctx context.Context, path string, data interface{}) ([]byte, error) {
    jsonData, err := json.Marshal(data)
    if err != nil {
        return nil, err
    }
    
    resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
        req, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+path, bytes.NewBuffer(jsonData))
        if err != nil {
            return nil, err
        }
        req.Header.Set("Content-Type", "application/json")
        
        resp, err := c.httpClient.Do(req)
        if err != nil {
            return nil, err
        }
        
        defer resp.Body.Close()
        body, err := io.ReadAll(resp.Body)
        if err != nil {
            return nil, err
        }
        
        if resp.StatusCode >= 400 {
            return nil, fmt.Errorf("服务调用失败: %s", body)
        }
        
        return body, nil
    })
    
    if err != nil {
        return nil, err
    }
    
    return resp.([]byte), nil
}
```

### **服务间调用示例**
```go
// services/post-service/client/user_client.go
package client

import (
    "context"
    "encoding/json"
    "fmt"
    
    "github.com/your-org/social-app/common/dto"
)

type UserClient struct {
    client *ServiceClient
}

func NewUserClient() *UserClient {
    baseURL := config.LoadConfig().UserServiceURL
    return &UserClient{
        client: NewServiceClient("user-service", baseURL),
    }
}

func (c *UserClient) GetUser(ctx context.Context, userID int64) (*dto.UserDTO, error) {
    var user dto.UserDTO
    
    cacheKey := fmt.Sprintf("user:%d", userID)
    path := fmt.Sprintf("/api/users/%d", userID)
    
    if err := c.client.GetWithCache(ctx, path, cacheKey, &user); err != nil {
        return nil, fmt.Errorf("获取用户信息失败: %v", err)
    }
    
    return &user, nil
}

func (c *UserClient) ValidateToken(ctx context.Context, token string) (*dto.AuthResponse, error) {
    var authResp dto.AuthResponse
    
    data := map[string]string{
        "token": token,
    }
    
    body, err := c.client.Post(ctx, "/api/auth/validate", data)
    if err != nil {
        return nil, err
    }
    
    if err := json.Unmarshal(body, &authResp); err != nil {
        return nil, err
    }
    
    return &authResp, nil
}
```

## 🌐 **7. HTTPS + 固定域名配置**

### **域名和 SSL 配置**
```yaml
# zeabur.yaml 的域名部分
domains:
  - host: api.yourdomain.com
    type: custom
    tls:
      enabled: true
      type: managed  # Zeabur 自动管理证书
      issuer: letsencrypt
      auto_renew: true
      http_challenge: true
    
  - host: social.yourdomain.com
    type: custom
    tls:
      enabled: true
    
  # Zeabur 默认域名（备用）
  - host: ${ZEABUR_PROJECT_ID}-gateway.zeabur.app
    type: zeabur
    tls: true

# DNS 记录配置
dns:
  - type: CNAME
    name: api
    value: ${ZEABUR_PROJECT_ID}.zeabur.app
    ttl: 300
    proxied: true  # 通过 Cloudflare 代理
  
  - type: CNAME
    name: social
    value: ${ZEABUR_PROJECT_ID}.zeabur.app
    ttl: 300
    proxied: true
  
  # 如果需要 A 记录（固定 IP）
  # - type: A
  #   name: @
  #   value: ${ZEABUR_STATIC_IP}
  #   ttl: 300
```

### **获取固定 IP 的解决方案**
由于 Zeabur 使用动态 IP，可以通过以下方式实现"固定"访问：

#### **方案1：Cloudflare 代理 + Workers**
```javascript
// cloudflare-worker.js
addEventListener('fetch', event => {
    event.respondWith(handleRequest(event.request))
})

async function handleRequest(request) {
    // 获取 Zeabur 服务的当前 IP
    const zeaburIP = await getZeaburIP()
    
    // 修改请求头
    const newRequest = new Request(request, {
        headers: {
            ...request.headers,
            'X-Real-IP': request.headers.get('CF-Connecting-IP'),
            'X-Forwarded-For': request.headers.get('CF-Connecting-IP'),
        }
    })
    
    // 转发到 Zeabur
    const url = new URL(request.url)
    url.hostname = zeaburIP
    
    return fetch(new Request(url, newRequest))
}

// 定期更新 IP
async function getZeaburIP() {
    // 从 KV 存储获取缓存的 IP
    let ip = await SOCIAL_IPS.get('zeabur_ip')
    
    if (!ip) {
        // 解析域名获取最新 IP
        const response = await fetch('https://dns.google/resolve?name=your-project.zeabur.app&type=A')
        const data = await response.json()
        ip = data.Answer[0].data
        
        // 缓存 5 分钟
        await SOCIAL_IPS.put('zeabur_ip', ip, { expirationTtl: 300 })
    }
    
    return ip
}
```

#### **方案2：动态 DNS 更新脚本**
```bash
#!/bin/bash
# scripts/update-dns.sh

#!/bin/bash
# 自动更新 DNS 记录到当前 Zeabur IP

ZEABUR_DOMAIN="your-project.zeabur.app"
CUSTOM_DOMAIN="api.yourdomain.com"
CLOUDFLARE_API_TOKEN="your-api-token"
ZONE_ID="your-zone-id"

# 获取 Zeabur 当前 IP
CURRENT_IP=$(dig +short $ZEABUR_DOMAIN | head -n1)

# 获取当前 DNS 记录 IP
DNS_IP=$(dig +short $CUSTOM_DOMAIN | head -n1)

# 如果 IP 变化，更新 DNS
if [ "$CURRENT_IP" != "$DNS_IP" ]; then
    echo "IP 变化: $DNS_IP -> $CURRENT_IP"
    
    # 获取 DNS 记录 ID
    RECORD_ID=$(curl -s -X GET "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records?name=$CUSTOM_DOMAIN" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" | jq -r '.result[0].id')
    
    # 更新 DNS 记录
    curl -s -X PUT "https://api.cloudflare.com/client/v4/zones/$ZONE_ID/dns_records/$RECORD_ID" \
        -H "Authorization: Bearer $CLOUDFLARE_API_TOKEN" \
        -H "Content-Type: application/json" \
        --data "{\"type\":\"A\",\"name\":\"api\",\"content\":\"$CURRENT_IP\",\"ttl\":300,\"proxied\":true}"
    
    echo "DNS 记录已更新"
else
    echo "IP 未变化: $CURRENT_IP"
fi
```

### **HTTPS 中间件配置**
```go
// services/gateway/middleware/security.go
package middleware

import (
    "github.com/gin-gonic/gin"
    "net/http"
    "strings"
)

// 强制 HTTPS 重定向
func ForceHTTPS() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 在 Zeabur 中，X-Forwarded-Proto 表示原始协议
        if c.GetHeader("X-Forwarded-Proto") == "http" {
            url := c.Request.URL
            url.Scheme = "https"
            url.Host = c.Request.Host
            
            // 301 永久重定向
            c.Redirect(http.StatusMovedPermanently, url.String())
            c.Abort()
            return
        }
        c.Next()
    }
}

// HSTS 头设置
func HSTS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains; preload")
        c.Next()
    }
}

// 安全头设置
func SecurityHeaders() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Header("X-Content-Type-Options", "nosniff")
        c.Header("X-Frame-Options", "DENY")
        c.Header("X-XSS-Protection", "1; mode=block")
        c.Header("Referrer-Policy", "strict-origin-when-cross-origin")
        c.Header("Content-Security-Policy", "default-src 'self'")
        c.Next()
    }
}
```

## 🚀 **8. 完整部署流程**

### **第一步：准备工作**
```bash
# 1. 安装 Zeabur CLI
npm install -g @zeabur/cli

# 2. 登录
zeabur login

# 3. 创建项目
zeabur project create social-app

# 4. 获取项目 ID
PROJECT_ID=$(zeabur project list | grep social-app | awk '{print $1}')
echo "项目ID: $PROJECT_ID"
```

### **第二步：配置环境变量**
```bash
# 设置环境变量
zeabur variable set ENVIRONMENT=production
zeabur variable set LOG_LEVEL=info

# 设置 JWT 密钥
zeabur secret set JWT_PUBLIC_KEY "$(cat public.pem)"
zeabur secret set JWT_PRIVATE_KEY "$(cat private.pem)"

# 设置数据库加密密钥
zeabur secret set DATABASE_ENCRYPT_KEY "$(openssl rand -base64 32)"
```

### **第三步：部署服务**
```bash
#!/bin/bash
# scripts/deploy-all.sh

echo "🚀 开始部署到 Zeabur..."

# 1. 构建并推送镜像
echo "构建镜像..."
docker build -t user-service ./services/user-service
docker build -t post-service ./services/post-service
docker build -t gateway ./services/gateway

# 2. 应用配置
echo "应用配置..."
zeabur apply -f zeabur.yaml

# 3. 等待部署完成
echo "等待部署..."
zeabur status --watch

# 4. 获取服务地址
GATEWAY_URL=$(zeabur service get gateway --output json | jq -r '.domains[0]')
echo "网关地址: $GATEWAY_URL"

# 5. 测试部署
echo "测试部署..."
./scripts/test-deployment.sh $GATEWAY_URL

echo "✅ 部署完成!"
```

### **第四步：绑定域名**
```bash
#!/bin/bash
# scripts/setup-domain.sh

# 1. 添加自定义域名
zeabur domain add api.yourdomain.com

# 2. 获取 DNS 验证记录
zeabur domain verify api.yourdomain.com

# 3. 到域名服务商添加 CNAME 记录
# api.yourdomain.com CNAME -> your-project-id.zeabur.app

# 4. 等待 SSL 证书签发
zeabur domain status api.yourdomain.com

echo "域名配置完成"
```

### **第五步：验证部署**
```bash
#!/bin/bash
# scripts/verify-deployment.sh

BASE_URL=${1:-https://api.yourdomain.com}

echo "🔍 验证部署..."

# 1. 检查 HTTPS
echo "1. 检查 HTTPS..."
curl -I $BASE_URL/health --insecure

# 2. 检查服务健康
echo "2. 检查服务健康..."
curl -s $BASE_URL/health | jq '.'

# 3. 检查数据库连接
echo "3. 检查数据库连接..."
curl -s $BASE_URL/api/debug/db | jq '.'

# 4. 检查 Redis 连接
echo "4. 检查 Redis 连接..."
curl -s $BASE_URL/api/debug/redis | jq '.'

# 5. 性能测试
echo "5. 性能测试..."
ab -n 100 -c 10 $BASE_URL/health

echo "✅ 验证完成"
```

## 📊 **9. 监控和维护**

### **查看日志**
```bash
# 实时查看网关日志
zeabur logs --service gateway --tail 100 --follow

# 查看用户服务错误日志
zeabur logs --service user-service --level error --since 1h

# 导出日志到文件
zeabur logs --service post-service --since 24h > post-service.log
```

### **监控指标**
```bash
# 查看资源使用
zeabur metrics --service user-service --period 1h

# 查看请求统计
zeabur stats --service gateway --since 1d

# 查看错误率
zeabur alerts list
```

### **备份和恢复**
```bash
# 备份数据库
zeabur backup create user-data --service user-service

# 恢复数据库
zeabur backup restore user-data-20240101 --service user-service

# 导出配置
zeabur export --output zeabur-backup.yaml
```

## 🔧 **10. 故障排除**

### **常见问题解决**

#### **问题1：服务启动失败**
```bash
# 查看详细日志
zeabur logs --service user-service --since 10m --level debug

# 检查环境变量
zeabur variable list --service user-service

# 重新部署
zeabur redeploy user-service
```

#### **问题2：数据库连接失败**
```bash
# 检查数据库状态
zeabur addon status mysql-db

# 重置数据库密码
zeabur addon reset-password mysql-db

# 测试连接
zeabur exec user-service -- curl http://mysql-db:3306
```

#### **问题3：服务间通信失败**
```bash
# 检查网络连接
zeabur exec user-service -- ping post-service

# 检查 DNS 解析
zeabur exec user-service -- nslookup post-service

# 测试服务端点
zeabur exec user-service -- curl http://post-service:8082/health
```

### **性能优化建议**
1. **启用缓存**：Redis 缓存热点数据
2. **数据库优化**：添加索引，使用连接池
3. **CDN 加速**：静态资源通过 CDN 分发
4. **负载均衡**：多个实例负载均衡
5. **监控告警**：设置性能阈值告警

## 📞 **总结**

通过以上详细配置，你的 Go 多服务项目可以在 Zeabur 上实现：

✅ **环境变量管理** - 安全、分环境配置  
✅ **Redis 缓存** - 高性能缓存支持  
✅ **数据库持久化** - 数据安全存储  
✅ **服务间通信** - 可靠、带缓存的通信  
✅ **HTTPS 支持** - 自动 SSL 证书管理  
✅ **固定域名访问** - 稳定可访问的域名  
✅ **监控维护** - 完整的运维支持  

部署完成后，你的服务将通过 `https://api.yourdomain.com` 对外提供服务，具备生产级别的稳定性和安全性。