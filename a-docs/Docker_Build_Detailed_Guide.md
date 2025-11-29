# Docker 镜像构建详细指南

本文档详细介绍本项目（微服务架构）的Docker镜像构建方法，包括单服务构建和多服务构建，并对各个参数进行详细说明。

## 1. 项目Docker架构概述

本项目采用微服务架构，包含两个主要服务：
- **user-service**: 用户服务，提供用户管理功能
- **post-service**: 动态服务，提供内容发布功能

每个服务都有独立的Dockerfile，同时通过docker-compose实现多服务的协同构建和运行。

## 2. 单服务Docker镜像构建

### 2.1 用户服务镜像构建

**命令格式：**
```bash
docker build -t [镜像名称:标签] -f [Dockerfile路径] [构建上下文路径]
```

**示例命令：**
```bash
# 在项目根目录执行
docker build -t user-service:latest -f services/user-service/Dockerfile services/user-service

# 或直接在服务目录下执行
cd services/user-service
docker build -t user-service:latest .
```

**参数说明：**
- `-t user-service:latest`: 指定镜像名称为`user-service`，标签为`latest`
  - `镜像名称`: 建议使用服务名称，便于识别
  - `标签`: 可使用版本号(如`1.0.0`)或环境标识(如`staging`、`pre`、`prod`)
- `-f services/user-service/Dockerfile`: 指定Dockerfile路径（如果不在当前目录）
- `services/user-service`: 指定构建上下文，Docker将在此目录下查找文件

### 2.2 动态服务镜像构建

**示例命令：**
```bash
# 在项目根目录执行
docker build -t post-service:latest -f services/post-service/Dockerfile services/post-service

# 或直接在服务目录下执行
cd services/post-service
docker build -t post-service:latest .
```

### 2.3 自定义标签构建

可以为镜像添加特定版本或环境标签：

```bash
# 构建带版本号的镜像
docker build -t user-service:1.0.0 services/user-service

# 构建带环境标识的镜像
docker build -t user-service:staging services/user-service
docker build -t user-service:pre services/user-service
docker build -t user-service:prod services/user-service
```

## 3. 多服务Docker镜像构建（推荐）

使用docker-compose可以同时构建所有服务，是最便捷的方式。

### 3.1 构建镜像（使用构建配置文件）

**命令：**
```bash
# 使用docker-compose.build.yml构建镜像（不区分环境）
docker-compose -f docker-compose.build.yml build

# 指定版本号构建
IMAGE_TAG=v1.0.0 docker-compose -f docker-compose.build.yml build
```

**参数说明：**
- `-f docker-compose.build.yml`: 指定使用的docker-compose构建配置文件
- `build`: 构建命令，会构建配置文件中定义的所有服务
- `IMAGE_TAG`: 可选，指定镜像版本号，默认使用 latest

### 3.2 指定环境构建

项目提供了多个环境的docker-compose配置文件：

```bash
# 构建staging环境镜像
docker-compose -f docker-compose-staging.yml build

# 构建pre环境镜像
docker-compose -f docker-compose-pre.yml build

# 构建prod环境镜像
docker-compose -f docker-compose-prod.yml build
```

### 3.3 构建指定服务

可以只构建特定的服务：

```bash
# 只构建user-service
docker-compose -f docker-compose.build.yml build user-service

# 只构建post-service
docker-compose -f docker-compose.build.yml build post-service
```

### 3.4 并行构建加速

使用`--parallel`参数可以并行构建多个服务，加快构建速度：

```bash
docker-compose -f docker-compose.build.yml build --parallel
```

## 4. Dockerfile构建参数详解

每个服务的Dockerfile采用多阶段构建策略，以下是关键参数说明：

### 4.1 构建阶段参数

```dockerfile
# 构建阶段基础镜像
FROM golang:1.21-alpine AS builder

# 构建命令参数
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o user-service .
```

**参数说明：**
- `golang:1.21-alpine`: 使用Go 1.21版本的Alpine基础镜像，体积小且包含Go环境
- `AS builder`: 为构建阶段命名为builder，方便后续阶段引用
- `CGO_ENABLED=1`: 开启CGO支持，因为项目使用SQLite数据库
- `GOOS=linux`: 指定目标操作系统为Linux
- `-a`: 强制重新编译所有包，确保依赖正确
- `-installsuffix cgo`: 为CGO编译的二进制文件添加后缀，避免缓存问题
- `-o user-service`: 指定输出二进制文件名

### 4.2 运行阶段参数

```dockerfile
# 运行阶段基础镜像
FROM alpine:latest

# 安装运行时依赖
RUN apk --no-cache add ca-certificates sqlite-libs

# 暴露端口
EXPOSE 8081

# 启动命令
CMD ["./user-service"]
```

**参数说明：**
- `alpine:latest`: 使用最新版Alpine Linux作为运行环境，体积仅约5MB
- `--no-cache`: 不缓存安装包索引，减小镜像体积
- `ca-certificates`: 安装证书库，支持HTTPS请求
- `sqlite-libs`: 安装SQLite运行时库
- `EXPOSE 8081`: 声明容器监听的端口（仅用于文档和默认映射）
- `CMD ["./user-service"]`: 容器启动时执行的命令

## 5. docker-compose配置参数详解

docker-compose文件定义了服务的构建和运行配置：

```yaml
services:
  user-service:
    build:
      context: ./services/user-service
      dockerfile: Dockerfile
    container_name: user-service
    ports:
      - "8081:8081"
    environment:
      - ENV=staging
      - SERVER_PORT=8081
      - DB_DSN=dbs/staging/user_staging.db
      - POST_SERVICE_URL=http://post-service:8082
    volumes:
      - ./services/user-service/dbs:/app/dbs
    networks:
      - microservices-network
    restart: unless-stopped
```

**参数说明：**

### 5.1 构建配置
- `build.context`: 构建上下文路径，Docker将在此目录下查找文件
- `build.dockerfile`: Dockerfile文件名（如果不是默认的Dockerfile）

### 5.2 容器配置
- `container_name`: 容器名称，方便管理和引用
- `restart`: 重启策略，`unless-stopped`表示除非手动停止，否则总是重启

### 5.3 网络配置
- `ports`: 端口映射，格式为`宿主机端口:容器端口`
- `networks`: 容器加入的网络，实现服务间通信

### 5.4 数据持久化
- `volumes`: 数据卷挂载，格式为`宿主机路径:容器路径`
  - `./services/user-service/dbs:/app/dbs`: 将本地数据库目录挂载到容器中

### 5.5 环境变量
- `environment`: 设置容器内的环境变量
  - `ENV`: 运行环境（staging/pre/prod）
  - `SERVER_PORT`: 服务监听端口
  - `DB_DSN`: 数据库连接字符串
  - `POST_SERVICE_URL`/`USER_SERVICE_URL`: 服务间通信地址

### 5.6 服务依赖
- `depends_on`: 定义服务间依赖关系，确保启动顺序

## 6. 构建后运行服务

### 6.1 单服务运行

```bash
# 运行用户服务（staging环境）
docker run -d \
  --name user-service \
  -p 8081:8081 \
  -v $(pwd)/services/user-service/dbs:/app/dbs \
  -e ENV=staging \
  -e SERVER_PORT=8081 \
  -e DB_DSN=dbs/staging/user_staging.db \
  -e POST_SERVICE_URL=http://localhost:8082 \
  user-service:latest

# 运行动态服务（staging环境）
docker run -d \
  --name post-service \
  -p 8082:8082 \
  -v $(pwd)/services/post-service/dbs:/app/dbs \
  -e ENV=staging \
  -e SERVER_PORT=8082 \
  -e DB_DSN=dbs/staging/post_staging.db \
  -e USER_SERVICE_URL=http://localhost:8081 \
  post-service:latest
```

### 6.2 多服务运行（推荐）

使用docker-compose可以一键启动所有服务：

```bash
# 启动staging环境服务
docker-compose -f docker-compose-staging.yml up -d

# 启动pre环境服务
docker-compose -f docker-compose-pre.yml up -d

# 启动prod环境服务
docker-compose -f docker-compose-prod.yml up -d
```

**参数说明：**
- `-d`: 后台运行模式
- `up`: 启动服务

## 7. 镜像管理常用命令

### 7.1 查看镜像
```bash
# 查看所有镜像
docker images

# 查看特定镜像
docker images user-service
```

### 7.2 删除镜像
```bash
# 删除特定镜像
docker rmi user-service:latest

# 强制删除运行中容器的镜像
docker rmi -f user-service:latest
```

### 7.3 构建缓存清理
```bash
# 清理构建缓存
docker builder prune

# 清理所有未使用的镜像、容器和网络
docker system prune -a
```

## 8. 故障排除

### 8.1 构建失败排查

1. **依赖下载失败**
   - 检查网络连接
   - 考虑使用国内代理：`docker build --build-arg HTTP_PROXY=http://your-proxy:port -t user-service .`

2. **编译错误**
   - 检查Go版本兼容性（项目使用Go 1.21）
   - 确保CGO已启用（`CGO_ENABLED=1`）

3. **权限问题**
   - 确保Dockerfile中的命令有正确权限
   - 检查宿主机目录权限是否允许挂载

### 8.2 运行时问题

1. **端口冲突**
   - 修改端口映射：`-p 8083:8081`
   - 停止占用端口的其他服务

2. **服务通信失败**
   - 确保所有服务在同一网络中
   - 检查环境变量中的服务地址是否正确

3. **数据库问题**
   - 确保数据卷挂载正确
   - 检查数据库文件权限

## 9. 镜像优化建议

1. **使用多阶段构建**：如当前Dockerfile所示，减小最终镜像体积
2. **使用Alpine基础镜像**：体积小，安全性高
3. **合理使用缓存层**：将不常变化的指令（如下载依赖）放在前面
4. **清理构建缓存**：定期执行`docker builder prune`
5. **使用`.dockerignore`文件**：避免将不必要的文件复制到镜像中

## 10. 总结

本项目支持多种Docker镜像构建方式：

1. **单服务构建**：适用于开发和调试单个服务
2. **多服务构建**：使用docker-compose，推荐用于整体部署
3. **多环境支持**：通过不同的docker-compose文件支持staging、pre、prod环境

通过本指南提供的命令和参数说明，您可以灵活地构建和管理项目的Docker镜像，满足不同的开发和部署需求。