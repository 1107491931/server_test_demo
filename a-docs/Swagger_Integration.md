# Swagger API 文档配置指南

本文档详细介绍了如何在本项目中配置和使用 Swagger 生成 API 文档。

## 1. 简介

本项目使用 [swag](https://github.com/swaggo/swag) 工具和 [gin-swagger](https://github.com/swaggo/gin-swagger) 中间件来自动生成和展示 API 文档。

## 2. 环境准备

在开始之前，请确保已安装 `swag` 命令行工具。

### 安装 swag CLI

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

验证安装：

```bash
swag -v
```

## 3. 项目配置

### 3.1 引入依赖

在 `go.mod` 中需要包含以下依赖（本项目已添加）：

```go
require (
    github.com/swaggo/files v1.0.1
    github.com/swaggo/gin-swagger v1.6.0
    github.com/swaggo/swag v1.16.2
)
```

### 3.2 Main 函数注解

在 `main.go` 的 `main` 函数上方，添加了全局 Swagger 注解：

```go
// @title           Test Demo API
// @version         2.0
// @description     This is a sample server for Test Demo.
// @host            localhost:8081
// @BasePath        /
func main() {
    // ...
}
```

### 3.3 路由配置

在 `main.go` 中注册了 Swagger 的路由：

```go
import (
    // ...
    swaggerFiles "github.com/swaggo/files"
    ginSwagger "github.com/swaggo/gin-swagger"
    _ "server_test_demo/docs" // ⚠️ 重要：必须导入生成的 docs 包
)

func main() {
    // ...
    r := gin.Default()

    // Swagger 文档路由
    r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
    // ...
}
```

## 4. 编写接口注解

在 Handler 层（如 `handler/user_handler.go`）的方法上添加注解。

### 示例：用户注册接口

```go
// @Summary      用户注册
// @Description  注册新用户
// @Tags         users
// @Accept       json
// @Produce      json
// @Param        request body RegisterRequest true "注册信息"
// @Success      200  {object}  map[string]interface{}
// @Router       /register [post]
func Register(c *gin.Context) {
    // ...
}
```

### 常用注解说明

- `@Summary`: 接口简短摘要
- `@Description`: 接口详细描述
- `@Tags`: 接口标签（用于分组）
- `@Accept`: 请求内容类型 (e.g., json)
- `@Produce`: 响应内容类型 (e.g., json)
- `@Param`: 请求参数说明
    - 格式: `参数名 参数类型 数据类型 是否必填 "描述"`
    - 示例: `request body RegisterRequest true "注册信息"`
- `@Success`: 成功响应说明
    - 格式: `状态码 {数据类型} 返回结构 "描述"`
- `@Router`: 路由路径和方法
    - 格式: `/path [method]`

## 5. 生成文档

每次修改了接口注解后，都需要重新生成文档。

在项目根目录下运行：

```bash
swag init
```

执行成功后，`docs` 目录下会生成/更新以下文件：
- `docs.go`
- `swagger.json`
- `swagger.yaml`

## 6. 访问文档

启动项目：

```bash
go run main.go
```

打开浏览器访问：

[http://localhost:8081/swagger/index.html](http://localhost:8081/swagger/index.html)


如果需要根据域名进行访问，比如`http://api.staging.myapp.com/swagger/index.html`, 则需要在Nginx中配置
- 参考项目根目录下`nginx/testdemo.conf`
- 将文件内容copy到系统目录下`/opt/homebrew/etc/nginx/servers/testdemo.conf`

## 7. 常见问题

### Q: 访问文档页面显示 404？
A: 检查 `main.go` 中是否正确导入了 `_ "server_test_demo/docs"` 包。

### Q: 修改了注解但文档没更新？
A: 必须重新运行 `swag init` 命令，并重启 Go 服务。

### Q: 无法解析依赖包中的类型？
A: 如果使用了外部依赖的结构体作为响应对象，可能需要使用 `--parseDependency` 参数：
```bash
swag init --parseDependency --parseInternal
```
