# 🚀 JD任务平台 - Go语言版本

## ✨ 特性

- ✅ **完整的 Swagger API 文档** - 类似 FastAPI 的在线文档
- ✅ **自动连接现有 MySQL 数据库** - 无需修改数据库结构
- ✅ **JWT 认证** - 安全的用户认证
- ✅ **RESTful API** - 标准的 API 设计
- ✅ **高性能** - Go 语言原生并发支持
- ✅ **易于部署** - 单个可执行文件

## 📋 快速开始

### 1. 安装依赖

```bash
# 安装 Go 依赖
go mod tidy

# 安装 Swag 工具（用于生成 Swagger 文档）
go install github.com/swaggo/swag/cmd/swag@latest
```

### 2. 生成 Swagger 文档

```bash
# 生成 API 文档
swag init

# 这会在项目根目录生成 docs 文件夹
```

### 3. 运行项目

```bash
# 方式1: 直接运行
go run main.go

# 方式2: 编译后运行
go build -o jd-task-platform.exe
.\jd-task-platform.exe
```

### 4. 访问 API 文档

启动成功后，访问以下地址：

- **Swagger UI**: http://localhost:5001/swagger/index.html
- **API 文档**: http://localhost:5001/docs/index.html
- **根路径**: http://localhost:5001/

## 📚 API 文档预览

Swagger 文档包含所有 API 接口，并可以直接在线测试：

- **认证模块**
  - POST /api/auth/register - 用户注册
  - POST /api/auth/login - 用户登录
  - POST /api/auth/refresh - 刷新令牌

- **用户模块**
  - GET /api/users/me - 获取当前用户信息
  - PUT /api/users/password - 修改密码
  - POST /api/users/api-key - 生成 API Key

- **任务模块**
  - GET /api/tasks - 获取任务列表
  - POST /api/tasks - 创建任务
  - GET /api/tasks/:id - 获取任务详情
  - PUT /api/tasks/:id - 更新任务
  - DELETE /api/tasks/:id - 删除任务

## 🔧 配置说明

### 数据库配置

默认连接到现有的 MySQL Docker 容器：

```go
dsn := "jduser:jdpass123@tcp(localhost:3306)/jd?charset=utf8mb4&parseTime=True&loc=Asia%2FShanghai"
```

### 服务端口

默认端口：5001（避免与 Python 版本冲突）

修改端口：在 `main.go` 中修改 `port` 变量

## 📦 项目结构

```
jd-task-platform-go/
├── main.go                 # 主入口文件
├── deploy/                 # 部署配置
│   └── database/          # 数据库部署配置
├── docs/                   # Swagger 文档（自动生成）
├── internal/
│   ├── handlers/          # 请求处理器
│   │   ├── auth.go       # 认证处理器
│   │   ├── user.go       # 用户处理器
│   │   ├── task.go       # 任务处理器
│   │   └── ...
│   ├── middleware/        # 中间件
│   │   ├── auth.go       # 认证中间件
│   │   └── cors.go       # CORS中间件
│   └── models/            # 数据模型
│       ├── user.go
│       ├── task.go
│       └── device.go
├── pkg/
│   ├── response/          # 统一响应
│   └── utils/             # 工具函数
│       └── jwt.go        # JWT工具
├── go.mod
├── go.sum
└── README.md
```

## 🎯 Swagger 注释示例

在 Go 中添加 Swagger 注释非常简单：

\`\`\`go
// Login 用户登录
// @Summary 用户登录
// @Description 用户登录获取访问令牌
// @Tags 认证模块
// @Accept json
// @Produce json
// @Param request body models.LoginRequest true "登录信息"
// @Success 200 {object} response.Response{data=models.LoginResponse}
// @Failure 401 {object} response.Response
// @Router /auth/login [post]
func (h *AuthHandler) Login(c *gin.Context) {
    // 实现代码...
}
\`\`\`

## 🔑 使用示例

### 1. 用户注册

```bash
curl -X POST http://localhost:5001/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "test123",
    "nickname": "测试用户"
  }'
```

### 2. 用户登录

```bash
curl -X POST http://localhost:5001/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "password": "test123"
  }'
```

### 3. 获取用户信息（需要 Token）

```bash
curl -X GET http://localhost:5001/api/users/me \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

## 🌟 与 Python 版本对比

| 特性 | Python (Flask) | Go (Gin) |
|------|---------------|----------|
| API 文档 | Flasgger | Swaggo ✅ |
| 性能 | 中等 | 高 ✅ |
| 部署 | 需要 Python 环境 | 单个可执行文件 ✅ |
| 并发 | 受限于 GIL | 原生协程 ✅ |
| 启动速度 | 快 | 非常快 ✅ |
| 开发效率 | 高 | 高 ✅ |

## 📖 开发指南

### 添加新的 API 接口

1. 在 `internal/models/` 添加请求/响应模型
2. 在 `internal/handlers/` 添加处理器函数
3. 添加 Swagger 注释
4. 在 `main.go` 注册路由
5. 运行 `swag init` 重新生成文档

### Swagger 注释标签说明

- `@Summary` - 接口简短描述
- `@Description` - 接口详细描述
- `@Tags` - 接口分组标签
- `@Accept` - 接受的内容类型
- `@Produce` - 返回的内容类型
- `@Param` - 参数说明
- `@Success` - 成功响应
- `@Failure` - 失败响应
- `@Router` - 路由路径和方法
- `@Security` - 认证要求

## 🛠️ 故障排查

### 1. Swagger 文档不显示

```bash
# 重新生成文档
swag init

# 确保 docs 文件夹存在
```

### 2. 数据库连接失败

```bash
# 检查 MySQL 容器是否运行
docker ps | grep jd-task-mysql

# 启动容器
cd deploy/database
docker-compose up -d
```

### 3. 端口被占用

修改 `main.go` 中的端口号

## 📝 待完成功能

- [ ] 用户处理器完整实现
- [ ] 任务处理器完整实现
- [ ] 设备处理器完整实现
- [ ] 仪表板处理器实现
- [ ] WebSocket 支持
- [ ] 单元测试
- [ ] Docker 部署

## 🤝 贡献

欢迎提交 Pull Request 或 Issue！

## 📄 License

MIT License

---

**注意**: 此项目使用与 Python 版本相同的 MySQL 数据库，确保数据完全兼容。
