这是一个简单的 Go 语言测试项目，用于演示多环境部署的流程。
# V2.0.0
项目采用了微服务架构，有两个微服务：登录注册模块、发布动态模块。
示例中的`http://localhost:8082`都可以替代为 `http://api.staging.myapp.com`

相关功能介绍：
## 用户模块
### 运行项目
提前Nginx需要运行起来： `sudo nginx`, 方便使用本地域名
```
ENV=staging \
SERVER_PORT=8081 \
DB_DSN=dbs/staging/user_staging.db \
go run main.go
```
### 接口测试
```
// 1. GET 请求（获取所有用户）
curl http://api.staging.myapp.com/api/v1/users

// 2. GET 请求（获取指定用户）
curl http://api.staging.myapp.com/api/v1/users/1

// 3. POST 请求（注册用户）
curl -X POST http://api.staging.myapp.com/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username": "李四", "email": "lisi@example.com", "phone": "13800138001", "password": "123456"}'

// 4. POST 请求（登录）
curl -X POST http://api.staging.myapp.com/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "password": "123456"}'
```

## 动态模块
### 运行项目
```
ENV=staging \
SERVER_PORT=8082 \
DB_DSN=dbs/staging/post_staging.db \
go run main.go
```
### 接口测试
```
# GET 请求（获取所有动态）
curl http://api.staging.myapp.com/api/v1/posts

# GET 请求（获取指定动态）
curl http://api.staging.myapp.com/api/v1/posts/1

# POST 请求（创建动态）
curl -X POST http://api.staging.myapp.com/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{"user_id": 1, "content": "测试动态", "images": []}'

# POST 请求（点赞）
curl -X POST http://api.staging.myapp.com/api/v1/posts/1/like

# POST 请求（转发）
curl -X POST http://api.staging.myapp.com/api/v1/posts/1/forward

# POST 请求（收藏）
curl -X POST http://api.staging.myapp.com/api/v1/posts/1/favorite
```

## 服务间通信
### 根据user_id获取所有动态列表
项目内运行的话，需要设置`POST_SERVICE_URL=http://localhost:8082`,这个url参数用于在用户模块对动态模块发起请求，传入user_id， 动态模块查询到列表数据返回。
- 运行用户服务：
```
ENV=staging \
SERVER_PORT=8081 \
DB_DSN=dbs/staging/user_staging.db \
POST_SERVICE_URL=http://localhost:8082 \
go run main.go
```

- 运行动态服务
```
ENV=staging \
SERVER_PORT=8082 \
DB_DSN=dbs/staging/post_staging.db \
go run main.go
```
- 接口测试
```
// jq '.'作用是将json数据格式化，否则全部显示在一行
curl -s "http://localhost:8081/api/v1/users/1/posts?page=1&page_size=5" | jq '.'
```
输出示例：
```
{
  "code": 200,
  "message": "success",
  "data": {
    "user_id": 1,
    "username": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "created_at": "2025-11-30 00:36:54",
    "posts": [
      {
        "post_id": 6,
        "user_id": 1,
        "content": "测试动态",
        "images": [
          "https://example.com/test.jpg"
        ],
        "like_count": 0,
        "forward_count": 0,
        "favorite_count": 0,
        "created_at": "2025-11-30 11:24:48"
      },
      {
        "post_id": 5,
        "user_id": 1,
        "content": "测试动态",
        "images": [
          "https://example.com/test.jpg"
        ],
        "like_count": 0,
        "forward_count": 0,
        "favorite_count": 0,
        "created_at": "2025-11-30 11:24:47"
      },
      {
        "post_id": 4,
        "user_id": 1,
        "content": "测试动态",
        "images": [
          "https://example.com/test.jpg"
        ],
        "like_count": 0,
        "forward_count": 0,
        "favorite_count": 0,
        "created_at": "2025-11-30 11:24:46"
      },
      {
        "post_id": 3,
        "user_id": 1,
        "content": "测试动态",
        "images": [
          "https://example.com/test.jpg"
        ],
        "like_count": 0,
        "forward_count": 0,
        "favorite_count": 0,
        "created_at": "2025-11-30 11:24:45"
      },
      {
        "post_id": 2,
        "user_id": 1,
        "content": "测试动态",
        "images": [
          "https://example.com/test.jpg"
        ],
        "like_count": 0,
        "forward_count": 0,
        "favorite_count": 0,
        "created_at": "2025-11-30 11:24:44"
      }
    ],
    "total": 6
  }
}
```
### 根据动态id获取用户信息
其中设置`POST_SERVICE_URL=http://localhost:8081`是为了请求用户服务，相当于是请求用户服务的BaseUrl

- 运行动态模块
```
ENV=staging \
SERVER_PORT=8082 \
DB_DSN=dbs/staging/post_staging.db \
POST_SERVICE_URL=http://localhost:8081 \
go run main.go
```
- 运行用户模块
```
ENV=staging \
SERVER_PORT=8081 \
DB_DSN=dbs/staging/user_staging.db \
go run main.go
```
- 接口测试
-s参数是隐藏请求进度信息和错误信息
```
curl -s http://localhost:8082/api/v1/posts/1/user | jq '.'
或者
curl -s http://api.staging.myapp.com/api/v1/posts/1/user | jq '.'
```
## Dock镜像构建
参考文档`a-docs/Docker_Build_More_Service.md`
- 包含手动构建、docker-compose构建
- 支持设置版本号
- 介绍了单个镜像启动、多个镜像同时启动

## Swggger
参考`a-docs/Swagger_Integration.md`
- 是单个服务的配置，每个服务单独配置，会产生各个服务的Swagger文档
- 执行命令：`swag init --parseDependency --parseInternal` 构建Swagger，因为服务依赖了外部服务代码common
- 用户服务或镜像启动后，就可以访问Swagger页面了：
  - `http://localhost:8081/swagger/index.html`
  - 

## 问题
##### POST_SERVICE_URL为何不能使用环境配置中的值
比如运行项目时，参数使用的`POST_SERVICE_URL=http://localhost:8081`, 而环境配置中使用的`USER_SERVICE_URL=http://user-service:8081`.

原因：user-service 是 Docker 容器网络中的服务名，本地运行无法解析。服务运行在 Docker 容器网络中，Docker 内置 DNS 可以解析容器名称（如 post-service），容器之间可以通过服务名互相访问。

TODO： 
- 代码理解
- Docker构建镜像、运行镜像
- Swagger
- 相册截图， 即宿主机端口、容器端口
- 飞书copy出来的md内容部分
- 某个服务回滚等