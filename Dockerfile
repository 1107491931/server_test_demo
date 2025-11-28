# 第一阶段：构建阶段
# 使用官方 Go 镜像作为构建环境，别名 builder
FROM golang:alpine AS builder

# 安装构建依赖
# 由于 go-sqlite3 库需要 CGO 支持（因为它包含 C 代码），
# 所以我们需要安装 gcc 和 musl-dev 编译器
RUN apk add --no-cache gcc musl-dev

WORKDIR /app

# 复制依赖定义文件
# 先只复制 go.mod 和 go.sum，这样如果依赖没有变化，
# Docker 可以利用缓存跳过下载步骤，加速构建
COPY go.mod go.sum ./

# 下载依赖
# 根据 go.mod 下载项目所需的所有依赖包
RUN go mod download

# 复制源代码
# 将当前目录下的所有源代码复制到容器的工作目录中
COPY . .

# 编译应用程序
# - CGO_ENABLED=1: 开启 CGO，这是使用 go-sqlite3 所必需的
# - GOOS=linux: 指定目标操作系统为 Linux
# - -o main: 指定输出的二进制文件名为 main
RUN CGO_ENABLED=1 GOOS=linux go build -o main .

# 第二阶段：运行阶段
# 使用轻量级的 Alpine Linux 作为最终运行环境，可以显著减小镜像体积
FROM alpine:latest

WORKDIR /app

# 安装运行时依赖
# 安装 ca-certificates 证书库，防止应用发起 HTTPS 请求时报错
# --no-cache 表示不缓存索引，减小镜像体积
RUN apk add --no-cache ca-certificates

# 复制二进制文件
# 从 builder 阶段（第一阶段）复制编译好的二进制文件到当前镜像中
COPY --from=builder /app/main .

# 创建数据库目录
# 预先创建 dbs 目录，用于存放 SQLite 数据库文件
# 建议运行时挂载此目录以持久化数据
RUN mkdir -p dbs

# 暴露端口
# 声明容器运行时监听的端口，方便阅读和映射
EXPOSE 8081

# 启动命令
# 容器启动时默认执行的命令
CMD ["./main"]
