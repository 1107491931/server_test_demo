# Test Demo - 多环境配置说明

本项目已配置完成 **hosts + Nginx 反向代理** 多环境方案。

## 📁 项目结构

```
server_test_demo/
├── a-docs/                          # 文档目录
│   ├── Application.md               # 应用说明文档
│   ├── Docker_Build.md              # Docker构建文档
│   └── Multi_Environment_Setup.md   # 多环境配置文档（详细教程）⭐
├── config/                          # 环境配置目录
│   ├── .env.dev                     # Dev环境配置
│   ├── .env.staging                 # Staging环境配置
│   ├── .env.pre                     # Pre环境配置
│   └── .env.prod                    # Production环境配置
├── nginx/                           # Nginx配置目录
│   └── server_test_demo.conf        # Nginx反向代理配置
├── scripts/                         # 脚本目录
│   ├── setup-multi-env.sh           # 一键配置脚本 ⭐
│   └── verify-setup.sh              # 验证脚本
├── dbs/                             # 数据库目录
│   ├── dev/                         # Dev环境数据库
│   ├── staging/                     # Staging环境数据库
│   ├── pre/                         # Pre环境数据库
│   └── prod/                        # Production环境数据库
├── docker-compose.yml               # Docker Compose配置
├── Dockerfile                       # Docker镜像构建文件
└── main.go                          # 主程序
```

## 🚀 快速开始

### 方式1：使用一键配置脚本（推荐）

```bash
# 运行一键配置脚本
./scripts/setup-multi-env.sh
```

这个脚本会自动完成：
- ✅ 配置 hosts 文件
- ✅ 清除 DNS 缓存
- ✅ 配置 Nginx
- ✅ 启动 Docker 容器
- ✅ 验证所有配置

### 方式2：手动配置

详细步骤请参考：[a-docs/Multi_Environment_Setup.md](a-docs/Multi_Environment_Setup.md)

## 🌐 访问地址

| 环境 | 域名访问 | 端口访问 | Swagger 文档 |
|------|---------|---------|-------------|
| Dev | http://api.dev.myapp.com | http://127.0.0.1:8081 | http://api.dev.myapp.com/swagger/index.html |
| Staging | http://api.staging.myapp.com | http://127.0.0.1:8082 | http://api.staging.myapp.com/swagger/index.html |
| Pre | http://api.pre.myapp.com | http://127.0.0.1:8083 | http://api.pre.myapp.com/swagger/index.html |
| Prod | http://api.prod.myapp.com | http://127.0.0.1:8084 | http://api.prod.myapp.com/swagger/index.html |

## ✅ 验证配置

```bash
# 运行验证脚本
./scripts/verify-setup.sh
```

## 📋 常用命令

### Docker 容器管理

```bash
# 查看运行的容器
docker ps

# 启动所有环境
docker-compose up -d

# 启动单个环境
docker-compose up -d staging

# 停止所有环境
docker-compose down

# 查看日志
docker logs -f server_test_demo_staging

# 重启容器
docker restart server_test_demo_staging
```

### Nginx 管理

```bash
# 测试配置
sudo nginx -t

# 重新加载配置
sudo nginx -s reload

# 启动 Nginx
sudo nginx

# 停止 Nginx
sudo nginx -s stop

# 查看错误日志
tail -f /usr/local/var/log/nginx/server_test_demo_staging_error.log
```

### 测试接口

```bash
# 查询用户列表
curl http://api.staging.myapp.com/users

# 注册用户
curl -X POST http://api.staging.myapp.com/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "test",
    "phone": "13800138000",
    "password": "123456"
  }'

# 登录
curl -X POST http://api.staging.myapp.com/login \
  -H "Content-Type: application/json" \
  -d '{
    "phone": "13800138000",
    "password": "123456"
  }'
```

## 🔧 故障排查

### 域名无法访问

```bash
# 1. 检查 hosts 配置
cat /etc/hosts | grep "myapp.com"

# 2. 测试 DNS 解析
ping -c 3 api.staging.myapp.com

# 3. 清除 DNS 缓存
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder
```

### Nginx 502 错误

```bash
# 1. 检查容器是否运行
docker ps

# 2. 检查端口是否可访问
curl http://127.0.0.1:8082/users

# 3. 查看 Nginx 错误日志
tail -f /usr/local/var/log/nginx/server_test_demo_staging_error.log

# 4. 重启 Nginx
sudo nginx -s reload
```

### 容器无法启动

```bash
# 1. 查看容器状态
docker ps -a

# 2. 查看容器日志
docker logs server_test_demo_staging

# 3. 检查端口占用
lsof -i :8082

# 4. 重新启动容器
docker restart server_test_demo_staging
```

## 📚 文档

- [应用说明](a-docs/Application.md)
- [Docker构建指南](a-docs/Docker_Build.md)
- [多环境配置详细教程](a-docs/Multi_Environment_Setup.md) ⭐

## 🎯 下一步

1. **前端对接**：前端可以通过环境变量配置不同的 API 地址
2. **CI/CD 集成**：将部署流程自动化
3. **HTTPS 配置**：为本地环境配置 SSL 证书
4. **监控告警**：添加日志监控和告警机制

## 📝 注意事项

1. **数据隔离**：每个环境使用独立的数据库文件
2. **配置安全**：不要将 `.env.*` 文件提交到 Git（已在 .gitignore 中）
3. **端口冲突**：确保 8081-8084 端口未被占用
4. **Nginx 权限**：某些操作需要 sudo 权限

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
