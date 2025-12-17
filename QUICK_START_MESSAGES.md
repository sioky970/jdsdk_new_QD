# API 返回消息优化 - 快速开始

## 🚀 5分钟快速上手

### 1. 导入包

在你的handler文件顶部添加：

```go
import (
    "jd-task-platform-go/internal/constants"
    "jd-task-platform-go/pkg/response"
)
```

### 2. 替换错误消息

**查找并替换**:

```go
// 旧代码
response.Error(c, http.StatusBadRequest, "请求参数错误")

// 新代码
response.Error(c, http.StatusBadRequest, constants.MsgParamError)
```

### 3. 添加成功消息

**查找**:
```go
response.Success(c, data)
```

**替换为**:
```go
response.SuccessWithMsg(c, constants.MsgXxxSuccess, data)
```

### 4. 使用格式化消息

**带参数的消息**:

```go
// 旧代码
msg := fmt.Sprintf("京豆余额不足，需要%d京豆", amount)
response.Error(c, http.StatusBadRequest, msg)

// 新代码
response.Errorf(c, http.StatusBadRequest, 
    constants.MsgTaskBalanceInsufficient, 
    amount, balance)
```

## 📋 常用消息速查

### 认证模块
```go
constants.MsgRegisterSuccess      // "注册成功，欢迎加入！"
constants.MsgLoginSuccess         // "登录成功，欢迎回来！"
constants.MsgLogoutSuccess        // "已安全退出，期待您再次光临！"
constants.MsgLoginFailed          // "用户名或密码不正确，请重新输入"
constants.MsgUsernameExists       // "该用户名已被注册，请换一个用户名试试"
```

### 任务模块
```go
constants.MsgTaskCreated          // "任务创建成功！已消耗 %d 京豆"
constants.MsgTaskCancelled        // "任务已取消，%d 京豆已退还"
constants.MsgTaskTypeUpdated      // "任务类型配置已更新"
constants.MsgTaskNotFound         // "未找到该任务，可能已被取消"
constants.MsgTaskNoPermission     // "您没有权限操作该任务"
```

### 用户模块
```go
constants.MsgPasswordChanged      // "密码修改成功，请使用新密码登录"
constants.MsgAPIKeyGenerated      // "API密钥已生成，请妥善保管"
constants.MsgUserNotFound         // "未找到该用户信息"
```

### 设备模块
```go
constants.MsgDeviceStatusUpdated  // "设备状态已更新"
constants.MsgDeviceTaskAssigned   // "任务分配成功"
constants.MsgDeviceNoTask         // "暂无可执行任务"
```

## 🔥 实战示例

### 示例1：优化登录接口

```go
func (h *AuthHandler) Login(c *gin.Context) {
    var req models.LoginRequest
    if err := c.ShouldBindJSON(&req); err != nil {
        // 旧: response.Error(c, http.StatusBadRequest, "请求参数错误")
        response.Error(c, http.StatusBadRequest, constants.MsgLoginParamError)
        return
    }

    var user models.User
    if err := h.db.Where("username = ?", req.Username).First(&user).Error; err != nil {
        // 旧: response.Error(c, http.StatusUnauthorized, "用户名或密码错误")
        response.Error(c, http.StatusUnauthorized, constants.MsgLoginFailed)
        return
    }

    if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
        response.Error(c, http.StatusUnauthorized, constants.MsgLoginFailed)
        return
    }

    // 生成Token...

    // 旧: response.Success(c, loginResp)
    response.SuccessWithMsg(c, constants.MsgLoginSuccess, loginResp)
}
```

### 示例2：优化任务创建

```go
func (h *TaskHandler) CreateTask(c *gin.Context) {
    // 检查余额
    if user.JingdouBalance < consumeJingdou {
        // 旧: response.Error(c, http.StatusBadRequest, "京豆余额不足")
        response.Errorf(c, http.StatusBadRequest, 
            constants.MsgTaskBalanceInsufficient, 
            consumeJingdou, user.JingdouBalance)
        return
    }

    // 创建任务...

    // 旧: response.Success(c, gin.H{...})
    response.SuccessWithDataAndMsgf(c, gin.H{
        "task_id": task.ID,
        "balance": user.JingdouBalance,
    }, constants.MsgTaskCreated, consumeJingdou)
}
```

## ✅ 检查清单

优化完一个接口后，检查：

- [ ] 所有硬编码字符串都替换为常量
- [ ] 成功操作有友好提示
- [ ] 错误消息清晰明确
- [ ] 代码编译通过
- [ ] 测试验证通过

## 🎯 下一步

1. 选择一个handler文件（推荐从auth.go开始）
2. 按照本指南逐个接口优化
3. 编译并测试
4. 提交代码

完整文档参考：
- `API_MESSAGES_OPTIMIZATION.md` - 详细方案
- `API_MESSAGES_IMPLEMENTATION_GUIDE.md` - 实施指南
- `API_MESSAGES_SUMMARY.md` - 完整总结

---

**开始优化，让API更友好！** 🎉
