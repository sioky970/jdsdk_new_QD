# JD任务平台 - 前端部署指南

## 📋 环境变量配置说明

项目提供了三个环境配置文件，方便在不同环境下使用：

### 1. `.env` - 基础配置
- 所有环境共享的基础配置
- 包含默认值和开发环境配置

### 2. `.env.development` - 开发环境配置
- 本地开发时使用
- 后端地址: `http://localhost:5001`
- 启用开发工具和调试功能

### 3. `.env.production` - 生产环境配置
- 生产构建时使用
- **需要修改后端地址为实际域名**
- 禁用调试功能，优化性能

### 4. `.env.production.example` - 生产配置示例
- 提供三种部署方案参考
- 复制并修改为 `.env.production` 使用

---

## 🚀 开发环境使用

### 1. 安装依赖
```bash
pnpm install
```

### 2. 启动开发服务器
```bash
pnpm dev
```

默认访问地址: http://localhost:9528/

开发环境会自动使用 `.env.development` 中的配置。

---

## 📦 生产环境部署

### 方案一：独立 API 域名（推荐）

**适用场景**: 前后端完全分离，后端有独立域名

#### 1. 修改 `.env.production`
```env
VITE_SERVICE_BASE_URL=https://api.yourdomain.com
VITE_SERVICE_API_PREFIX=/api
```

#### 2. 构建项目
```bash
pnpm build
```

#### 3. 部署 `dist` 目录
将 `dist` 目录上传到服务器（如 Nginx、Apache）

#### 4. 后端配置
确保后端服务运行在 `https://api.yourdomain.com:443`

#### 5. CORS 配置
后端需要允许前端域名的跨域请求：
```go
// 示例: Go Gin 框架
router.Use(cors.New(cors.Config{
    AllowOrigins:     []string{"https://yourdomain.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
    AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
    AllowCredentials: true,
}))
```

---

### 方案二：同域名部署

**适用场景**: 前后端部署在同一服务器，共享域名

#### 1. 修改 `.env.production`
```env
VITE_SERVICE_BASE_URL=https://yourdomain.com
VITE_SERVICE_API_PREFIX=/api
```

#### 2. 构建项目
```bash
pnpm build
```

#### 3. Nginx 配置示例
```nginx
server {
    listen 80;
    server_name yourdomain.com;

    # 前端静态文件
    location / {
        root /var/www/jd-task-platform/frontend;
        try_files $uri $uri/ /index.html;
    }

    # 后端 API 代理
    location /api {
        proxy_pass http://localhost:5001;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
    }
}
```

---

### 方案三：子域名部署

**适用场景**: 使用子域名访问后端 API

#### 1. 修改 `.env.production`
```env
VITE_SERVICE_BASE_URL=https://task.yourdomain.com
VITE_SERVICE_API_PREFIX=/api
```

#### 2. 构建和部署
同方案一

---

## 🔧 环境变量说明

### 关键配置项

| 变量名 | 说明 | 开发环境 | 生产环境 |
|--------|------|----------|----------|
| `VITE_SERVICE_BASE_URL` | 后端服务地址 | `http://localhost:5001` | `https://yourdomain.com` |
| `VITE_SERVICE_API_PREFIX` | API 路径前缀 | `/api` | `/api` |
| `VITE_SERVICE_SUCCESS_CODE` | 成功响应码 | `0` | `0` |
| `VITE_HTTP_PROXY` | 是否启用代理 | `Y` | `N` |
| `VITE_SOURCE_MAP` | 是否生成 sourcemap | `Y` | `N` |
| `VITE_ROUTER_HISTORY_MODE` | 路由模式 | `history` | `history` |

---

## 📝 部署检查清单

### 发布前检查

- [ ] 修改 `.env.production` 中的 `VITE_SERVICE_BASE_URL` 为实际域名
- [ ] 确认后端服务正常运行
- [ ] 确认 CORS 配置正确
- [ ] 执行 `pnpm build` 构建成功
- [ ] 检查 `dist` 目录是否生成

### 部署后验证

- [ ] 访问前端页面是否正常加载
- [ ] 登录功能是否正常
- [ ] 浏览器控制台无 404 或 CORS 错误
- [ ] API 请求是否正确发送到后端
- [ ] Token 认证是否正常工作

---

## 🐛 常见问题

### 1. API 请求 404 错误
**原因**: 后端地址配置错误或后端服务未启动

**解决**:
- 检查 `.env.production` 中的 `VITE_SERVICE_BASE_URL`
- 确认后端服务已启动并监听正确端口
- 检查 Nginx/Apache 代理配置

### 2. CORS 跨域错误
**原因**: 后端未允许前端域名的跨域请求

**解决**:
- 在后端添加 CORS 中间件
- 允许前端域名和必要的请求头

### 3. 刷新页面 404 错误
**原因**: 使用 history 模式但服务器未配置正确

**解决**:
- Nginx 添加 `try_files $uri $uri/ /index.html;`
- Apache 添加 `.htaccess` 重定向规则

### 4. 静态资源 404
**原因**: 部署路径与配置不匹配

**解决**:
- 如果部署在子目录，修改 `VITE_BASE_URL=/subdir/`
- 确保服务器静态文件路径正确

---

## 🔐 安全建议

1. **HTTPS**: 生产环境必须使用 HTTPS
2. **环境变量**: 不要在代码中硬编码敏感信息
3. **Token 安全**: 
   - Token 存储在 localStorage，XSS 风险需注意
   - 建议启用 HttpOnly Cookie（需后端配合）
4. **API 认证**: 所有敏感接口必须验证 Token
5. **CORS 限制**: 仅允许可信域名的跨域请求

---

## 📞 技术支持

如遇到部署问题，请检查：
1. 浏览器控制台 (F12) 的错误信息
2. 后端服务日志
3. Nginx/Apache 错误日志

常见日志位置：
- Nginx: `/var/log/nginx/error.log`
- Apache: `/var/log/apache2/error.log`
