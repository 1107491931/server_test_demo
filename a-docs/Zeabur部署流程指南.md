# Zeabur 部署流程指南

本指南将帮助您完成 WeCircle 社交媒体微服务项目在 Zeabur 平台上的部署。

## 📋 目录

- [部署前准备](#部署前准备)
- [步骤 1：配置环境变量](#步骤-1配置环境变量)
- [步骤 2：推送代码到 Git 仓库](#步骤-2推送代码到-git-仓库)
- [步骤 3：创建 Zeabur 项目](#步骤-3创建-zeabur-项目)
- [步骤 4：配置服务依赖](#步骤-4配置服务依赖)
- [步骤 5：启动部署](#步骤-5启动部署)
- [步骤 6：验证部署](#步骤-6验证部署)
- [步骤 7：配置自定义域名](#步骤-7配置自定义域名)
- [故障排查](#故障排查)
- [后续维护](#后续维护)

## 🚀 部署前准备

### 1. 检查项目文件

确保以下文件已正确配置：

```bash
# 项目根目录
/Users/chomay/Desktop/go/server_test_demo/
├── zeabur.yaml                    # ✅ Zeabur 部署配置
├── zeabur-env-example.yaml        # ✅ 环境变量模板
├── private.pem                    # ✅ JWT 私钥
├── public.pem                     # ✅ JWT 公钥
├── docker-compose.build.yml       # ✅ Docker 构建配置
├── services/
│   ├── user-service/
│   │   ├── Dockerfile             # ✅ User Service Dockerfile
│   │   ├── handler/health.go      # ✅ 健康检查处理器
│   │   └── ...
│   └── post-service/
│       ├── Dockerfile             # ✅ Post Service Dockerfile
│       ├── handler/health.go      # ✅ 健康检查处理器
│       └── ...
└── ...
```

### 2. 验证 JWT 密钥

确保私钥和公钥文件存在且格式正确：

```bash
# 检查私钥
head -n 5 private.pem
# 应该显示：-----BEGIN RSA PRIVATE KEY-----

# 检查公钥
head -n 5 public.pem
# 应该显示：-----BEGIN PUBLIC KEY-----
```

### 3. 本地测试（可选但推荐）

在部署前先本地测试：

```bash
# 1. 构建 Docker 镜像
docker-compose -f docker-compose.build.yml build

# 2. 启动服务（使用本地 SQLite）
docker-compose -f docker-compose-staging.yml up -d

# 3. 测试健康检查端点
curl http://localhost:8081/health
curl http://localhost:8082/health

# 4. 测试 API 接口
# User Service: http://localhost:8081
# Post Service: http://localhost:8082
```

## 📝 步骤 1：配置环境变量

### 1.1 准备敏感信息

在 Zeabur 控制台配置前，需要准备以下敏感信息：

#### 生成 JWT 密钥（如果还没有）
```bash
# 生成 RSA 私钥（2048 位）
openssl genrsa -out private.pem 2048

# 从私钥生成公钥
openssl rsa -in private.pem -pubout -out public.pem

# 验证密钥格式
openssl rsa -in private.pem -check -noout
```

#### 生成强密码
```bash
# 使用 openssl 生成强密码
openssl rand -base64 32
# 或者使用 pwgen
brew install pwgen  # macOS
pwgen -s 32 1
```

### 1.2 在 Zeabur 控制台配置环境变量

1. **登录 Zeabur 控制台**
   - 访问：https://dash.zeabur.com
   - 登录您的账户

2. **进入项目设置**
   - 点击您创建的项目
   - 找到 **"Environment Variables"** 或 **"Variables"** 区域

3. **添加环境变量**
   
   对于 **user-service**：
   | 变量名 | 值 | 说明 |
   |--------|-----|------|
   | `ENV` | `prod` | 环境标识 |
   | `SERVER_PORT` | `8081` | 服务端口 |
   | `SERVICE_NAME` | `user-service` | 服务名称 |
   | `JWT_ISSUER` | `we-circle-prod` | JWT 发行者 |
   | `JWT_ALGORITHM` | `RS256` | JWT 算法 |
   | `REDIS_ENABLED` | `true` | 启用 Redis |
   | `LOG_LEVEL` | `info` | 日志级别 |
   | `CORS_ALLOWED_ORIGINS` | `https://*.zeabur.app` | 允许的域名 |
   | `RATE_LIMIT_ENABLED` | `true` | 启用速率限制 |
   | `RATE_LIMIT_REQUESTS` | `100` | 每分钟请求数 |
   | `RATE_LIMIT_WINDOW` | `60` | 时间窗口（秒） |
   | `MYSQL_PASSWORD` | *[您的强密码]* | **敏感：MySQL 密码** |
   | `JWT_PRIVATE_KEY` | *[完整的私钥内容]* | **敏感：JWT 私钥** |
   | `JWT_PUBLIC_KEY` | *[完整的公钥内容]* | **敏感：JWT 公钥** |

   对于 **post-service**：
   | 变量名 | 值 | 说明 |
   |--------|-----|------|
   | `ENV` | `prod` | 环境标识 |
   | `SERVER_PORT` | `8082` | 服务端口 |
   | `SERVICE_NAME` | `post-service` | 服务名称 |
   | `JWT_ISSUER` | `we-circle-prod` | JWT 发行者 |
   | `JWT_ALGORITHM` | `RS256` | JWT 算法 |
   | `REDIS_ENABLED` | `true` | 启用 Redis |
   | `LOG_LEVEL` | `info` | 日志级别 |
   | `CORS_ALLOWED_ORIGINS` | `https://*.zeabur.app` | 允许的域名 |
   | `RATE_LIMIT_ENABLED` | `true` | 启用速率限制 |
   | `RATE_LIMIT_REQUESTS` | `100` | 每分钟请求数 |
   | `RATE_LIMIT_WINDOW` | `60` | 时间窗口（秒） |
   | `MYSQL_PASSWORD` | *[与 user-service 相同的密码]* | **敏感：MySQL 密码** |
   | `JWT_PUBLIC_KEY` | *[完整的公钥内容]* | **敏感：JWT 公钥** |

4. **使用 Secrets 管理敏感信息**（推荐）
   - 点击 **"Secrets"** 或 **"Sensitive"** 选项
   - 将密码、密钥等敏感变量标记为 secrets
   - Zeabur 会自动加密这些值

### 1.3 配置文件 vs 控制台变量

- **配置文件 (`zeabur.yaml`)**：定义服务结构、依赖关系、资源限制
- **控制台变量**：配置敏感信息和运行时参数

建议将敏感信息放在控制台，通用配置放在文件。

## 📤 步骤 2：推送代码到 Git 仓库

### 2.1 创建 Git 仓库（如果还没有）

**GitHub：**
1. 访问 https://github.com/new
2. 仓库名称：`we-circle-backend`
3. 描述：`WeCircle 社交媒体微服务后端`
4. 选择 **Private** 或 **Public**
5. 不要初始化 README（我们已有代码）

**GitLab：**
1. 访问 https://gitlab.com/projects/new
2. 类似配置

### 2.2 初始化并推送代码

```bash
# 进入项目目录
cd /Users/chomay/Desktop/go/server_test_demo

# 1. 初始化 Git 仓库（如果还没有）
git init
git add .
git commit -m "feat: 初始化 WeCircle 微服务项目

- 添加 user-service 和 post-service 微服务
- 集成 JWT 认证系统
- 配置 Zeabur 部署模板
- 添加健康检查端点
- 完善 Docker 构建配置"

# 2. 添加远程仓库
git remote add origin https://github.com/您的用户名/we-circle-backend.git

# 3. 推送代码
git branch -M main
git push -u origin main
```

### 2.3 验证推送成功

```bash
# 查看远程仓库
git remote -v

# 应该显示：
# origin  https://github.com/您的用户名/we-circle-backend.git (fetch)
# origin  https://github.com/您的用户名/we-circle-backend.git (push)
```

## 🏗 步骤 3：创建 Zeabur 项目

### 3.1 登录 Zeabur

1. 访问 https://dash.zeabur.com
2. 使用 GitHub 账号登录（推荐）或邮箱注册

### 3.2 创建新项目

1. **创建项目**
   - 点击 **"New Project"** 或 **"Create Project"**
   - 项目名称：`we-circle-production`
   - 区域：选择 **US West (N. California)** 或 **Asia Pacific**
   - 点击 **"Create"**

2. **连接 Git 仓库**
   - 在项目页面，点击 **"Add Service"**
   - 选择 **"Import from Git"**
   - 找到并选择您的仓库：`we-circle-backend`
   - Zeabur 会自动检测 `zeabur.yaml` 文件

### 3.3 配置服务

Zeabur 会自动识别 `zeabur.yaml` 中的服务配置：

1. **user-service**
   - 应该自动检测为 Docker 服务
   - 构建目录：`services/user-service`
   - 端口：`8081`
   - 检查健康检查路径：`/health`

2. **post-service**
   - 自动检测为 Docker 服务
   - 构建目录：`services/post-service`
   - 端口：`8082`
   - 检查健康检查路径：`/health`

3. **mysql**
   - 自动检测为 MySQL 8.0 服务
   - 需要配置数据库名称：`app_db`

4. **redis**
   - 自动检测为 Redis 7.0 服务
   - 无需额外配置

### 3.4 验证服务依赖

确保服务依赖正确：
- `user-service` → 依赖 `mysql`, `redis`
- `post-service` → 依赖 `mysql`, `redis`, `user-service`

## ⚙️ 步骤 4：配置服务依赖

### 4.1 确认 MySQL 配置

1. 点击 **mysql** 服务
2. 检查配置：
   - **Database**: `app_db`
   - **Username**: `zeabur_user`
   - **Password**: `${MYSQL_PASSWORD}`（从环境变量注入）
3. 点击 **"Deploy"** 启动 MySQL

### 4.2 确认 Redis 配置

1. 点击 **redis** 服务
2. 配置（通常自动）：
   - **Port**: `6379`
   - **Version**: `7.0`
3. 点击 **"Deploy"** 启动 Redis

### 4.3 配置 User Service

1. 点击 **user-service** 服务
2. **Environment Variables** 区域：
   - 添加所有必需的环境变量（见步骤 1.2）
   - 确保 `MYSQL_PASSWORD` 和 `JWT_*` 密钥已配置
3. **Resources** 区域：
   - CPU: `500m`
   - Memory: `512Mi`
4. 点击 **"Deploy"** 启动

### 4.4 配置 Post Service

1. 点击 **post-service** 服务
2. **Environment Variables** 区域：
   - 添加所有必需的环境变量
   - 确保 `MYSQL_PASSWORD` 和 `JWT_PUBLIC_KEY` 已配置
3. **Resources** 区域：
   - CPU: `500m`
   - Memory: `512Mi`
4. 点击 **"Deploy"** 启动

## 🚀 步骤 5：启动部署

### 5.1 部署顺序

按以下顺序启动服务：

1. **MySQL** → 等待状态变为 **"Running"** ✅
2. **Redis** → 等待状态变为 **"Running"** ✅
3. **user-service** → 等待状态变为 **"Running"** ✅
4. **post-service** → 等待状态变为 **"Running"** ✅

### 5.2 监控部署进度

1. 在服务详情页，点击 **"Logs"** 标签
2. 观察构建和启动日志：
   ```bash
   # 典型日志输出
   [Build] Step 1/10 : FROM golang:1.21-alpine
   [Build] Step 2/10 : WORKDIR /app
   ...
   [Deploy] Starting service...
   [Health] Health check passed
   ```

3. 如果构建失败，检查错误日志并修复

### 5.3 常见构建问题

**问题：Go 模块下载超时**
```
go: github.com/gin-gonic/gin@v1.9.1: Get "https://proxy.golang.org/...": dial tcp: lookup proxy.golang.org: i/o timeout
```

**解决方案：**
1. 在 `Dockerfile` 中添加 GOPROXY：
   ```dockerfile
   ENV GOPROXY=https://goproxy.cn,direct
   ```

**问题：内存不足**
```
Killed signal terminated program as current=memory=512Mi limit=512Mi
```

**解决方案：**
1. 在 `zeabur.yaml` 中增加内存限制：
   ```yaml
   resources:
     memory: 1Gi
   ```

## ✅ 步骤 6：验证部署

### 6.1 检查服务状态

所有服务状态应为 **"Running"**（绿色勾选）

### 6.2 测试健康检查端点

```bash
# User Service 健康检查
curl https://user-service.zeabur.app/health
# 期望响应：{"status": "ok"}

# Post Service 健康检查
curl https://post-service.zeabur.app/health
# 期望响应：{"status": "ok"}
```

### 6.3 测试 API 接口

**User Service API：**
```bash
# 注册用户
curl -X POST https://user-service.zeabur.app/api/users/register \
  -H "Content-Type: application/json" \
  -d '{"username":"testuser","email":"test@example.com","password":"password123"}'

# 用户登录
curl -X POST https://user-service.zeabur.app/api/users/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'
```

**Post Service API：**
```bash
# 获取动态列表（需要认证）
curl -H "Authorization: Bearer <token>" \
  https://post-service.zeabur.app/api/posts

# 发布动态（需要认证）
curl -X POST -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{"content":"Hello WeCircle!","visibility":"public"}' \
  https://post-service.zeabur.app/api/posts
```

### 6.4 验证数据库连接

1. 在 Zeabur 控制台，点击 **mysql** 服务
2. 打开 **"Logs"** 标签
3. 检查是否有连接错误：
   ```
   [Info] Ready to accept connections on port 3306
   ```

### 6.5 验证 Redis 连接

1. 在 Zeabur 控制台，点击 **redis** 服务
2. 打开 **"Logs"** 标签
3. 检查连接状态：
   ```
   Ready to accept connections on port 6379
   ```

## 🌐 步骤 7：配置自定义域名（可选）

### 7.1 购买域名

推荐域名提供商：
- **Namecheap**: https://www.namecheap.com
- **阿里云**: https://www.aliyun.com
- **腾讯云**: https://cloud.tencent.com

域名示例：`api.we-circle.com`

### 7.2 添加域名到 Zeabur

1. 在 Zeabur 控制台，进入项目
2. 点击 **"Settings"** 或 **"Domains"**
3. 点击 **"Add Domain"**
4. 输入您的域名：`api.we-circle.com`
5. 选择服务：`user-service` 或 `post-service`
6. 点击 **"Add"**

### 7.3 配置 DNS

在域名提供商控制台添加 DNS 记录：

| 类型 | 主机记录 | 记录值 | TTL |
|------|----------|--------|-----|
| CNAME | api | your-project-id.zeabur.app | 300 |
| CNAME | www | your-project-id.zeabur.app | 300 |

### 7.4 等待 DNS 生效

- DNS 传播通常需要 5-30 分钟
- 最多可能需要 24 小时
- 使用以下命令验证：
  ```bash
  dig api.we-circle.com
  nslookup api.we-circle.com
  ```

### 7.5 配置 HTTPS

Zeabur 自动为自定义域名配置 HTTPS（Let's Encrypt）：

1. 确保 DNS 已正确配置
2. Zeabur 会自动获取 SSL 证书
3. 检查 **"HTTPS"** 状态变为 **"Active"**

## 🔧 故障排查

### 常见问题

#### 问题 1：服务无法启动

**症状：**
- 服务状态为 **"Failed"** 或 **"Crashed"**
- 日志显示崩溃信息

**排查步骤：**
1. 查看服务日志：
   ```bash
   # 在 Zeabur 控制台
   Service → Logs
   ```

2. 常见原因：
   - 环境变量缺失
   - 数据库连接失败
   - 内存不足
   - 端口冲突

**解决方案：**
```bash
# 1. 检查必需的环境变量
# 确保 JWT_PRIVATE_KEY, JWT_PUBLIC_KEY, MYSQL_PASSWORD 已配置

# 2. 检查数据库连接
# 确认 MySQL 服务已启动

# 3. 增加资源限制
# 在 zeabur.yaml 中增加 CPU 和内存
```

#### 问题 2：健康检查失败

**症状：**
- 健康检查持续失败
- 服务重启循环

**排查步骤：**
```bash
# 1. 测试健康检查端点
curl https://your-service.zeabur.app/health

# 2. 检查服务日志中的健康检查错误
```

**解决方案：**
1. 确保健康检查处理器存在：
   ```go
   // services/user-service/handler/health.go
   func HealthCheck(c *gin.Context) {
       c.JSON(200, gin.H{"status": "ok"})
   }
   ```

2. 在路由中注册健康检查：
   ```go
   // services/user-service/initialize/router.go
   router.GET("/health", handler.HealthCheck)
   ```

#### 问题 3：数据库连接失败

**症状：**
- 日志显示：`dial tcp: i/o timeout`
- 或 `Access denied for user`

**排查步骤：**
1. 确认 MySQL 服务状态
2. 检查环境变量：
   ```yaml
   # zeabur.yaml 或控制台
   MYSQL_HOST: mysql
   MYSQL_PORT: "3306"
   MYSQL_USER: zeabur_user
   MYSQL_PASSWORD: ${MYSQL_PASSWORD}
   ```

**解决方案：**
1. 等待 MySQL 完全启动（可能需要 1-2 分钟）
2. 验证密码是否正确
3. 检查依赖顺序：user-service/post-service 应在 MySQL 之后启动

#### 问题 4：JWT 认证失败

**症状：**
- API 返回 401 Unauthorized
- 日志显示 `invalid token`

**排查步骤：**
1. 验证 JWT 密钥配置：
   ```bash
   # 确保私钥和公钥匹配
   openssl rsa -in private.pem -pubout -out public_check.pem
   diff public.pem public_check.pem
   ```

2. 检查 JWT 发行者：
   ```yaml
   JWT_ISSUER: we-circle-prod
   ```

**解决方案：**
1. 重新配置正确的 JWT 密钥
2. 确保两个服务使用相同的 `JWT_ISSUER`

#### 问题 5：Redis 连接失败

**症状：**
- 日志显示：`dial tcp: connection refused`
- 或 `NOAUTH Authentication required`

**排查步骤：**
1. 检查 Redis 服务状态
2. 验证密码配置：
   ```yaml
   REDIS_HOST: redis
   REDIS_PORT: "6379"
   REDIS_PASSWORD: ${REDIS_PASSWORD}
   ```

**解决方案：**
1. 如果 Redis 没有设置密码，确保 `REDIS_PASSWORD` 为空
2. 如果设置了密码，在环境变量中配置正确的密码

### 获取帮助

如果以上方法无法解决问题：

1. **查看 Zeabur 文档**：
   - https://zeabur.com/docs

2. **联系 Zeabur 支持**：
   - 在控制台点击 **"Help"** 或 **"Support"**

3. **查看项目文档**：
   - `a-docs/Zeabur云部署指南.md`
   - `a-docs/故障排查指南.md`

## 🔄 后续维护

### 部署更新

当代码有更新时：

```bash
# 1. 本地提交更改
git add .
git commit -m "feat: 新功能描述"
git push origin main

# 2. Zeabur 自动检测并重新部署
# 在控制台查看部署进度
```

### 监控和日志

1. **监控服务状态**：
   - 在 Zeabur 控制台查看服务健康状态
   - 设置告警通知

2. **查看日志**：
   - 控制台 → Service → Logs
   - 支持实时日志流

3. **性能监控**：
   - 查看 CPU、内存使用情况
   - 设置资源告警

### 扩缩容

**水平扩展（增加副本数）：**
```yaml
# zeabur.yaml
deploy:
  replicas:
    user-service: 2
    post-service: 2
```

**垂直扩展（增加资源）：**
```yaml
# zeabur.yaml
resources:
  cpu: 1000m
  memory: 1Gi
```

### 备份和恢复

**数据库备份：**
- Zeabur MySQL 自动每日备份
- 可在控制台手动创建备份

**手动备份：**
```bash
# 导出数据
mysqldump -h mysql.zeabur.app -u zeabur_user -p app_db > backup.sql
```

### 安全最佳实践

1. **定期轮换密钥**：
   ```bash
   # 生成新密钥
   openssl genrsa -out new_private.pem 2048
   openssl rsa -in new_private.pem -pubout -out new_public.pem
   
   # 在 Zeabur 控制台更新密钥
   ```

2. **监控异常访问**：
   - 查看访问日志
   - 设置速率限制告警

3. **保持依赖更新**：
   - 定期更新 Go 依赖
   - 修复安全漏洞

## 📚 相关文档

- **Zeabur 官方文档**: https://zeabur.com/docs
- **项目部署指南**: `a-docs/Zeabur云部署指南.md`
- **JWT 认证**: `a-docs/JWT认证系统集成指南.md`
- **微服务通信**: `a-docs/微服务通信.md`
- **Docker 配置**: `a-docs/Docker构建多服务镜像.md`

## ✅ 部署检查清单

在开始部署前，请确认以下项目：

- [ ] JWT 密钥已生成并验证
- [ ] MySQL 强密码已准备
- [ ] 代码已推送到 Git 仓库
- [ ] zeabur.yaml 配置文件已更新
- [ ] zeabur-env-example.yaml 已参考
- [ ] 本地测试通过
- [ ] 域名已购买（如果需要自定义域名）
- [ ] DNS 配置已了解

---

**恭喜！** 完成以上步骤后，您的 WeCircle 微服务项目就成功部署在 Zeabur 平台上了！

如有问题，请参考故障排查部分或联系支持。
