# Server Test Demo - 微服务架构项目

## 🎉 项目重构完成！

本项目已成功从单体应用重构为**微服务架构**，现在包含两个独立的微服务：

### 📦 微服务列表

1. **用户服务 (User Service)** 
   - 端口: `8081`
   - 功能: 用户注册、登录、用户信息管理
   
2. **动态服务 (Post Service)**
   - 端口: `8082`
   - 功能: 动态发布、查询、点赞、转发、收藏

---

## 📚 重要文档（请按顺序阅读）

### 1️⃣ [重构完成总结.md](./重构完成总结.md) ⭐ **从这里开始**
- 📋 所有创建的文件清单
- 🏗️ 完整的项目结构
- 🎯 核心功能说明
- 🚀 快速启动方法
- 🧪 快速测试示例

### 2️⃣ [微服务架构设计文档.md](./微服务架构设计文档.md)
- 🏛️ 整体架构设计
- 📊 数据库详细设计
- 📡 完整的API接口文档
- 🔗 服务间通信设计
- 💡 优化建议

### 3️⃣ [微服务快速使用指南.md](./微服务快速使用指南.md)
- 🚀 三种启动方式
- 🧪 API测试示例
- ✅ 服务间通信验证
- 📝 完整测试流程
- ❓ 常见问题解答

### 4️⃣ [迁移指南.md](./迁移指南.md)
- 🔄 新旧项目对比
- 📦 数据迁移方案
- 🔌 API接口变化
- 💻 前端代码迁移
- 📋 详细迁移步骤

---

## 🚀 快速开始（3步启动）

### 步骤1: 赋予执行权限
```bash
chmod +x scripts/start_microservices.sh
```

### 步骤2: 启动所有服务
```bash
./scripts/start_microservices.sh
```

### 步骤3: 验证服务
```bash
# 检查用户服务
curl http://localhost:8081/health

# 检查动态服务
curl http://localhost:8082/health
```

**就这么简单！** 🎉

---

## 🧪 快速测试（复制粘贴即可）

```bash
# 1. 注册用户
curl -X POST http://localhost:8081/api/v1/users/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "张三",
    "email": "zhangsan@example.com",
    "phone": "13800138000",
    "password": "password123"
  }'

# 2. 发布动态
curl -X POST http://localhost:8082/api/v1/posts \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": 1,
    "content": "这是我的第一条动态！",
    "images": ["https://example.com/image1.jpg"]
  }'

# 3. 获取动态详情（会自动获取用户信息）
curl http://localhost:8082/api/v1/posts/1
```

---

## 📁 项目结构一览

```
server_test_demo/
├── 📄 重构完成总结.md           ⭐ 从这里开始
├── 📄 微服务架构设计文档.md      📖 详细设计
├── 📄 微服务快速使用指南.md      🚀 快速上手
├── 📄 迁移指南.md               🔄 迁移参考
├── 📄 README_MICROSERVICES.md   📋 项目README
│
├── 📁 services/
│   ├── 📁 user-service/         👤 用户服务 (8081)
│   └── 📁 post-service/         📝 动态服务 (8082)
│
├── 📁 scripts/
│   └── start_microservices.sh   🚀 启动脚本
│
└── 📄 docker-compose-microservices.yml  🐳 Docker部署
```

---

## ✨ 核心特性

✅ **完全解耦** - 两个服务完全独立，可独立开发和部署  
✅ **HTTP通信** - 使用标准的HTTP + JSON进行服务间通信  
✅ **数据隔离** - 每个服务有独立的数据库  
✅ **易于扩展** - 可轻松添加新的微服务（如评论服务、通知服务）  
✅ **容器化部署** - 支持Docker和Docker Compose  
✅ **完整文档** - 4份详细文档，覆盖设计、使用、迁移  

---

## 🔗 服务地址

| 服务 | 地址 | 健康检查 |
|-----|------|---------|
| 用户服务 | http://localhost:8081 | http://localhost:8081/health |
| 动态服务 | http://localhost:8082 | http://localhost:8082/health |

---

## 📖 API 快速参考

### 用户服务 API (8081)

```
POST /api/v1/users/register      # 用户注册
POST /api/v1/users/login         # 用户登录
GET  /api/v1/users/:user_id      # 获取用户信息
GET  /api/v1/users               # 获取所有用户
```

### 动态服务 API (8082)

```
POST /api/v1/posts                    # 发布动态
GET  /api/v1/posts/:post_id           # 获取动态详情
GET  /api/v1/posts/user/:user_id      # 获取用户的所有动态
POST /api/v1/posts/:post_id/like      # 点赞动态
POST /api/v1/posts/:post_id/forward   # 转发动态
POST /api/v1/posts/:post_id/favorite  # 收藏动态
```

---

## 🎯 下一步

1. ✅ **阅读文档** - 从 `重构完成总结.md` 开始
2. ✅ **启动服务** - 使用启动脚本快速启动
3. ✅ **测试API** - 使用提供的测试用例
4. ✅ **查看代码** - 了解微服务实现细节
5. ✅ **扩展功能** - 根据需求添加新功能

---

## 💡 提示

- 📖 **新手**: 先阅读 `重构完成总结.md` 和 `微服务快速使用指南.md`
- 🏗️ **架构师**: 重点查看 `微服务架构设计文档.md`
- 🔄 **迁移团队**: 参考 `迁移指南.md`
- 🐛 **遇到问题**: 查看各文档中的"常见问题"部分

---

## 🎓 技术栈

- **语言**: Go 1.21
- **Web框架**: Gin
- **ORM**: GORM
- **数据库**: SQLite
- **容器化**: Docker, Docker Compose
- **通信**: HTTP + JSON

---

## 📞 获取帮助

遇到问题？查看相应文档：

- 🏗️ 架构问题 → `微服务架构设计文档.md`
- 🚀 使用问题 → `微服务快速使用指南.md`
- 🔄 迁移问题 → `迁移指南.md`
- 📋 项目概览 → `重构完成总结.md`

---

## 🎉 开始使用

**准备好了吗？让我们开始吧！**

```bash
# 一键启动
./scripts/start_microservices.sh
```

**祝你使用愉快！** 🚀