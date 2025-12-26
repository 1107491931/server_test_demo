# 第一阶段：构建
FROM golang:1.24-alpine AS builder

# 安装构建依赖 (gcc 等)
RUN apk add --no-cache build-base

# 配置 Go 模块代理（解决网络问题）
ENV GOPROXY=https://goproxy.cn,https://proxy.golang.org,direct
ENV GO111MODULE=on

WORKDIR /build

# 复制 common 模块
# 构建上下文是项目根目录，所以路径是 services/common
COPY services/common ./common

# 复制 post-service 的 go.mod 和 go.sum
COPY services/post-service/go.mod services/post-service/go.sum ./post-service/

# 切换到 post-service 目录
WORKDIR /build/post-service

# 自动更新 go.sum 以包含新添加的 MySQL 驱动
RUN go mod tidy

# 下载依赖（添加重试机制）
RUN go mod download || go mod download || go mod download

# 复制 post-service 源代码
COPY services/post-service .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -tags musl -a -installsuffix cgo -o post-service .

# 第二阶段：运行
FROM alpine:latest

# 更新包索引并安装依赖（添加重试机制）
RUN apk update && \
    apk add --no-cache ca-certificates sqlite-libs || \
    (apk update && apk add --no-cache ca-certificates sqlite-libs)

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/post-service/post-service .

# 创建数据库目录和密钥目录
RUN mkdir -p /app/dbs/staging /app/dbs/pre /app/dbs/prod /app/keys

# 暴露端口
EXPOSE 8082

# 运行应用
CMD ["./post-service"]
