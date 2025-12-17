# API Key 删除功能 + UI 优化

## 📋 更新内容

### 1️⃣ 新增删除API Key功能

#### 后端实现

**新增接口**: `DELETE /api/apikey`

**文件**: `internal/handlers/apikey.go`

```go
// DeleteAPIKey 删除API密钥（JWT认证）
// @Summary 删除API密钥
// @Description 删除用户的API密钥
// @Tags API密钥
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} response.Response{data=object}
// @Router /apikey [delete]
func (h *APIKeyHandler) DeleteAPIKey(c *gin.Context) {
    userID, _ := c.Get("user_id")

    var user models.User
    if err := h.db.First(&user, userID).Error; err != nil {
        response.Error(c, http.StatusNotFound, "用户不存在")
        return
    }

    // 删除API密钥
    user.ApiKey = ""
    user.ApiKeyCreatedAt = nil
    user.ApiKeyLastUsedAt = nil

    if err := h.db.Save(&user).Error; err != nil {
        response.Error(c, http.StatusInternalServerError, "删除API密钥失败")
        return
    }

    response.SuccessWithMsg(c, "API密钥已删除", nil)
}
```

**删除逻辑**：
- ✅ 清空 `api_key` 字段
- ✅ 清空 `api_key_created_at` 时间戳
- ✅ 清空 `api_key_last_used_at` 时间戳
- ✅ 返回成功消息

**路由配置**: `main.go`
```go
apikey.DELETE("", apikeyHandler.DeleteAPIKey)
```

#### 前端实现

**新增接口**: `src/service/api/apikey.ts`

```typescript
/**
 * 删除API Key
 */
export function deleteApiKey() {
  return request({
    url: '/api/apikey',
    method: 'delete'
  });
}
```

**页面功能**: `src/views/api-key/index.vue`

```typescript
// 删除API Key
function handleDelete() {
  dialog.warning({
    title: '确认删除',
    content: '删除后，您的API Key将无法使用，所有使用该API Key的请求将被拒绝。确定要删除吗？',
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { error } = await deleteApiKey();
      if (!error) {
        message.success('API Key已删除');
        loadApiKey();
      }
    }
  });
}
```

### 2️⃣ UI 优化

#### 改进1：API Key明文显示

**修改前**：
```vue
<NInput 
  :value="apiKey" 
  type="password"  <!-- 密码形式 -->
  readonly 
/>
```

**修改后**：
```vue
<NInput 
  :value="apiKey" 
  readonly  <!-- 明文显示 -->
/>
```

**效果**：用户可以直接看到完整的API Key，方便复制和验证

#### 改进2：按钮逻辑优化

**修改前**：
```vue
<NButton v-if="!apiKey" type="primary">生成 API Key</NButton>
<NButton v-if="apiKey" type="warning">重置 API Key</NButton>
```

**修改后**：
```vue
<!-- 未生成时：显示"生成"按钮 -->
<NButton v-if="!apiKey" type="primary">生成 API Key</NButton>

<!-- 已生成时：显示"重置"和"删除"按钮 -->
<template v-if="apiKey">
  <NButton type="warning">重置</NButton>
  <NButton type="error">删除</NButton>
</template>
```

**按钮状态图**：

| 状态 | 显示按钮 | 颜色 | 功能 |
|------|---------|------|------|
| 未生成 | 生成 API Key | 蓝色(primary) | 生成新的API Key |
| 已生成 | 重置 | 黄色(warning) | 重新生成API Key |
| 已生成 | 删除 | 红色(error) | 删除API Key |

#### 改进3：按钮文案简化

**修改前**：
- 生成 API Key
- 重置 API Key

**修改后**：
- 生成 API Key
- 重置 ✅（简化文案）
- 删除 ✅（新增）

## 🔄 完整的状态流转

```
┌─────────────┐
│  未生成状态   │
│  api_key=""  │
└──────┬──────┘
       │
       │ [点击"生成"]
       ▼
┌─────────────┐
│  已生成状态   │
│ api_key="sk_"│
│ created_at   │
│ last_used_at │
└──────┬──────┘
       │
       ├─[点击"重置"]──► 生成新的key，旧key失效
       │
       └─[点击"删除"]──► 返回未生成状态
```

## 🎯 API接口清单

| 接口 | 方法 | 说明 | 状态 |
|------|------|------|------|
| `/api/apikey` | GET | 获取API Key信息 | ✅ |
| `/api/apikey/generate` | POST | 生成新的API Key | ✅ |
| `/api/apikey/reset` | POST | 重置API Key | ✅ |
| `/api/apikey` | DELETE | **删除API Key** | ✅ 新增 |
| `/api/apikey/logs` | GET | 获取API调用记录 | ✅ |

## ⚠️ 删除功能注意事项

### 1. 数据清理
删除操作会清空以下字段：
- `api_key` → `""`
- `api_key_created_at` → `null`
- `api_key_last_used_at` → `null`

### 2. 安全确认
- ✅ 删除前需要二次确认
- ✅ 提示用户删除后果（所有请求将被拒绝）
- ✅ 使用警告样式对话框（warning类型）

### 3. 前端状态同步
删除成功后：
- ✅ 自动刷新页面数据
- ✅ 显示空状态页面
- ✅ 按钮切换为"生成"模式

## 🎨 UI 展示效果

### 未生成状态
```
┌─────────────────────────────────────┐
│  API Key 管理         [生成 API Key] │
├─────────────────────────────────────┤
│                                     │
│         🔑                          │
│     暂无 API Key                    │
│  点击右上角"生成 API Key"按钮创建    │
│                                     │
└─────────────────────────────────────┘
```

### 已生成状态
```
┌─────────────────────────────────────┐
│  API Key 管理    [重置]  [删除]     │
├─────────────────────────────────────┤
│  API Key: sk_1234567890abcdef... 📋 │
│  创建时间: 2025-12-12 10:00:00      │
│  最后使用: 2025-12-12 10:30:00      │
│  状态: 正常 ✓                       │
└─────────────────────────────────────┘
```

## 📝 测试场景

### 场景1：生成API Key
1. 用户首次进入页面 → 显示空状态
2. 点击"生成 API Key" → 弹出确认对话框
3. 确认生成 → 显示新生成的API Key（明文）

### 场景2：重置API Key
1. 已有API Key → 点击"重置"按钮
2. 弹出警告对话框 → 提示旧key将失效
3. 确认重置 → 生成新的key，时间戳更新

### 场景3：删除API Key
1. 已有API Key → 点击"删除"按钮
2. 弹出警告对话框 → 提示删除后无法使用
3. 确认删除 → 回到未生成状态

### 场景4：复制API Key
1. 已有API Key → 点击复制图标
2. API Key复制到剪贴板
3. 显示成功提示

## ✅ 开发完成清单

- [x] 后端删除接口实现
- [x] 后端路由配置
- [x] 前端删除接口封装
- [x] 前端删除功能实现
- [x] API Key改为明文显示
- [x] 按钮逻辑优化（未生成/已生成）
- [x] 按钮文案简化
- [x] 二次确认对话框
- [x] Go服务器重启
- [x] 功能测试验证

## 🎉 总结

本次更新完成了以下优化：

1. **功能增强**：新增API Key删除功能
2. **UI优化**：明文显示API Key，方便查看和复制
3. **交互优化**：智能按钮显示（未生成显示"生成"，已生成显示"重置"+"删除"）
4. **安全提升**：删除前二次确认，防止误操作

所有功能已测试通过，可以在前端页面正常使用！

---

**开发日期**：2025-12-12  
**开发者**：AI Assistant
