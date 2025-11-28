# 1_test_demo

这是一个简单的 Go 语言测试项目，演示了如何使用 Gin 和 GORM 实现用户注册、登录和信息查询功能。

## 功能特性

- **用户注册**: 支持用户名、手机号、密码注册。
- **用户登录**: 支持手机号、密码登录。
- **用户信息查询**:
    - 根据手机号查询单个用户信息。
    - 获取所有用户信息。
- **数据库**: 使用 SQLite 存储数据，无需额外安装数据库服务。

## 技术栈

- [Gin](https://github.com/gin-gonic/gin): HTTP Web 框架
- [GORM](https://gorm.io/): ORM 库
- [SQLite](https://www.sqlite.org/): 嵌入式数据库

## 快速开始

### 1. 安装依赖

```bash
go mod tidy
```

### 2. 运行项目

```bash
go run main.go
```

服务启动后将监听 `:8081` 端口。

## API 接口文档

### 1. 注册

- **URL**: `/register`
- **Method**: `POST`
- **Body**:
    ```json
    {
        "username": "testuser",
        "phone": "13800138000",
        "password": "password123"
    }
    ```

### 2. 登录

- **URL**: `/login`
- **Method**: `POST`
- **Body**:
    ```json
    {
        "phone": "13800138000",
        "password": "password123"
    }
    ```

### 3. 获取用户信息 (根据手机号)

- **URL**: `/users/:phone`
- **Method**: `GET`
- **Example**: `/users/13800138000`

### 4. 获取所有用户信息

- **URL**: `/users`
- **Method**: `GET`

go run main.go运行项目后，通过以下地址访问API文档：
http://localhost:8081/swagger/index.html