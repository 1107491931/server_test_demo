# 第一阶段：构建
FROM golang:1.21-alpine AS builder

# 安装构建依赖 (gcc 等)
RUN apk add --no-cache build-base

WORKDIR /build

# 复制 common 模块 (因为 go.mod 中使用了 replace common => ../common)
# - 当你运行 docker build ... services/user-service 时，Docker 的构建上下文（Context）被限制在 services/user-service 目录下，
# - 因此 Docker 无法访问上级目录中的 common 模块，导致 go mod download 失败。
COPY common ./common

# 复制 user-service 的 go.mod 和 go.sum
COPY user-service/go.mod user-service/go.sum ./user-service/

# 切换到 user-service 目录
WORKDIR /build/user-service

# 下载依赖
RUN go mod download

# 复制 user-service 源代码
COPY user-service .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux CGO_CFLAGS="-D_LARGEFILE64_SOURCE" go build -tags musl -a -installsuffix cgo -o user-service .

# 第二阶段：运行
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/user-service/user-service .

# 创建数据库目录
RUN mkdir -p /app/dbs/staging /app/dbs/pre /app/dbs/prod

# 暴露端口
EXPOSE 8081

# 运行应用
CMD ["./user-service"]
