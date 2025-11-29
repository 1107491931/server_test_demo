# 用户模块与发布动态模块通信流程

本文档介绍用户服务（user-service）和动态服务（post-service）之间的服务间通信流程。

## 目录

1. [概述](#概述)
2. [功能说明](#功能说明)
3. [通信流程](#通信流程)
4. [API接口](#api接口)
5. [示例请求](#示例请求)

---

## 概述

本项目采用微服务架构，用户服务和动态服务相互独立，通过HTTP API进行服务间通信：

- **用户服务（user-service）**：负责用户信息管理
- **动态服务（post-service）**：负责动态内容管理
- **共享基础库（common）**：提供HTTP客户端和统一响应工具

---

## 功能说明

### 1. 用户模块调用动态模块

**功能**：根据用户ID获取该用户发布的所有动态

**场景**：查看某个用户的个人主页，需要展示该用户的所有动态

**实现**：
- 用户服务通过 `PostClient` 调用动态服务的API
- 接口路径：`GET /api/v1/users/{user_id}/posts`

### 2. 动态模块调用用户模块

**功能**：根据动态中的用户ID获取完整的用户信息

**场景**：
- 查看动态详情时，需要显示发布者的用户信息
- 查看动态列表时，需要显示每个动态的发布者信息

**实现**：
- 动态服务通过 `UserClient` 调用用户服务的API
- 在以下接口中实现：
  - `GET /api/v1/posts/{post_id}` - 获取动态详情（包含用户信息）
  - `GET /api/v1/posts/user/{user_id}` - 获取用户的所有动态（包含用户信息）
  - `GET /api/v1/posts` - 获取所有动态（批量获取用户信息）

---

## 通信流程

### 流程1：用户模块获取用户动态列表

```
客户端请求
    ↓
用户服务 (user-service)
    ↓
1. 验证用户ID是否存在
    ↓
2. 调用 PostClient.GetUserPosts(userID)
    ↓
3. HTTP请求 → 动态服务
    GET /api/v1/posts/user/{user_id}?page=1&page_size=10
    ↓
动态服务 (post-service)
    ↓
4. 查询数据库获取动态列表
    ↓
5. 返回动态列表数据
    ↓
用户服务
    ↓
6. 组装响应（用户信息 + 动态列表）
    ↓
返回给客户端
```

### 流程2：动态模块获取用户信息

#### 2.1 获取单个动态详情（包含用户信息）

```
客户端请求
    ↓
动态服务 (post-service)
    ↓
1. 根据动态ID查询动态信息
    ↓
2. 调用 UserClient.GetUserInfo(userID)
    ↓
3. HTTP请求 → 用户服务
    GET /api/v1/users/{user_id}
    ↓
用户服务 (user-service)
    ↓
4. 查询数据库获取用户信息
    ↓
5. 返回用户信息
    ↓
动态服务
    ↓
6. 组装响应（动态信息 + 用户信息）
    ↓
返回给客户端
```

#### 2.2 获取动态列表（批量获取用户信息）

```
客户端请求
    ↓
动态服务 (post-service)
    ↓
1. 查询数据库获取动态列表
    ↓
2. 提取所有唯一的用户ID
    ↓
3. 调用 UserClient.BatchGetUserInfo(userIDs)
    ↓
4. HTTP请求 → 用户服务
    POST /api/v1/users/batch
    Body: {"user_ids": [1, 2, 3]}
    ↓
用户服务 (user-service)
    ↓
5. 批量查询数据库获取用户信息
    ↓
6. 返回用户信息Map
    ↓
动态服务
    ↓
7. 将用户信息关联到每个动态
    ↓
返回给客户端
```

---

## API接口

### 用户服务接口

#### 1. 获取用户信息及其动态列表

**接口**：`GET /api/v1/users/{user_id}/posts`

**请求参数**：
- `user_id` (path): 用户ID
- `page` (query, 可选): 页码，默认1
- `page_size` (query, 可选): 每页数量，默认10

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "created_at": "2024-01-01 10:00:00",
    "posts": [
      {
        "post_id": 1,
        "user_id": 1,
        "content": "这是我的第一条动态",
        "images": [],
        "like_count": 10,
        "forward_count": 5,
        "favorite_count": 3,
        "created_at": "2024-01-02 12:00:00"
      }
    ],
    "total": 1
  }
}
```

### 动态服务接口

#### 1. 获取动态详情（包含用户信息）

**接口**：`GET /api/v1/posts/{post_id}`

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "post_id": 1,
    "user_id": 1,
    "user_info": {
      "user_id": 1,
      "username": "张三",
      "email": "zhangsan@example.com",
      "phone": "13800138000",
      "created_at": "2024-01-01 10:00:00"
    },
    "content": "这是我的第一条动态",
    "images": [],
    "like_count": 10,
    "forward_count": 5,
    "favorite_count": 3,
    "created_at": "2024-01-02 12:00:00"
  }
}
```

#### 2. 获取用户的所有动态（包含用户信息）

**接口**：`GET /api/v1/posts/user/{user_id}`

**请求参数**：
- `user_id` (path): 用户ID
- `page` (query, 可选): 页码，默认1
- `page_size` (query, 可选): 每页数量，默认10

**响应示例**：
```json
{
  "code": 200,
  "message": "success",
  "data": {
    "total": 1,
    "page": 1,
    "page_size": 10,
    "posts": [
      {
        "post_id": 1,
        "user_id": 1,
        "user_info": {
          "user_id": 1,
          "username": "张三",
          "email": "zhangsan@example.com",
          "phone": "13800138000",
          "created_at": "2024-01-01 10:00:00"
        },
        "content": "这是我的第一条动态",
        "images": [],
        "like_count": 10,
        "forward_count": 5,
        "favorite_count": 3,
        "created_at": "2024-01-02 12:00:00"
      }
    ]
  }
}
```

---

## 示例请求

### 示例1：获取用户及其动态列表

```bash
# 请求
curl -X GET "http://localhost:8081/api/v1/users/1/posts?page=1&page_size=10"

# 响应
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "created_at": "2024-01-01 10:00:00",
    "posts": [...],
    "total": 5
  }
}
```

### 示例2：获取动态详情（包含用户信息）

```bash
# 请求
curl -X GET "http://localhost:8082/api/v1/posts/1"

# 响应
{
  "code": 200,
  "message": "success",
  "data": {
    "post_id": 1,
    "user_id": 1,
    "user_info": {
      "user_id": 1,
      "username": "张三",
      "email": "zhangsan@example.com",
      "phone": "13800138000",
      "created_at": "2024-01-01 10:00:00"
    },
    "content": "这是我的第一条动态",
    ...
  }
}
```

---

## 技术实现

### 服务间通信

两个服务通过共享基础库 `common` 中的 `HTTPClient` 进行通信：

- **用户服务**：使用 `PostClient` 调用动态服务
- **动态服务**：使用 `UserClient` 调用用户服务

### 环境变量配置

服务间通信的URL通过环境变量配置：

- `POST_SERVICE_URL`：动态服务地址（默认：http://localhost:8082）
- `USER_SERVICE_URL`：用户服务地址（默认：http://localhost:8081）

### 错误处理

- 如果服务间调用失败，会记录错误日志
- 部分接口在服务调用失败时仍会返回主要数据（如动态列表），但不包含关联信息（如用户信息）

---

## 总结

1. **用户服务**可以通过 `PostClient` 调用动态服务，获取用户发布的所有动态
2. **动态服务**可以通过 `UserClient` 调用用户服务，获取动态发布者的完整用户信息
3. 两个服务相互独立，通过HTTP API进行通信，实现了服务解耦
4. 使用共享基础库统一管理HTTP客户端和响应格式，提高了代码复用性

