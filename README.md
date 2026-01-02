这是一个简单的 Go 语言测试项目，用于演示多环境部署的流程。
本项目需要通过brew安装的插件：
```
brew install go
brew install redis
brew install nginx
brew install flyctl // 云部署
```

>首次运行项目，可以通过openssl生成私钥和公钥, 方便通过环境变量运行服务， 可参考`a-docs/JWT认证系统集成指南.md`

# v2.0.3 
- 增加了zeabur.yaml配置文件， 用于zeabur部署

# v2.0.2
- 所有接口改成POST请求
- common模块定义所以依赖库版本，各个服务使用common中定义的版本，避免版本不一致
- 添加了请求/响应拦截器， 用于记录请求/响应信息，参考文档 `a-docs/请求响应拦截器中间件.md`
- 增加token相关内容，参考文档 `a-docs/JWT认证系统集成指南.md`

# V2.0.1
在V2.0.1基础上补充的功能
- 速率限制， 防止暴力请求数据， 参考文档 `a-docs/请求速率限制.md`
- common模块中增加rete_limiter.go
- user-service、post-service中的main.go中增加引用，设置API的请求速率


# V2.0.0
项目采用了微服务架构，有两个微服务：登录注册模块、发布动态模块。
示例中的`http://localhost:8082`都可以替代为 `http://we-circle-staging.duckdns.org`
项目中使用到的环境变量：
- `ENV`: 环境变量, 目前仅用于在main.go中打印
- `SERVER_PORT`: 服务端口， 用于指定服务运行的端口, 如用户模块运行在8081端口, 动态模块运行在8082端口
- `DB_DSN`: 数据库连接字符串， 用于指定数据库连接信息, 如 `dbs/staging/user_staging.db`
- `POST_SERVICE_URL`: 动态服务URL，运行用户模块时，用于对动态模块发起请求, 如 `http://localhost:8082`
- `USER_SERVICE_URL`: 用户服务URL，运行动态模块时，用于对用户模块发起请求, 如 `http://localhost:8081`
- `JWT_PRIVATE_KEY`: JWT认证系统的私钥， 用于生成token，通过命令openssl生成， 如private.pem
- `JWT_PUBLIC_KEY`: JWT认证系统的公钥， 用于验证token， 通过命令openssl将私钥导出公钥, 如public.pem
- `JWT_ISSUER`: 用于指定JWT的发行者, 如 `we-circle-staging`, 默认`we-circle-prod`, 在生成token以及验证token有效性时会使用这个值
- `REDIS_HOST`: 用于指定Redis数据库的主机地址, 可以是本地地址如 `localhost`, 可以是远程地址如 `we-circle-staging.duckdns.org`, 默认值`localhost`
- `REDIS_PORT`: 用于指定Redis数据库的端口号, 如 `6379`, 默认值`6379`
- `REDIS_PASSWORD`: 用于指定Redis数据库的密码, 如 `123456`, 开发阶段可是设置密码为空， prod环境必须设置密码，否则大家都可以访问Redis数据库, 默认值为空
- `REDIS_DB`: 用于指定Redis数据库的数据库分层， 如果不设置，则所有缓存数据都在一起，设置不同的值则是区分不同的场景缓存, 如 `0` 表示默认数据库， `1` 表示第二个数据库, 以此类推, 默认值`0`
- `LOKI_URL`: 用于指定Grafana Loki日志系统的URL, 如 `https://logs-prod-021.grafana.net/loki/api/v1/push`
- `LOKI_USER_ID`: 用于指定Grafana Loki日志系统的用户ID, 如 `1441206`
- `LOKI_TOKEN`: 用于指定Grafana Loki日志系统的token, 如 `sa-1-go-app-logger-0630bb76-d8ac-4fc9-8ef5-fa9c3acfc962`
- `LOG_LEVEL="development"` 用于设置日志级别，development为开发环境, 传入其它值为生产环境. 环境变量代码在`common/config/config.go`中

redis几个参数解释可以见：https://ai.feishu.cn/docx/WKkkd6nqToAjz4xrEzScKlrjnxb

相关功能介绍：
## 用户模块
### 运行项目
提前Nginx需要运行起来： `sudo nginx`, 方便使用本地域名
```
ENV=staging \
SERVER_PORT=8081 \
DB_DSN=dbs/staging/user_staging.db \
POST_SERVICE_URL=http://localhost:8082 \
JWT_PRIVATE_KEY="$(cat ../../private.pem)" \
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
GRAFANA_TOKEN="<YOUR_GRAFANA_TOKEN>" \
LOKI_URL="https://logs-prod-021.grafana.net/loki/api/v1/push" \
LOKI_USER_ID="<YOUR_LOKI_USER_ID>" \
PROM_REMOTE_URL="https://prometheus-prod-36-prod-us-west-0.grafana.net/api/prom/push" \
PROM_USER_ID="<YOUR_PROM_USER_ID>" \
go run main.go
```

Grafana token获取地址：https://chomayvip.grafana.net/a/grafana-auth-app， token获取流程：
- 左边栏 -> 管理 -> 用户和访问权限 -> Cloud access policies -> Create Access Policy
- 输入名称、realms选择chomayvip、scopes中logs与metrics勾选read与write
- 点击crate后，在Access Policies列表中会看到新增的项
- 点击Add Token，会生成token，复制token

Grafana中URL与userId获取地址：https://chomayvip.grafana.net/connections/datasources/edit/grafanacloud-logs
- 左侧栏 -> Connections -> Data Sources -> 搜索logs并点击grafanacloud-chomayvip-logs -> 可以看到URL、User


### 接口测试
```
// 1. GET 请求（获取所有用户）
curl http://we-circle-staging.duckdns.org/api/v1/users

// 2. GET 请求（获取指定用户）
curl http://we-circle-staging.duckdns.org/api/v1/users/1

// 3. POST 请求（注册用户）
curl -X POST http://we-circle-staging.duckdns.org/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{"username": "李四", "email": "lisi@example.com", "phone": "13800138001", "password": "123456"}'

// 4. POST 请求（登录）
curl -X POST http://we-circle-staging.duckdns.org/api/v1/users/login \
  -H "Content-Type: application/json" \
  -d '{"phone": "13800138000", "password": "123456"}'
```

## 动态模块
### 运行项目
```
// 环境变量目前仅设置部分参数，其它参数使用了默认值
ENV=staging \
SERVER_PORT=8082 \
DB_DSN=dbs/staging/post_staging.db \
USER_SERVICE_URL=http://localhost:8081 \
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
GRAFANA_TOKEN="<YOUR_GRAFANA_TOKEN>" \
LOKI_URL="https://logs-prod-021.grafana.net/loki/api/v1/push" \
LOKI_USER_ID="<YOUR_LOKI_USER_ID>" \
PROM_REMOTE_URL="https://prometheus-prod-36-prod-us-west-0.grafana.net/api/prom/push" \
PROM_USER_ID="<YOUR_PROM_USER_ID>" \
go run main.go
```
### 接口测试
```
# GET 请求（获取所有动态）
curl http://we-circle-staging.duckdns.org/api/v1/posts

# GET 请求（获取指定动态）
curl http://we-circle-staging.duckdns.org/api/v1/posts/1

# POST 请求（创建动态）
curl -X POST http://we-circle-staging.duckdns.org/api/v1/posts/create \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -d '{"user_id": 1, "content": "测试动态", "images": []}'

# POST 请求（点赞）
curl -X POST http://we-circle-staging.duckdns.org/api/v1/posts/1/like

# POST 请求（转发）
curl -X POST http://we-circle-staging.duckdns.org/api/v1/posts/forward \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -d '{"user_id": 1, "post_id": 1, "content": "转发内容"}'

# POST 请求（收藏）
curl -X POST http://we-circle-staging.duckdns.org/api/v1/posts/favorite \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -d '{"user_id": 1, "post_id": 1}'
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
JWT_PRIVATE_KEY="$(cat ../../private.pem)" \
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
go run main.go
```

- 运行动态服务
```
ENV=staging \
SERVER_PORT=8082 \
DB_DSN=dbs/staging/post_staging.db \
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
go run main.go
```
- 接口测试
```
// jq '.'作用是将json数据格式化，否则全部显示在一行
curl -s -X POST "http://localhost:8081/api/v1/users/get_with_posts" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
  -d '{"user_id": 1, "page": 1, "page_size": 5}' | jq '.'
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
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
go run main.go
```
- 运行用户模块
```
ENV=staging \
SERVER_PORT=8081 \
DB_DSN=dbs/staging/user_staging.db \
JWT_PRIVATE_KEY="$(cat ../../private.pem)" \
JWT_PUBLIC_KEY="$(cat ../../public.pem)" \
go run main.go
```
- 接口测试
-s参数是隐藏请求进度信息和错误信息
```
curl -s -X POST http://localhost:8082/api/v1/posts/get_user_by_post_id \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
    -d '{"post_id": 1}' | jq '.'
或者
curl -s -X POST http://we-circle-staging.duckdns.org/api/v1/posts/get_user_by_post_id \
    -H "Content-Type: application/json" \
    -H "Authorization: Bearer <YOUR_ACCESS_TOKEN>" \
    -d '{"post_id": 1}' | jq '.'
```
## Dock镜像构建
参考文档`a-docs/Docker构建多服务镜像.md`
- 包含手动构建、docker-compose构建
- 支持设置版本号
- 介绍了单个镜像启动、多个镜像同时启动

## Swggger
参考`a-docs/Swagger集成.md`
- 是单个服务的配置，每个服务单独配置，会产生各个服务的Swagger文档
- 执行命令：`swag init --parseDependency --parseInternal` 构建Swagger，因为服务依赖了外部服务代码common
- 用户服务或镜像启动后，就可以访问Swagger页面了：
  - `http://localhost:8081/swagger/index.html`
  - 

## 问题
### POST_SERVICE_URL为何不能使用环境配置中的值
比如运行项目时，参数使用的`POST_SERVICE_URL=http://localhost:8081`, 而环境配置中使用的`USER_SERVICE_URL=http://user-service:8081`.

原因：user-service 是 Docker 容器网络中的服务名，本地运行无法解析。服务运行在 Docker 容器网络中，Docker 内置 DNS 可以解析容器名称（如 post-service），容器之间可以通过服务名互相访问。

### 宿主端口、容器端口的理解
流程：
  - 用户请求 127.0.0.1:8082 → 到达服务器（宿主机）的 8082 端口
  - Docker 检测到宿主机 8082 有请求
  - Docker 根据映射规则 "8082:8082" 转发请求
  - 请求到达容器内的 8082 端口
  - 容器内应用（监听 8082）接收并处理请求

docker-compose.yml中ports: "8082:8082"的含义：
- 第一个8082：代表宿主机端口(服务器端口)，用户通过接口请求127.0.0.1:8082请求到了服务器
- 第二个8082：代表容器端口号， 宿主机收到请求后，经过docker转发，将请求转发到容器内8082端口号的服务上。
- main.go中监听的端口号，其实就是容器内部服务的端口号， 也用于在不同微服务之间进行通信使用

服务器运行服务时，两个服务的端口不能相同，否则在运行第二个服务时会报错。