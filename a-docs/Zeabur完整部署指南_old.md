# Zeabur 完整部署指南

## 📋 目录

1. [项目配置文件说明](#项目配置文件说明)
2. [Zeabur 平台配置流程](#zeabur-平台配置流程)
3. [部署后测试](#部署后测试)
4. [常见问题排查](#常见问题排查)

---

## 项目配置文件说明

### 1. 核心配置文件

#### `zeabur.yaml`
**作用**：Zeabur 平台的主配置文件，定义了所有服务的构建、部署和运行配置。

**关键配置项**：
```yaml
services:
  user-service:
    build:
      context: .                                    # 从项目根目录构建
      dockerfile: services/user-service/Dockerfile  # Dockerfile 路径
    env:
      DB_DSN: /app/dbs/prod/user_prod.db           # 数据库连接
      JWT_PRIVATE_KEY: ${JWT_PRIVATE_KEY}          # JWT 私钥（从 secrets 引用）
      JWT_PUBLIC_KEY: ${JWT_PUBLIC_KEY}            # JWT 公钥（从 secrets 引用）
    dependsOn:
      - mysql                                       # 依赖 MySQL 服务
      - redis                                       # 依赖 Redis 服务
```

**重要说明**：
- `context: .` 表示从项目根目录作为构建上下文
- 环境变量中的 `${JWT_PRIVATE_KEY}` 会从 Zeabur 控制台配置的环境变量中读取
- `dependsOn` 确保服务按正确顺序启动

---

#### `zeabur-env-example.yaml`
**作用**：环境变量配置示例和说明文档，不会被 Zeabur 直接使用。

**用途**：
1. 作为环境变量配置的参考文档
2. 说明每个环境变量的作用和配置方法
3. 提供 JWT 密钥配置的详细步骤

**重要内容**：
- JWT 密钥获取命令
- 数据库 DSN 配置方式（SQLite vs MySQL）
- Redis 配置说明
- 部署步骤概览

---

### 2. Docker 配置文件

#### `services/user-service/Dockerfile`
**作用**：用户服务的 Docker 镜像构建配置。

**构建流程**：
```dockerfile
# 第一阶段：构建
FROM golang:1.24-alpine AS builder
COPY services/common ./common                      # 复制共享模块
COPY services/user-service/go.mod ...              # 复制依赖文件
RUN go mod download                                 # 下载依赖
COPY services/user-service .                        # 复制源代码
RUN go build -o user-service .                      # 编译

# 第二阶段：运行
FROM alpine:latest
COPY --from=builder /build/user-service/user-service .  # 复制编译结果
RUN mkdir -p /app/dbs/prod /app/keys                # 创建必要目录
CMD ["./user-service"]                               # 启动服务
```

**关键点**：
- 使用多阶段构建减小镜像体积
- Go 1.24 版本与 go.mod 要求一致
- 创建 `/app/keys` 目录用于存放 JWT 密钥（如果需要）
- 创建 `/app/dbs/prod` 目录用于 SQLite 数据库

---

#### `services/post-service/Dockerfile`
**作用**：帖子服务的 Docker 镜像构建配置。

**与 user-service Dockerfile 的区别**：
- 端口号不同（8082 vs 8081）
- 数据库文件路径不同（post_prod.db vs user_prod.db）
- 其他构建流程完全相同

---

#### `docker-compose.build.yml`
**作用**：本地 Docker 镜像构建配置，用于本地开发和测试。

**使用方法**：
```bash
# 构建所有服务
docker-compose -f docker-compose.build.yml build --parallel

# 构建并指定版本号
IMAGE_TAG=v1.0.0 docker-compose -f docker-compose.build.yml build
```

**配置说明**：
```yaml
services:
  user-service:
    build:
      context: .                                    # 与 Zeabur 配置一致
      dockerfile: services/user-service/Dockerfile
    image: user-service:${IMAGE_TAG:-latest}        # 镜像名称和标签
```

**重要**：构建上下文已更新为项目根目录（`.`），与 Zeabur 配置保持一致。

---

### 3. JWT 密钥文件

#### `private.pem`
**作用**：JWT 签名的 RSA 私钥，用于生成 JWT Token。

**使用场景**：
- user-service 使用私钥签名 JWT Token
- 用户登录时生成 Access Token 和 Refresh Token

**安全要求**：
- ⚠️ 绝对不能泄露或提交到公开仓库
- 已在 `.gitignore` 中排除
- 部署时通过环境变量传递给容器

---

#### `public.pem`
**作用**：JWT 验证的 RSA 公钥，用于验证 JWT Token 的有效性。

**使用场景**：
- user-service 和 post-service 都使用公钥验证 Token
- 验证用户请求中的 Authorization header

**安全说明**：
- 公钥可以公开，但建议通过环境变量传递
- 确保与私钥配对使用

---

### 4. 其他配置文件

#### `docker-compose-staging.yml` / `docker-compose-pre.yml` / `docker-compose-prod.yml`
**作用**：不同环境的 Docker Compose 运行配置。

**用途**：本地多环境部署测试，Zeabur 部署时不使用这些文件。

---

## Zeabur 平台配置流程

### 前置准备

#### 1. 推送代码到 Git 仓库

```bash
# 确保所有修改已提交
git status

# 提交修改
git add .
git commit -m "完成 Zeabur 部署配置"

# 推送到远程仓库
git push origin main
```

#### 2. 准备 JWT 密钥内容

**获取私钥**：
```bash
# macOS
cat private.pem | pbcopy

# Linux
cat private.pem | xclip -selection clipboard

# Windows (PowerShell)
Get-Content private.pem | Set-Clipboard

# 或者直接查看
cat private.pem
```

将输出的完整内容（包括 `-----BEGIN PRIVATE KEY-----` 和 `-----END PRIVATE KEY-----`）保存到文本文件。

**获取公钥**：
```bash
# macOS
cat public.pem | pbcopy

# Linux
cat public.pem | xclip -selection clipboard

# 或者直接查看
cat public.pem
```

同样保存完整内容。

---

### 第一步：登录 Zeabur

1. 访问 https://zeabur.com
2. 点击右上角 **"Sign In"** 或 **"登录"**
3. 选择登录方式：
   - GitHub（推荐）
   - GitLab
   - Email

4. 授权 Zeabur 访问你的 Git 账号

---

### 第二步：创建项目

1. 登录后，点击 **"New Project"** 或 **"创建项目"**
2. 填写项目信息：
   - **Project Name**：`we-circle-social-app`（或自定义名称）
   - **Region**：选择 **US West (Silicon Valley)** 或其他区域
   
3. 点击 **"Create"** 或 **"创建"**

---

### 第三步：添加 Redis 服务

1. 在项目页面，点击 **"Add Service"** 或 **"添加服务"**
2. 选择 **"Prebuilt Service"** 或 **"预构建服务"**
3. 在服务列表中找到 **"Redis"**
4. 配置 Redis：
   - **Version**：选择 `7.0` 或最新稳定版
   - **Name**：保持默认 `redis` 或自定义
   
5. 点击 **"Deploy"** 或 **"部署"**
6. 等待 Redis 服务状态变为 **"Running"**（绿色）

**验证**：
- 服务列表中显示 Redis 服务
- 状态为 Running
- 可以看到 Redis 的连接信息（Host、Port）

---

### 第四步：（可选）添加 MySQL 服务

如果需要使用 MySQL 而不是 SQLite：

1. 点击 **"Add Service"**
2. 选择 **"Prebuilt Service"**
3. 找到并选择 **"MySQL"**
4. 配置 MySQL：
   - **Version**：`8.0`
   - **Database Name**：`app_db`
   - **Username**：`zeabur_user`
   - **Password**：自动生成（Zeabur 会自动设置）
   
5. 点击 **"Deploy"**
6. 等待 MySQL 服务启动完成

**注意**：
- Zeabur 会自动注入环境变量：`MYSQL_HOST`、`MYSQL_PORT`、`MYSQL_USER`、`MYSQL_PASSWORD`、`MYSQL_DATABASE`
- 如果使用 MySQL，需要在后续步骤中修改 `DB_DSN` 环境变量

---

### 第五步：导入 Git 仓库

1. 点击 **"Add Service"**
2. 选择 **"Git"** 或 **"从 Git 部署"**
3. 选择 Git 提供商：
   - GitHub
   - GitLab
   - Bitbucket
   
4. 如果是首次使用，需要授权 Zeabur 访问你的仓库
5. 在仓库列表中选择 `server_test_demo`
6. 选择分支：`main` 或 `master`
7. Zeabur 会自动检测到 `zeabur.yaml` 配置文件
8. **暂时不要点击部署**，先配置环境变量

---

### 第六步：配置环境变量（关键步骤）

#### 方式一：项目级别配置（推荐）

1. 在项目页面，点击顶部的 **"Settings"** 或 **"设置"** 标签
2. 在左侧菜单找到 **"Environment Variables"** 或 **"环境变量"**
3. 点击 **"Add Variable"** 或 **"添加变量"**

**添加以下环境变量**：

| 变量名 | 变量值 | 类型 | 说明 |
|--------|--------|------|------|
| `JWT_PRIVATE_KEY` | 粘贴 `private.pem` 的完整内容 | Secret | JWT 签名私钥 |
| `JWT_PUBLIC_KEY` | 粘贴 `public.pem` 的完整内容 | Secret | JWT 验证公钥 |

**配置步骤**：

1. **添加 JWT_PRIVATE_KEY**：
   - Variable Name: `JWT_PRIVATE_KEY`
   - Variable Value: 粘贴之前复制的 `private.pem` 完整内容
   - Type: 选择 **"Secret"**（加密存储）
   - 点击 **"Add"** 或 **"添加"**

2. **添加 JWT_PUBLIC_KEY**：
   - Variable Name: `JWT_PUBLIC_KEY`
   - Variable Value: 粘贴之前复制的 `public.pem` 完整内容
   - Type: 选择 **"Secret"**
   - 点击 **"Add"**

**验证**：
- 环境变量列表中显示两个变量
- Secret 类型的变量值会显示为 `***`（已加密）

---

#### 方式二：服务级别配置（可选）

如果需要为每个服务单独配置不同的环境变量：

1. 点击具体的服务（如 `user-service`）
2. 进入服务详情页
3. 点击 **"Variables"** 或 **"环境变量"** 标签
4. 添加服务专属的环境变量

**注意**：
- 服务级别的环境变量会覆盖项目级别的同名变量
- 对于 JWT 密钥，建议在项目级别配置，所有服务共享

---

#### 如果使用 MySQL（可选配置）

如果在第四步添加了 MySQL 服务，需要修改 `DB_DSN` 环境变量：

**方式 1：使用 Zeabur 自动注入的变量**

添加环境变量：
- Variable Name: `DB_DSN`
- Variable Value: `${MYSQL_USER}:${MYSQL_PASSWORD}@tcp(${MYSQL_HOST}:${MYSQL_PORT})/${MYSQL_DATABASE}?charset=utf8mb4&parseTime=True&loc=Local`

**方式 2：修改代码自动构建 DSN**

在 `services/user-service/initialize/config.go` 中添加逻辑，自动从环境变量构建 MySQL DSN。

---

### 第七步：部署服务

1. 返回项目主页
2. Zeabur 会自动检测到代码变更和配置
3. 点击服务的 **"Deploy"** 或 **"部署"** 按钮
4. 或者等待自动部署触发

**部署顺序**（Zeabur 自动处理）：
1. ✅ Redis 服务（已启动）
2. ✅ MySQL 服务（如果添加了）
3. 🔄 user-service（依赖 Redis，正在构建）
4. 🔄 post-service（依赖 Redis 和 user-service，等待中）

**构建过程**：
- Zeabur 会从 Git 仓库拉取代码
- 根据 `zeabur.yaml` 配置构建 Docker 镜像
- 注入环境变量
- 启动容器
- 执行健康检查

**预计时间**：
- 首次构建：3-5 分钟
- 后续部署：1-3 分钟

---

### 第八步：查看部署状态

#### 1. 查看服务列表

在项目主页可以看到所有服务：

| 服务名 | 状态 | 端口 | 说明 |
|--------|------|------|------|
| redis | 🟢 Running | 6379 | Redis 缓存服务 |
| mysql | 🟢 Running | 3306 | MySQL 数据库（可选） |
| user-service | 🟢 Running | 8081 | 用户服务 |
| post-service | 🟢 Running | 8082 | 帖子服务 |

#### 2. 查看构建日志

1. 点击具体的服务（如 `user-service`）
2. 点击 **"Logs"** 或 **"日志"** 标签
3. 选择 **"Build Logs"** 或 **"构建日志"**

**关键日志信息**：
```
[+] Building Docker image...
[+] Copying services/common...
[+] Downloading Go dependencies...
[+] Building application...
[+] Image built successfully
```

#### 3. 查看运行日志

在日志页面选择 **"Runtime Logs"** 或 **"运行日志"**：

**user-service 启动日志**：
```
========================================
Service:     User Service
Environment: prod
Database:    /app/dbs/prod/user_prod.db
Port:        8081
========================================
Redis connected successfully
TokenManager initialized successfully
User Service is running on :8081
```

**post-service 启动日志**：
```
========================================
Service:     Post Service
Environment: prod
Database:    /app/dbs/prod/post_prod.db
Port:        8082
========================================
Redis connected successfully
TokenManager initialized successfully
Post Service is running on :8082
```

#### 4. 检查健康状态

Zeabur 会自动执行健康检查（配置在 `zeabur.yaml` 中）：

```yaml
healthCheck:
  http:
    path: /health
    port: 8081
  interval: 30s
  timeout: 10s
  retries: 3
```

**验证方式**：
- 服务状态显示为 🟢 Running
- 没有重启循环
- 健康检查通过

---

### 第九步：获取服务访问地址

#### 1. 查看默认域名

Zeabur 会为每个服务自动分配域名：

1. 点击服务（如 `user-service`）
2. 在服务详情页找到 **"Domains"** 或 **"域名"** 部分
3. 复制默认域名，格式类似：
   - `user-service-xxx.zeabur.app`
   - `post-service-xxx.zeabur.app`

#### 2. 配置自定义域名（可选）

如果有自己的域名：

1. 在域名服务商添加 DNS 记录：
   - 类型：CNAME
   - 名称：api（或其他子域名）
   - 值：`user-service-xxx.zeabur.app`
   
2. 在 Zeabur 服务页面：
   - 点击 **"Add Domain"** 或 **"添加域名"**
   - 输入你的域名：`api.yourdomain.com`
   - 点击 **"Add"**
   
3. 等待 DNS 生效（通常 5-30 分钟）
4. Zeabur 会自动配置 HTTPS 证书

---

## 部署后测试

### 1. 健康检查测试

#### 测试 user-service

```bash
# 替换为你的实际域名
curl https://user-service-xxx.zeabur.app/health
```

**预期响应**：
```json
{
  "status": "healthy",
  "service": "user-service",
  "timestamp": "2025-12-26T14:30:00Z"
}
```

#### 测试 post-service

```bash
curl https://post-service-xxx.zeabur.app/health
```

**预期响应**：
```json
{
  "status": "healthy",
  "service": "post-service",
  "timestamp": "2025-12-26T14:30:00Z"
}
```

---

### 2. 用户注册测试

```bash
curl -X POST https://user-service-xxx.zeabur.app/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456",
    "nickname": "测试用户"
  }'
```

**预期响应**：
```json
{
  "code": 200,
  "message": "注册成功",
  "data": {
    "user_id": 1,
    "username": "testuser",
    "nickname": "测试用户"
  }
}
```

---

### 3. 用户登录测试

```bash
curl -X POST https://user-service-xxx.zeabur.app/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "Test123456"
  }'
```

**预期响应**：
```json
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "access_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9...",
    "expires_in": 86400
  }
}
```

**保存 access_token**，后续请求需要使用。

---

### 4. 创建帖子测试

```bash
# 替换 YOUR_ACCESS_TOKEN 为上一步获取的 token
curl -X POST https://post-service-xxx.zeabur.app/posts \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -d '{
    "content": "这是我的第一条动态！",
    "images": ["https://example.com/image1.jpg"]
  }'
```

**预期响应**：
```json
{
  "code": 200,
  "message": "创建成功",
  "data": {
    "post_id": 1,
    "content": "这是我的第一条动态！",
    "user_id": 1,
    "created_at": "2025-12-26T14:35:00Z"
  }
}
```

---

### 5. 获取帖子列表测试

```bash
curl -X GET https://post-service-xxx.zeabur.app/posts \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "message": "获取成功",
  "data": {
    "posts": [
      {
        "post_id": 1,
        "content": "这是我的第一条动态！",
        "user_id": 1,
        "username": "testuser",
        "nickname": "测试用户",
        "created_at": "2025-12-26T14:35:00Z"
      }
    ],
    "total": 1
  }
}
```

---

### 6. 服务间通信测试

验证 post-service 能否正确调用 user-service：

```bash
# 点赞帖子（会触发 post-service 调用 user-service 验证 token）
curl -X POST https://post-service-xxx.zeabur.app/posts/1/like \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "message": "点赞成功",
  "data": {
    "post_id": 1,
    "likes_count": 1
  }
}
```

**验证点**：
- post-service 成功验证了 JWT token（使用公钥）
- 服务间内部网络通信正常
- Redis 缓存工作正常

---

### 7. Redis 缓存测试

#### 测试 Token 黑名单功能

```bash
# 1. 登出（将 token 加入黑名单）
curl -X POST https://user-service-xxx.zeabur.app/logout \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**预期响应**：
```json
{
  "code": 200,
  "message": "登出成功"
}
```

```bash
# 2. 使用已登出的 token 访问（应该失败）
curl -X GET https://post-service-xxx.zeabur.app/posts \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**预期响应**：
```json
{
  "code": 401,
  "message": "Token 已失效"
}
```

**验证点**：
- Redis 成功存储了黑名单
- user-service 和 post-service 都能访问 Redis
- Token 验证逻辑正确

---

### 8. 数据库持久化测试

#### 测试数据持久化

```bash
# 1. 创建多条帖子
for i in {1..5}; do
  curl -X POST https://post-service-xxx.zeabur.app/posts \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
    -d "{\"content\": \"测试帖子 $i\"}"
done

# 2. 在 Zeabur 控制台重启 post-service

# 3. 重启后再次获取帖子列表
curl -X GET https://post-service-xxx.zeabur.app/posts \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**预期结果**：
- 重启后数据仍然存在
- 帖子数量和内容不变
- 证明 SQLite 数据库持久化成功

---

### 9. 性能和负载测试

#### 使用 Apache Bench 进行简单压测

```bash
# 安装 ab（如果没有）
# macOS: brew install httpd
# Ubuntu: sudo apt-get install apache2-utils

# 压测健康检查接口（100 并发，1000 请求）
ab -n 1000 -c 100 https://user-service-xxx.zeabur.app/health
```

**关注指标**：
- Requests per second（每秒请求数）
- Time per request（平均响应时间）
- Failed requests（失败请求数，应该为 0）

#### 使用 wrk 进行高级压测

```bash
# 安装 wrk
# macOS: brew install wrk
# Ubuntu: sudo apt-get install wrk

# 压测登录接口（10 线程，100 连接，持续 30 秒）
wrk -t10 -c100 -d30s \
  -s post.lua \
  https://user-service-xxx.zeabur.app/login
```

---

### 10. 监控和日志查看

#### 在 Zeabur 控制台查看

1. **实时日志**：
   - 进入服务详情页
   - 点击 **"Logs"** 标签
   - 选择 **"Runtime Logs"**
   - 可以看到实时的应用日志

2. **资源使用情况**：
   - 查看 CPU 使用率
   - 查看内存使用率
   - 查看网络流量

3. **请求统计**：
   - 查看请求数量
   - 查看响应时间
   - 查看错误率

---

## 常见问题排查

### 问题 1：服务无法启动

**症状**：
- 服务状态显示为 "Failed" 或不断重启
- 日志中出现错误信息

**排查步骤**：

1. **检查构建日志**：
   ```
   进入服务 -> Logs -> Build Logs
   ```
   查找构建错误，常见问题：
   - Go 版本不匹配
   - 依赖下载失败
   - 编译错误

2. **检查运行日志**：
   ```
   进入服务 -> Logs -> Runtime Logs
   ```
   查找运行时错误，常见问题：
   - 环境变量未配置
   - 数据库连接失败
   - Redis 连接失败

3. **检查环境变量**：
   ```
   进入服务 -> Variables
   ```
   确认必需的环境变量已配置：
   - `JWT_PRIVATE_KEY`
   - `JWT_PUBLIC_KEY`
   - `DB_DSN`

**解决方案**：
- 如果是环境变量问题，添加缺失的变量后重新部署
- 如果是依赖问题，检查 `go.mod` 和网络连接
- 如果是代码问题，修复后推送代码，触发重新部署

---

### 问题 2：JWT 认证失败

**症状**：
- 登录成功但后续请求返回 401
- 日志显示 "Token 验证失败"

**排查步骤**：

1. **检查 JWT 密钥配置**：
   ```bash
   # 在 Zeabur 控制台检查环境变量
   JWT_PRIVATE_KEY: *** (应该显示为加密)
   JWT_PUBLIC_KEY: *** (应该显示为加密)
   ```

2. **验证密钥格式**：
   - 确保包含完整的 PEM 头尾标记
   - 确保没有多余的空格或换行
   - 确保私钥和公钥是配对的

3. **检查服务日志**：
   ```
   查找 "TokenManager initialized successfully"
   如果没有，说明密钥加载失败
   ```

**解决方案**：
```bash
# 重新获取密钥
cat private.pem  # 完整复制
cat public.pem   # 完整复制

# 在 Zeabur 控制台重新配置环境变量
# 删除旧的 JWT_PRIVATE_KEY 和 JWT_PUBLIC_KEY
# 重新添加，确保粘贴完整内容
```

---

### 问题 3：服务间通信失败

**症状**：
- post-service 无法调用 user-service
- 日志显示 "connection refused" 或 "timeout"

**排查步骤**：

1. **检查服务依赖配置**：
   ```yaml
   # zeabur.yaml 中的配置
   post-service:
     dependsOn:
       - user-service  # 确保配置了依赖
   ```

2. **检查服务 URL 配置**：
   ```bash
   # 在 post-service 的环境变量中
   USER_SERVICE_URL: http://user-service:8081
   # 注意：使用内部服务名，不是外部域名
   ```

3. **验证 user-service 健康状态**：
   ```bash
   # 确保 user-service 正在运行
   curl https://user-service-xxx.zeabur.app/health
   ```

**解决方案**：
- 确保 `dependsOn` 配置正确
- 确保使用内部服务名（如 `http://user-service:8081`）
- 重新部署 post-service

---

### 问题 4：Redis 连接失败

**症状**：
- 日志显示 "Failed to connect to Redis"
- Token 黑名单功能不工作

**排查步骤**：

1. **检查 Redis 服务状态**：
   ```
   在 Zeabur 项目页面查看 Redis 服务
   状态应该是 Running (绿色)
   ```

2. **检查服务依赖**：
   ```yaml
   user-service:
     dependsOn:
       - redis  # 确保配置了依赖
   ```

3. **检查 Redis 配置**：
   ```bash
   # Zeabur 会自动注入这些环境变量
   REDIS_HOST: redis
   REDIS_PORT: 6379
   REDIS_PASSWORD: ***
   ```

**解决方案**：
- 确保 Redis 服务已启动
- 确保 `dependsOn` 配置正确
- 检查代码中的 Redis 连接逻辑

---

### 问题 5：数据库数据丢失

**症状**：
- 服务重启后数据消失
- SQLite 数据库文件不存在

**原因**：
- 容器重启后，未持久化的数据会丢失
- SQLite 文件存储在容器内部，未挂载到持久化存储

**解决方案**：

**方式 1：使用 MySQL（推荐生产环境）**
```bash
# 在 Zeabur 添加 MySQL 服务
# 修改 DB_DSN 环境变量为 MySQL 连接字符串
```

**方式 2：配置持久化存储（SQLite）**
```yaml
# 在 zeabur.yaml 中添加
user-service:
  volumes:
    - name: user-data
      mountPath: /app/dbs
      size: 1Gi
```

---

### 问题 6：HTTPS 证书问题

**症状**：
- 浏览器显示证书不安全
- 自定义域名无法访问

**排查步骤**：

1. **检查域名解析**：
   ```bash
   # 验证 DNS 是否生效
   nslookup api.yourdomain.com
   
   # 应该解析到 Zeabur 的 IP
   ```

2. **检查证书状态**：
   ```
   进入服务 -> Domains
   查看证书状态，应该显示 "Active"
   ```

**解决方案**：
- 等待 DNS 生效（5-30 分钟）
- 在 Zeabur 控制台重新触发证书签发
- 确保域名 CNAME 记录正确

---

### 问题 7：构建超时

**症状**：
- 构建过程卡住
- 显示 "Build timeout"

**原因**：
- 依赖下载慢
- 网络问题
- 构建资源不足

**解决方案**：

1. **优化 Dockerfile**：
   ```dockerfile
   # 添加 Go 模块代理
   ENV GOPROXY=https://goproxy.cn,direct
   ```

2. **使用构建缓存**：
   ```yaml
   # zeabur.yaml 中已配置
   cacheFrom:
     - golang:1.24-alpine
   ```

3. **重新触发构建**：
   ```
   在 Zeabur 控制台点击 "Redeploy"
   ```

---

### 问题 8：环境变量未生效

**症状**：
- 修改环境变量后服务行为未改变
- 日志显示旧的配置值

**原因**：
- 环境变量修改后未重启服务
- 环境变量配置在错误的层级

**解决方案**：

1. **重启服务**：
   ```
   进入服务 -> Settings -> Restart
   或者点击 "Redeploy"
   ```

2. **检查配置层级**：
   ```
   项目级别环境变量：所有服务共享
   服务级别环境变量：仅该服务使用
   服务级别会覆盖项目级别
   ```

---

## 总结

### 部署检查清单

- [ ] 代码已推送到 Git 仓库
- [ ] JWT 密钥文件存在（private.pem、public.pem）
- [ ] Zeabur 项目已创建
- [ ] Redis 服务已部署并运行
- [ ] （可选）MySQL 服务已部署并运行
- [ ] Git 仓库已导入
- [ ] 环境变量已配置（JWT_PRIVATE_KEY、JWT_PUBLIC_KEY）
- [ ] 服务已成功部署
- [ ] 健康检查通过
- [ ] 用户注册/登录测试通过
- [ ] 帖子创建/获取测试通过
- [ ] 服务间通信测试通过

### 关键配置文件

| 文件 | 作用 | 是否必需 |
|------|------|----------|
| `zeabur.yaml` | Zeabur 主配置 | ✅ 必需 |
| `zeabur-env-example.yaml` | 环境变量说明 | ℹ️ 参考 |
| `services/*/Dockerfile` | Docker 镜像构建 | ✅ 必需 |
| `docker-compose.build.yml` | 本地构建配置 | ℹ️ 本地使用 |
| `private.pem` | JWT 私钥 | ✅ 必需 |
| `public.pem` | JWT 公钥 | ✅ 必需 |

### 下一步

1. **监控和优化**：
   - 定期查看服务日志
   - 监控资源使用情况
   - 根据负载调整资源配置

2. **安全加固**：
   - 定期轮换 JWT 密钥
   - 配置 CORS 策略
   - 启用速率限制

3. **功能扩展**：
   - 添加更多 API 功能
   - 集成第三方服务
   - 实现数据备份

---

**文档版本**：1.0  
**最后更新**：2025-12-26  
**维护者**：WeCircle Team
