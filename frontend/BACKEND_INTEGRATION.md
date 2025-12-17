# JD任务平台 - 前后端对接指南

## 📌 后端服务信息

- **后端地址**: `http://localhost:5001`
- **API前缀**: `/api`
- **响应格式**: 
  ```json
  {
    "code": 0,      // 成功码为 0
    "msg": "操作成功",
    "data": { ... }
  }
  ```

## 🔧 配置步骤

### 1. 修改环境变量配置

在 `.env` 文件中添加以下配置：

```env
# 后端服务地址
VITE_SERVICE_BASE_URL=http://localhost:5001

# 成功响应码（后端返回0表示成功）
VITE_SERVICE_SUCCESS_CODE=0

# 其他服务地址（如需要）
VITE_OTHER_SERVICE_BASE_URL={}
```

### 2. API 接口适配

#### 后端实际接口：

| 功能 | 后端接口 | 方法 | 请求参数 |
|------|---------|------|---------|
| 登录 | `/api/auth/login` | POST | `{username, password}` |
| 注册 | `/api/auth/register` | POST | `{username, password, nickname}` |
| 刷新Token | `/api/auth/refresh` | POST | `{refresh_token}` |
| 获取用户信息 | `/api/users/me` | GET | - |
| 登出 | `/api/auth/logout` | POST | - |

#### 后端响应格式示例：

**登录成功**:
```json
{
  "code": 0,
  "msg": "登录成功，欢迎回来！",
  "data": {
    "id": 2,
    "username": "admin",
    "nickname": "管理员",
    "role": "admin",           // admin 或 common
    "access_token": "eyJ...",
    "refresh_token": "eyJ...",
    "expires": 1765447663000
  }
}
```

### 3. 需要修改的前端文件

#### 3.1 修改登录API (`src/service/api/auth.ts`)

```typescript
// 现有代码
export function fetchLogin(userName: string, password: string) {
  return request<Api.Auth.LoginToken>({
    url: '/auth/login',
    method: 'post',
    data: {
      userName,  // ❌ 需要改为 username
      password
    }
  });
}

// 修改后
export function fetchLogin(username: string, password: string) {
  return request<Api.Auth.LoginToken>({
    url: '/api/auth/login',  // 添加 /api 前缀
    method: 'post',
    data: {
      username,  // ✅ 与后端匹配
      password
    }
  });
}
```

#### 3.2 修改获取用户信息API

```typescript
// 现有代码
export function fetchGetUserInfo() {
  return request<Api.Auth.UserInfo>({ url: '/auth/getUserInfo' });
}

// 修改后
export function fetchGetUserInfo() {
  return request<Api.Auth.UserInfo>({ url: '/api/users/me' });
}
```

#### 3.3 修改刷新Token API

```typescript
// 现有代码
export function fetchRefreshToken(refreshToken: string) {
  return request<Api.Auth.LoginToken>({
    url: '/auth/refreshToken',
    method: 'post',
    data: {
      refreshToken
    }
  });
}

// 修改后
export function fetchRefreshToken(refreshToken: string) {
  return request<Api.Auth.LoginToken>({
    url: '/api/auth/refresh',
    method: 'post',
    data: {
      refresh_token: refreshToken  // 改为下划线格式
    }
  });
}
```

### 4. 类型定义适配

需要修改前端的类型定义以匹配后端响应：

#### 后端响应字段映射：

| 前端期望字段 | 后端实际字段 | 说明 |
|------------|------------|------|
| `userName` | `username` | 用户名 |
| `userId` | `id` | 用户ID |
| `userRole` | `role` | 用户角色 |
| `token` | `access_token` | 访问令牌 |

### 5. 代理配置（开发环境）

如果使用代理，确保 `vite.config.ts` 中的代理配置正确：

```typescript
server: {
  proxy: {
    '/proxy-default': {
      target: 'http://localhost:5001',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/proxy-default/, '')
    }
  }
}
```

## 🎨 主题配置（暗色模式）

根据你的偏好，Soybean Admin 默认启用暗色主题：

在 `src/theme/settings.ts` 中设置：

```typescript
const themeSettings: App.Theme.ThemeSetting = {
  themeScheme: 'dark',  // 默认暗色主题
  // ... 其他配置
};
```

## 🚀 启动项目

### 后端服务
```bash
cd d:\工程\jd-task-platform-go
.\bin\jd-task-platform.exe
```

### 前端服务
```bash
cd d:\工程\测试vue项目\soybean-admin
pnpm dev
```

访问地址：`http://localhost:5173`（或前端配置的端口）

## 📝 测试账号

- **管理员**: `admin` / `admin123`
- **普通用户**: `user001` / `pass123`

## ⚠️ 注意事项

1. **CORS 配置**: 确保后端已配置CORS允许前端访问
2. **Token 存储**: 前端使用 localStorage 存储 token
3. **响应码**: 后端成功码为 `0`，不是 `"0000"`
4. **字段命名**: 后端使用下划线命名（snake_case），前端需要适配

## 🔍 调试建议

1. 打开浏览器开发者工具 -> Network 查看API请求
2. 检查请求URL、请求头、请求体是否正确
3. 查看响应数据格式是否匹配类型定义
4. 使用 `console.log` 在关键位置输出调试信息

---

**下一步**: 需要我帮你自动修改这些文件吗？
