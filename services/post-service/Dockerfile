# 第一阶段：构建
FROM golang:1.21-alpine AS builder

# 安装构建依赖 (gcc 等)
RUN apk add --no-cache build-base

WORKDIR /build

# 复制 common 模块 (因为 go.mod 中使用了 replace common => ../common)
COPY common ./common

# 复制 post-service 的 go.mod 和 go.sum
COPY post-service/go.mod post-service/go.sum ./post-service/

# 切换到 post-service 目录
WORKDIR /build/post-service

# 下载依赖
RUN go mod download

# 复制 post-service 源代码
COPY post-service .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -tags musl -a -installsuffix cgo -o post-service .

# 第二阶段：运行
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/post-service/post-service .

# 创建数据库目录
RUN mkdir -p /app/dbs/staging /app/dbs/pre /app/dbs/prod

# 暴露端口
EXPOSE 8082

# 运行应用
CMD ["./post-service"]
