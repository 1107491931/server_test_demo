# 第一阶段：构建
FROM golang:1.21-alpine AS builder

WORKDIR /build

# 复制 go.mod 和 go.sum
COPY go.mod go.sum ./

# 下载依赖
RUN go mod download

# 复制源代码
COPY . .

# 构建应用
RUN CGO_ENABLED=1 GOOS=linux go build -a -installsuffix cgo -o user-service .

# 第二阶段：运行
FROM alpine:latest

RUN apk --no-cache add ca-certificates sqlite-libs

WORKDIR /app

# 从构建阶段复制二进制文件
COPY --from=builder /build/user-service .

# 创建数据库目录
RUN mkdir -p /app/dbs/staging /app/dbs/pre /app/dbs/prod

# 暴露端口
EXPOSE 8081

# 运行应用
CMD ["./user-service"]
