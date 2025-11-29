# 多环境配置指南（hosts + Nginx 反向代理方案）

本文档详细介绍如何使用 **hosts 文件 + Nginx 反向代理** 的方式，在本地搭建多个环境（staging、pre、prod），实现通过不同域名访问不同环境的效果。
简单总结流程就是：
- 本地Nginx配置负责将域名转发到127.0.0.1: 8082， 从而请求打到服务端
- 项目中配置的config下的不同环境，是为了在运行Docker时设置不同的测试环境
---

## 目录

1. [方案概述](#1-方案概述)
2. [准备工作](#2-准备工作)
3. [配置 hosts 文件](#3-配置-hosts-文件)
4. [安装 Nginx](#4-安装-nginx)
5. [配置 Nginx 反向代理](#5-配置-nginx-反向代理)
6. [创建环境配置文件](#6-创建环境配置文件)
7. [启动 Docker 容器](#7-启动-docker-容器)
8. [验证配置](#8-验证配置)
9. [使用 Docker Compose（可选）](#9-使用-docker-compose可选)
10. [常见问题排查](#10-常见问题排查)
11. [总结](#11-总结)

---

## 1. 方案概述

### 1.1 什么是多环境？

在实际开发中，我们通常需要多个环境：

- **Staging（测试环境）**：测试团队使用
- **Pre（预发布环境）**：上线前最后验证
- **Production（生产环境）**：正式对外服务

### 1.2 为什么使用 hosts + Nginx 方案？

**传统方案（端口区分）**：
```
http://127.0.0.1:8081  (staging)
http://127.0.0.1:8082  (pre)
http://127.0.0.1:8083  (prod)
```

**hosts + Nginx 方案（域名区分）**：
```
http://api.staging.myapp.com   (staging)
http://api.pre.myapp.com       (pre)
http://api.prod.myapp.com      (prod)
```

**优势**：
- ✅ 更接近生产环境（生产环境使用域名）
- ✅ 前端配置简单（统一域名格式）
- ✅ 易于记忆和管理
- ✅ 支持 HTTPS 配置
- ✅ 团队协作友好

### 1.3 整体架构

```
浏览器/客户端
    ↓
http://api.staging.myapp.com (hosts 解析到 127.0.0.1)
    ↓
Nginx (监听 80 端口，根据域名转发)
    ↓
    ├─ api.staging.myapp.com  → 127.0.0.1:8081 (Docker 容器)
    ├─ api.pre.myapp.com      → 127.0.0.1:8082 (Docker 容器)
    └─ api.prod.myapp.com     → 127.0.0.1:8083 (Docker 容器)
```

---

## 2. 准备工作

### 2.1 确认已安装 Docker

```bash
# 检查 Docker 是否安装
docker --version

# 应该输出类似：Docker version 24.0.0, build xxx
```

如果未安装，请参考 [Docker_Build.md](./Docker_Build.md) 文档。

### 2.2 确认已构建镜像

```bash
# 查看镜像
docker images | grep server_test_demo

# 应该看到：
# server_test_demo_1.0.0:latest   或
# chomay/server_test_demo:1.0.0
```

如果没有镜像，请先构建：

```bash
docker build -t server_test_demo_1.0.0 .
```

---

## 3. 配置 hosts 文件

### 3.1 什么是 hosts 文件？

`hosts` 文件是操作系统用于**域名解析**的本地配置文件。它的优先级**高于 DNS 服务器**。

**作用**：将域名映射到 IP 地址。

### 3.2 编辑 hosts 文件

#### macOS/Linux

```bash
# 使用 vim 编辑（需要管理员权限）
sudo vim /etc/hosts
```

#### Windows

```
1. 以管理员身份运行记事本
2. 打开文件：C:\Windows\System32\drivers\etc\hosts
```

### 3.3 添加域名映射

在 `hosts` 文件末尾添加以下内容：

```
# Test Demo Project - Multi Environment Setup
127.0.0.1  api.staging.myapp.com
127.0.0.1  api.pre.myapp.com
127.0.0.1  api.prod.myapp.com
```

**说明**：
- `127.0.0.1`：本地回环地址（localhost）
- `api.staging.myapp.com`：自定义的域名（可以随意命名）
- 所有域名都指向本地

### 3.4 保存并退出

**vim 操作**：
1. 按 `i` 进入编辑模式
2. 粘贴上面的内容
3. 按 `Esc` 退出编辑模式
4. 输入 `:wq` 保存并退出

### 3.5 清除 DNS 缓存（可选但推荐）

#### macOS

```bash
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

#### Windows

```cmd
ipconfig /flushdns
```

#### Linux

```bash
sudo systemd-resolve --flush-caches
```

### 3.6 验证 hosts 配置

```bash
# 使用 ping 测试
ping -c 3 api.staging.myapp.com

# 应该输出：
# PING api.staging.myapp.com (127.0.0.1): 56 data bytes
# 64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.051 ms
```

**注意**：
- ✅ `ping` 会查询 hosts 文件，能正常解析
- ❌ `nslookup` 不查询 hosts 文件，会显示找不到域名（这是正常的）

### 3.7 为什么 ping 成功但 nslookup 失败？（重要）

这是一个**非常常见的困惑**，让我详细解释：

#### 问题现象

```bash
# ping 成功 ✅
$ ping -c 3 api.staging.myapp.com
PING api.staging.myapp.com (127.0.0.1): 56 data bytes
64 bytes from 127.0.0.1: icmp_seq=0 ttl=64 time=0.051 ms

# nslookup 失败 ❌
$ nslookup api.staging.myapp.com
Server:         192.168.20.70
** server can't find api.staging.myapp.com: NXDOMAIN
```

#### 原因解释

**这是完全正常的！** 原因是这两个命令的域名解析方式不同：

| 命令 | 解析顺序 | 是否查询 hosts 文件 | 用途 |
|------|---------|-------------------|------|
| **ping** | 1. hosts 文件<br>2. DNS 服务器 | ✅ 是 | 测试网络连通性 |
| **nslookup** | 1. DNS 服务器 | ❌ 否（故意跳过） | 测试 DNS 服务器 |
| **curl** | 1. hosts 文件<br>2. DNS 服务器 | ✅ 是 | 实际应用访问 |
| **浏览器** | 1. hosts 文件<br>2. DNS 服务器 | ✅ 是 | 实际应用访问 |

#### 为什么 nslookup 不查询 hosts 文件？

`nslookup` 是一个 **DNS 调试工具**，它的设计目的是：
- 🔍 测试 DNS 服务器是否正常工作
- 🔍 查看 DNS 记录（A、CNAME、MX 等）
- 🔍 诊断 DNS 解析问题

如果 `nslookup` 也查询 hosts 文件，就无法测试真实的 DNS 服务器了！

#### 类比说明

```
hosts 文件 = 你的手机通讯录
DNS 服务器 = 114 查号台

ping 命令：
  "我要找张三的电话"
  → 先查通讯录（hosts）✅
  → 通讯录没有，再打 114（DNS）

nslookup 命令：
  "我要测试 114 查号台是否正常"
  → 直接打 114（DNS）✅
  → 不查通讯录（否则测不出 114 的问题）
```

#### 正确的验证方法

**验证 hosts 配置是否生效**，应该使用：

```bash
# 方法1：ping（推荐）
ping -c 3 api.staging.myapp.com

# 方法2：curl
curl http://api.staging.myapp.com

# 方法3：dscacheutil（macOS 专用）
dscacheutil -q host -a name api.staging.myapp.com
```

**不要使用** `nslookup` 或 `dig` 来验证 hosts 文件！

#### 总结

- ✅ **ping 成功** = hosts 配置正确
- ❌ **nslookup 失败** = 正常现象（DNS 服务器没有这个域名）
- 🎯 **实际应用**（浏览器、curl）会正常工作

---

## 4. 安装 Nginx

### 4.1 什么是 Nginx？

Nginx 是一个高性能的 **HTTP 服务器** 和 **反向代理服务器**。

**反向代理**：客户端访问 Nginx，Nginx 将请求转发到后端服务。

```
客户端 → Nginx (80端口) → 后端服务 (8081/8082/8083/8084端口)
```

### 4.2 安装 Nginx

#### macOS

```bash
# 使用 Homebrew 安装
brew install nginx

# 验证安装
nginx -v
# 输出：nginx version: nginx/1.25.x
```

#### Ubuntu/Debian

```bash
sudo apt update
sudo apt install nginx

# 验证安装
nginx -v
```

#### CentOS/RHEL

```bash
sudo yum install nginx

# 验证安装
nginx -v
```

### 4.3 Nginx 重要目录说明

**⚠️ 重要提示**：macOS 上通过 Homebrew 安装的 Nginx，路径会根据 Mac 的芯片类型不同而不同！

#### 如何判断您的 Mac 类型？

```bash
# 查看 Mac 芯片类型
uname -m

# 输出结果：
# x86_64  → Intel Mac
# arm64   → Apple Silicon Mac (M1/M2/M3 等)
```

或者：

```bash
# 查看 Homebrew 安装路径
brew --prefix

# 输出结果：
# /usr/local      → Intel Mac
# /opt/homebrew   → Apple Silicon Mac
```

#### macOS - Intel Mac (x86_64)

```
配置文件主目录：/usr/local/etc/nginx/
主配置文件：    /usr/local/etc/nginx/nginx.conf
自定义配置目录：/usr/local/etc/nginx/servers/
日志目录：      /usr/local/var/log/nginx/
Nginx 二进制：  /usr/local/bin/nginx
```

#### macOS - Apple Silicon Mac (ARM64) ⭐

```
配置文件主目录：/opt/homebrew/etc/nginx/
主配置文件：    /opt/homebrew/etc/nginx/nginx.conf
自定义配置目录：/opt/homebrew/etc/nginx/servers/
日志目录：      /opt/homebrew/var/log/nginx/
Nginx 二进制：  /opt/homebrew/bin/nginx
```

#### 路径对比表

| 项目 | Intel Mac | Apple Silicon Mac |
|------|-----------|-------------------|
| **配置目录** | `/usr/local/etc/nginx/` | `/opt/homebrew/etc/nginx/` |
| **自定义配置** | `/usr/local/etc/nginx/servers/` | `/opt/homebrew/etc/nginx/servers/` |
| **日志目录** | `/usr/local/var/log/nginx/` | `/opt/homebrew/var/log/nginx/` |
| **Nginx 命令** | `/usr/local/bin/nginx` | `/opt/homebrew/bin/nginx` |

**本文档后续示例**：
- 如果您使用 **Intel Mac**，请使用 `/usr/local/` 路径
- 如果您使用 **Apple Silicon Mac**，请使用 `/opt/homebrew/` 路径

#### Linux

```
配置文件主目录：/etc/nginx/
主配置文件：    /etc/nginx/nginx.conf
自定义配置目录：/etc/nginx/conf.d/
日志目录：      /var/log/nginx/
```

---

## 5. 配置 Nginx 反向代理

### 5.1 创建配置文件

#### 方式1：使用项目提供的模板文件（推荐）

项目中已经为您准备好了配置文件模板：`nginx/server_test_demo_arm64.conf`（适用于 Apple Silicon Mac）

**步骤**：

1. 首先确认您的 Mac 类型：

```bash
# 查看 Homebrew 路径
brew --prefix

# 输出 /opt/homebrew   → Apple Silicon Mac
# 输出 /usr/local      → Intel Mac
```

2. 复制配置文件到 Nginx 目录：

**Apple Silicon Mac (ARM64)**：
```bash
# 复制项目下的配置文件到 系统的 Nginx 目录：
sudo cp nginx/server_test_demo_arm64.conf /opt/homebrew/etc/nginx/servers/server_test_demo.conf

# 验证文件是否复制成功
ls -lh /opt/homebrew/etc/nginx/servers/server_test_demo.conf
```

**Intel Mac (x86_64)**：
```bash
# 复制配置文件
sudo cp nginx/server_test_demo.conf /usr/local/etc/nginx/servers/server_test_demo.conf

# 验证文件是否复制成功
ls -lh /usr/local/etc/nginx/servers/server_test_demo.conf
```

#### 方式2：手动创建配置文件

如果您想手动创建配置文件，请根据您的 Mac 类型选择对应的路径：

**Apple Silicon Mac (ARM64)**：
```bash
# 创建配置文件
sudo vim /opt/homebrew/etc/nginx/servers/server_test_demo.conf
```

**Intel Mac (x86_64)**：
```bash
# 创建配置文件
sudo vim /usr/local/etc/nginx/servers/server_test_demo.conf
```

**Linux**：
```bash
# 创建配置文件
sudo vim /etc/nginx/conf.d/server_test_demo.conf
```

### 5.2 配置文件内容

**⚠️ 重要提示**：如果您使用的是 **方式1（推荐）**，配置文件已经准备好，可以跳过此步骤直接到 5.3 节。

如果您选择手动创建配置文件，请根据您的系统类型选择对应的配置内容：

#### Apple Silicon Mac (ARM64) 配置

将以下内容复制到配置文件中（日志路径使用 `/opt/homebrew/`）：

```nginx
# Staging 环境
server {
    listen 80;
    server_name api.staging.myapp.com;
    
    access_log /opt/homebrew/var/log/nginx/server_test_demo_staging_access.log;
    error_log /opt/homebrew/var/log/nginx/server_test_demo_staging_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# Pre 环境
server {
    listen 80;
    server_name api.pre.myapp.com;
    
    access_log /opt/homebrew/var/log/nginx/server_test_demo_pre_access.log;
    error_log /opt/homebrew/var/log/nginx/server_test_demo_pre_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# Production 环境
server {
    listen 80;
    server_name api.prod.myapp.com;
    
    access_log /opt/homebrew/var/log/nginx/server_test_demo_prod_access.log;
    error_log /opt/homebrew/var/log/nginx/server_test_demo_prod_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8084;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

#### Intel Mac (x86_64) 配置

将以下内容复制到配置文件中（日志路径使用 `/usr/local/`）：

```nginx
# Staging 环境
server {
    listen 80;
    server_name api.staging.myapp.com;
    
    access_log /usr/local/var/log/nginx/server_test_demo_staging_access.log;
    error_log /usr/local/var/log/nginx/server_test_demo_staging_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# Pre 环境
server {
    listen 80;
    server_name api.pre.myapp.com;
    
    access_log /usr/local/var/log/nginx/server_test_demo_pre_access.log;
    error_log /usr/local/var/log/nginx/server_test_demo_pre_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# Production 环境
server {
    listen 80;
    server_name api.prod.myapp.com;
    
    access_log /usr/local/var/log/nginx/server_test_demo_prod_access.log;
    error_log /usr/local/var/log/nginx/server_test_demo_prod_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8084;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

#### Linux 配置

将以下内容复制到配置文件中（日志路径使用 `/var/log/nginx/`）：

```nginx
# Staging 环境
server {
    listen 80;
    server_name api.staging.myapp.com;
    
    access_log /var/log/nginx/server_test_demo_staging_access.log;
    error_log /var/log/nginx/server_test_demo_staging_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8082;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# Pre 环境
server {
    listen 80;
    server_name api.pre.myapp.com;
    
    access_log /var/log/nginx/server_test_demo_pre_access.log;
    error_log /var/log/nginx/server_test_demo_pre_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8083;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}

# Production 环境
server {
    listen 80;
    server_name api.prod.myapp.com;
    
    access_log /var/log/nginx/server_test_demo_prod_access.log;
    error_log /var/log/nginx/server_test_demo_prod_error.log;
    
    location / {
        proxy_pass http://127.0.0.1:8084;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        
        proxy_connect_timeout 60s;
        proxy_send_timeout 60s;
        proxy_read_timeout 60s;
    }
}
```

**配置说明**：

- `listen 80`：监听 80 端口（HTTP 默认端口）
- `server_name`：匹配的域名
- `proxy_pass`：转发到的后端地址
- `proxy_set_header`：传递请求头给后端服务

**Linux 用户注意**：将日志路径改为 `/var/log/nginx/`

### 5.3 测试配置文件

```bash
# 测试配置文件语法
sudo nginx -t

# 应该输出：
# nginx: the configuration file /usr/local/etc/nginx/nginx.conf syntax is ok
# nginx: configuration file /usr/local/etc/nginx/nginx.conf test is successful
```

如果有错误，请检查：
- 配置文件路径是否正确
- 语法是否有误（分号、大括号等）

### 5.4 启动 Nginx

#### macOS

```bash
# 启动 Nginx
sudo nginx

# 或者重新加载配置（如果已经启动）
sudo nginx -s reload

# 停止 Nginx
sudo nginx -s stop

# 立即停止
sudo nginx -s quit

# 强制杀死进程
sudo pkill nginx
```

#### Linux

```bash
# 启动 Nginx
sudo systemctl start nginx

# 设置开机自启
sudo systemctl enable nginx

# 重新加载配置
sudo systemctl reload nginx
```

### 5.5 验证 Nginx 状态

```bash
# 检查 Nginx 是否运行
ps aux | grep nginx

# 应该看到类似输出：
# nginx: master process nginx
# nginx: worker process
```

或者访问：

```bash
curl http://localhost

# 如果看到 Nginx 欢迎页面，说明启动成功
```

---

## 6. 创建环境配置文件

### 6.1 创建配置目录

```bash
# 在项目根目录下创建 config 目录
cd /Users/chomay/Desktop/go/1_test_demo
mkdir -p config
```

### 6.2 创建各环境配置文件

#### config/.env.staging

```bash
cat > config/.env.staging << 'EOF'
# Staging 环境配置
ENV=staging
SERVER_PORT=8081
DB_DSN=dbs/test_staging.db
LOG_LEVEL=info
EOF
```

#### config/.env.pre

```bash
cat > config/.env.pre << 'EOF'
# Pre 环境配置
ENV=pre
SERVER_PORT=8081
DB_DSN=dbs/test_pre.db
LOG_LEVEL=warn
EOF
```

#### config/.env.prod

```bash
cat > config/.env.prod << 'EOF'
# Production 环境配置
ENV=production
SERVER_PORT=8081
DB_DSN=dbs/test_prod.db
LOG_LEVEL=error
EOF
```

### 6.3 查看配置文件

```bash
# 查看所有配置文件
ls -la config/

# 应该看到：
# .env.staging
# .env.pre
# .env.prod
```

---

## 7. 启动 Docker 容器

### 7.1 创建数据库目录

```bash
# 为每个环境创建独立的数据库目录
mkdir -p dbs/staging
mkdir -p dbs/pre
mkdir -p dbs/prod
```

### 7.2 启动各环境容器

#### Staging 环境（8081 端口）

```bash
docker run -d \
  --name test_demo_staging \
  -p 8082:8081 \
  --env-file config/.env.staging \
  -v $(pwd)/dbs/staging:/app/dbs \
  test_demo_1.0.0
```

#### Pre 环境（8083 端口）

```bash
docker run -d \
  --name test_demo_pre \
  -p 8083:8081 \
  --env-file config/.env.pre \
  -v $(pwd)/dbs/pre:/app/dbs \
  test_demo_1.0.0
```

#### Production 环境（8084 端口）

```bash
docker run -d \
  --name test_demo_prod \
  -p 8084:8081 \
  --env-file config/.env.prod \
  -v $(pwd)/dbs/prod:/app/dbs \
  test_demo_1.0.0
```

**参数说明**：

- `-d`：后台运行
- `--name`：容器名称
- `-p 8082:8081`：端口映射（宿主机:容器）
- `--env-file`：加载环境变量文件
- `-v`：挂载数据卷（数据持久化）

`-p 8082:8081`进一步解释：
```
8082: 是宿主机的端口， 外部用户通过127.0.0.1:8082访问到了服务器
8081: 是容器内的端口， 是Docker容器内的端口
main.go 的 8081：决定了程序在容器里监听哪个端口
Docker 的 :8081：决定了 Docker 把流量转发到容器里的哪个端口。

Docker中一般会运行换一个Go程序，但程序中也可以监听多个端口， 8081 端口：提供 API 服务（给用户用）。9090 端口：提供 Metrics 监控数据（给 Prometheus 用）。这时候，Docker 就需要知道您想映射哪一个：
# 映射 API 端口
docker run -p 8082:8081 ...
# 映射监控端口
docker run -p 9091:9090 ...
# 或者同时映射
docker run -p 8082:8081 -p 9091:9090 ...

所以，Docker 必须明确知道："您想把外部流量转发到容器里的哪一个端口？"

比如
通过127.0.0.1: 8082 进行注册请求， 则需要对应到容器内的8081端口上。
通过127.0.0.1: 9000 上传数据监控数据，则需要对应到容器内的9000端口上

简单就是： 外部请求接口了，容器内部需要知道将数据转发到哪个端口上处理。
```

项目中测试：
```
// 仅当前命令命令生效
ENV=staging SERVER_PORT=8082 DB_DSN=dbs/staging/test_staging.db go run main.go
```

模拟请求：
```
// 1. 注册
curl -X POST http://api.staging.myapp.com/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","phone":"13800001111","password":"123456"}'

// 2. 登录
curl -X POST http://api.staging.myapp.com/login \
  -H "Content-Type: application/json" \
  -d '{"phone":"13800001111","password":"123456"}'

// 3. 获取用户信息
curl -X GET http://api.staging.myapp.com/users/13800001111 \
  -H "Authorization: Bearer <token>"

// 4. 获取所有用户
curl -X GET http://api.staging.myapp.com/users \
  -H "Authorization: Bearer <token>"
```
### 7.3 查看运行状态

```bash
# 查看所有容器
docker ps

# 应该看到 4 个容器在运行：
# test_demo_staging   (0.0.0.0:8081->8081/tcp)
# test_demo_staging   (0.0.0.0:8082->8081/tcp)
# test_demo_pre       (0.0.0.0:8083->8081/tcp)
# test_demo_prod      (0.0.0.0:8084->8081/tcp)
```

### 7.4 查看容器日志

```bash
# 查看 staging 环境日志
docker logs test_demo_staging

# 实时查看日志
docker logs -f test_demo_staging
```

---

## 8. 验证配置

### 8.1 测试端口访问（绕过 Nginx）

```bash
# 测试 Staging 环境（直接访问端口）
curl http://127.0.0.1:8081/users

# 测试 Pre 环境
curl http://127.0.0.1:8082/users

# 测试 Production 环境
curl http://127.0.0.1:8083/users
```

如果返回 JSON 数据或空数组 `[]`，说明容器运行正常。

### 8.2 测试域名访问（通过 Nginx）

```bash
# 测试 Dev 环境
curl http://api.staging.myapp.com/users

# 测试 Staging 环境
curl http://api.staging.myapp.com/users

# 测试 Pre 环境
curl http://api.pre.myapp.com/users

# 测试 Production 环境
curl http://api.prod.myapp.com/users
```

### 8.3 浏览器访问

打开浏览器，访问以下地址：

```
http://api.staging.myapp.com/swagger/index.html
http://api.pre.myapp.com/swagger/index.html
http://api.prod.myapp.com/swagger/index.html
```

应该能看到 Swagger API 文档页面。

### 8.4 测试注册接口

```bash
# 在 Staging 环境注册用户
curl -X POST http://api.staging.myapp.com/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test_staging",
    "phone": "13800138001",
    "password": "123456"
  }'

# 查询用户
curl http://api.staging.myapp.com/users
```

### 8.5 验证数据隔离

```bash
# 查看各环境的数据库文件
ls -lh dbs/staging/
ls -lh dbs/staging/
ls -lh dbs/pre/
ls -lh dbs/prod/

# 每个环境应该有独立的数据库文件
```

---

## 9. 使用 Docker Compose（可选）

### 9.1 创建 docker-compose.yml

在项目根目录创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  # Staging 环境
  staging:
    build: .
    container_name: test_demo_staging
    ports:
      - "8082:8081"
    env_file:
      - config/.env.staging
    volumes:
      - ./dbs/staging:/app/dbs
    restart: unless-stopped

  # Pre 环境
  pre:
    build: .
    container_name: test_demo_pre
    ports:
      - "8083:8081"
    env_file:
      - config/.env.pre
    volumes:
      - ./dbs/pre:/app/dbs
    restart: unless-stopped

  # Production 环境
  prod:
    build: .
    container_name: test_demo_prod
    ports:
      - "8084:8081"
    env_file:
      - config/.env.prod
    volumes:
      - ./dbs/prod:/app/dbs
    restart: unless-stopped
```

### 9.2 使用 Docker Compose

```bash
# 启动所有环境
docker-compose up -d

# 启动单个环境
docker-compose up -d staging

# 查看运行状态
docker-compose ps

# 查看日志
docker-compose logs -f staging

# 停止所有环境
docker-compose down

# 重启某个环境
docker-compose restart staging
```

---

## 10. 常见问题排查

### 10.1 域名无法解析

**问题**：`ping api.staging.myapp.com` 失败

**解决方案**：

```bash
# 1. 检查 hosts 文件
cat /etc/hosts | grep "myapp.com"

# 2. 确保格式正确（IP 和域名之间有空格）
127.0.0.1  api.staging.myapp.com

# 3. 清除 DNS 缓存
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

### 10.2 Nginx 配置错误

**问题**：`nginx -t` 报错

**解决方案**：

```bash
# 查看详细错误信息
sudo nginx -t

# 常见错误：
# 1. 缺少分号
# 2. 大括号不匹配
# 3. 路径错误

# 检查配置文件语法
cat /usr/local/etc/nginx/servers/testdemo.conf
```

### 10.3 端口被占用

**问题**：启动容器时提示端口已被占用

**解决方案**：

```bash
# 查看端口占用
lsof -i :8082

# 停止占用端口的进程
kill -9 <PID>

# 或者使用不同的端口
docker run -p 8085:8081 ...
```

### 10.4 容器无法启动

**问题**：`docker ps` 看不到容器

**解决方案**：

```bash
# 查看所有容器（包括停止的）
docker ps -a

# 查看容器日志
docker logs test_demo_staging

# 常见原因：
# 1. 环境变量文件路径错误
# 2. 数据卷挂载路径不存在
# 3. 镜像不存在
```

### 10.5 Nginx 无法访问后端

**问题**：访问域名返回 502 Bad Gateway

**解决方案**：

```bash
# 1. 检查容器是否运行
docker ps

# 2. 检查端口是否正确
curl http://127.0.0.1:8082/users

# 3. 查看 Nginx 错误日志
tail -f /usr/local/var/log/nginx/testdemo_staging_error.log

# 4. 重启 Nginx
sudo nginx -s reload
```

### 10.6 浏览器缓存问题

**问题**：修改配置后浏览器访问还是旧的

**解决方案**：

```bash
# Chrome 清除 DNS 缓存
# 访问：chrome://net-internals/#dns
# 点击 "Clear host cache"

# 或者使用隐私模式
# Cmd + Shift + N (macOS)
# Ctrl + Shift + N (Windows)
```

---

## 11. 总结

### 11.1 完整架构图

```
┌─────────────────────────────────────────────────────┐
│  浏览器/客户端                                        │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│  hosts 文件                                          │
│  127.0.0.1  api.staging.myapp.com                   │
│  127.0.0.1  api.pre.myapp.com                       │
│  127.0.0.1  api.prod.myapp.com                      │
└─────────────────────────────────────────────────────┘
                      ↓
┌─────────────────────────────────────────────────────┐
│  Nginx (80 端口)                                     │
│  根据 server_name 转发请求                           │
└─────────────────────────────────────────────────────┘
         ↓           ↓           ↓
    ┌────────┐  ┌────────┐  ┌────────┐
    │ :8081  │  │ :8082  │  │ :8083  │
    └────────┘  └────────┘  └────────┘
         ↓           ↓           ↓
    ┌────────┐  ┌────────┐  ┌────────┐
    │Staging │  │  Pre   │  │  Prod  │
    │ 容器   │  │ 容器   │  │ 容器   │
    └────────┘  └────────┘  └────────┘
```

### 11.2 访问地址汇总

| 环境 | 域名访问 | 端口访问 | Swagger 文档 |
|------|---------|---------|-------------|
| Staging | http://api.staging.myapp.com | http://127.0.0.1:8081 | http://api.staging.myapp.com/swagger/index.html |
| Pre | http://api.pre.myapp.com | http://127.0.0.1:8082 | http://api.pre.myapp.com/swagger/index.html |
| Prod | http://api.prod.myapp.com | http://127.0.0.1:8083 | http://api.prod.myapp.com/swagger/index.html |

### 11.3 常用命令汇总

```bash
# ===== hosts 相关 =====
# 编辑 hosts
sudo vim /etc/hosts

# 清除 DNS 缓存
sudo dscacheutil -flushcache && sudo killall -HUP mDNSResponder

# 验证域名解析
ping -c 3 api.staging.myapp.com

# ===== Nginx 相关 =====
# 测试配置
sudo nginx -t

# 启动 Nginx
sudo nginx

# 重新加载配置
sudo nginx -s reload

# 停止 Nginx
sudo nginx -s stop

# 查看错误日志
tail -f /usr/local/var/log/nginx/testdemo_staging_error.log

# ===== Docker 相关 =====
# 查看运行的容器
docker ps

# 启动容器
docker start test_demo_staging

# 停止容器
docker stop test_demo_staging

# 查看日志
docker logs -f test_demo_staging

# 重启容器
docker restart test_demo_staging

# 删除容器
docker rm test_demo_staging

# ===== Docker Compose 相关 =====
# 启动所有环境
docker-compose up -d

# 启动单个环境
docker-compose up -d staging

# 停止所有环境
docker-compose down

# 查看状态
docker-compose ps

# 查看日志
docker-compose logs -f staging
```

### 11.4 下一步

完成多环境配置后，您可以：

1. **前端对接**：前端可以通过配置环境变量切换不同环境的 API 地址
2. **CI/CD 集成**：将部署流程自动化
3. **HTTPS 配置**：为本地环境配置 SSL 证书
4. **监控告警**：添加日志监控和告警机制

---

## 附录

### A. 一键配置脚本

创建 `scripts/setup-multi-env.sh`：

```bash
#!/bin/bash

echo "================================"
echo "Multi-Environment Setup Script"
echo "================================"
echo ""

# 1. 配置 hosts
echo "Step 1: Configuring hosts file..."
if ! grep -q "api.staging.myapp.com" /etc/hosts; then
    sudo bash -c 'cat >> /etc/hosts << EOF

# Test Demo Project - Multi Environment
127.0.0.1  api.dev.myapp.com
127.0.0.1  api.staging.myapp.com
127.0.0.1  api.pre.myapp.com
127.0.0.1  api.prod.myapp.com
EOF'
    echo "✅ Hosts configured"
else
    echo "✅ Hosts already configured"
fi

# 2. 清除 DNS 缓存
echo ""
echo "Step 2: Flushing DNS cache..."
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
echo "✅ DNS cache cleared"

# 3. 创建配置目录
echo ""
echo "Step 3: Creating config directories..."
mkdir -p config
mkdir -p dbs/{dev,staging,pre,prod}
echo "✅ Directories created"

# 4. 启动 Docker 容器
echo ""
echo "Step 4: Starting Docker containers..."
docker-compose up -d
echo "✅ Containers started"

# 5. 重新加载 Nginx
echo ""
echo "Step 5: Reloading Nginx..."
sudo nginx -s reload 2>/dev/null || sudo nginx
echo "✅ Nginx reloaded"

# 6. 验证
echo ""
echo "Step 6: Verifying setup..."
sleep 3

for env in staging pre prod; do
    if curl -s http://api.$env.myapp.com/users > /dev/null; then
        echo "✅ $env environment is running"
    else
        echo "❌ $env environment failed"
    fi
done

echo ""
echo "================================"
echo "Setup Complete!"
echo "================================"
echo ""
echo "Access your environments:"
echo "  Staging: http://api.staging.myapp.com"
echo "  Pre:     http://api.pre.myapp.com"
echo "  Prod:    http://api.prod.myapp.com"
echo ""
echo "Swagger Documentation:"
echo "  http://api.staging.myapp.com/swagger/index.html"
```

使用：

```bash
chmod +x scripts/setup-multi-env.sh
./scripts/setup-multi-env.sh
```

### B. 卸载脚本

创建 `scripts/cleanup-multi-env.sh`：

```bash
#!/bin/bash

echo "Cleaning up multi-environment setup..."

# 停止并删除容器
docker-compose down

# 删除数据库文件（可选）
# rm -rf dbs/staging dbs/pre dbs/prod

# 从 hosts 文件移除配置（需要手动）
echo ""
echo "Please manually remove the following lines from /etc/hosts:"
echo "127.0.0.1  api.staging.myapp.com"
echo "127.0.0.1  api.pre.myapp.com"
echo "127.0.0.1  api.prod.myapp.com"

echo ""
echo "Cleanup complete!"
```
