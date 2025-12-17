# 任务类型管理 API 文档

## 概述
任务类型系统采用预设模式，系统预置6种任务类型，管理员只能修改配置参数，不能增删任务类型。

## API 接口

### 1. 获取任务类型列表

**接口**: `GET /api/tasks/types`

**认证**: 需要Bearer Token

**描述**: 获取所有任务类型配置，包括价格、启用状态和时间段限制

**响应示例**:
```json
{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "task_types": [
      {
        "id": 7,
        "type_code": "browse",
        "type_name": "浏览任务",
        "jingdou_price": 3,
        "is_active": true,
        "is_system_preset": true,
        "has_time_limit": true,
        "time_slot1_start": "08:00",
        "time_slot1_end": "09:00",
        "time_slots": ["08:00-09:00"],
        "created_at": "2025-12-11T15:50:16+08:00",
        "updated_at": "2025-12-11T15:55:26+08:00"
      },
      {
        "id": 8,
        "type_code": "search_browse",
        "type_name": "关键词搜索浏览任务",
        "jingdou_price": 3,
        "is_active": true,
        "is_system_preset": true,
        "has_time_limit": true,
        "time_slot1_start": "08:00",
        "time_slot1_end": "12:00",
        "time_slot2_start": "14:00",
        "time_slot2_end": "18:00",
        "time_slots": ["08:00-12:00", "14:00-18:00"],
        "created_at": "2025-12-11T15:50:16+08:00",
        "updated_at": "2025-12-11T15:50:16+08:00"
      }
    ],
    "total": 6
  }
}
```

**字段说明**:

| 字段 | 类型 | 说明 |
|------|------|------|
| id | int | 任务类型ID |
| type_code | string | 任务类型代码（唯一标识） |
| type_name | string | 任务类型名称 |
| jingdou_price | int | 单次任务价格（京豆） |
| is_active | bool | 是否启用 |
| is_system_preset | bool | 是否系统预设（固定为true） |
| has_time_limit | bool | 是否有时间限制 |
| time_slot1_start | string | 时间段1开始时间（HH:MM格式） |
| time_slot1_end | string | 时间段1结束时间（HH:MM格式） |
| time_slot2_start | string | 时间段2开始时间（可选） |
| time_slot2_end | string | 时间段2结束时间（可选） |
| time_slots | array | 格式化的时间段数组（便于前端展示） |
| created_at | string | 创建时间（ISO 8601格式） |
| updated_at | string | 更新时间（ISO 8601格式） |

---

### 2. 创建任务类型（已禁用）

**接口**: `POST /api/tasks/types`

**认证**: 需要Bearer Token（管理员）

**描述**: 此接口已禁用，系统预设所有任务类型，不允许创建新类型

**响应**:
```json
{
  "code": 403,
  "msg": "系统已预设所有任务类型，不允许创建新类型"
}
```

---

### 3. 更新任务类型配置

**接口**: `PUT /api/tasks/types/:id`

**认证**: 需要Bearer Token（管理员）

**描述**: 更新任务类型配置（仅允许修改价格、启用状态和时间段）

**请求参数**:
```json
{
  "jingdou_price": 5,
  "is_active": true,
  "time_slot1_start": "08:00",
  "time_slot1_end": "12:00",
  "time_slot2_start": "14:00",
  "time_slot2_end": "18:00"
}
```

**参数说明**:

| 参数 | 类型 | 必填 | 说明 |
|------|------|------|------|
| jingdou_price | int | 否 | 任务单价（京豆） |
| is_active | bool | 否 | 是否启用 |
| time_slot1_start | string | 否 | 时间段1开始（HH:MM） |
| time_slot1_end | string | 否 | 时间段1结束（HH:MM） |
| time_slot2_start | string | 否 | 时间段2开始（HH:MM，可选） |
| time_slot2_end | string | 否 | 时间段2结束（HH:MM，可选） |

**响应示例**:
```json
{
  "code": 0,
  "msg": "任务类型更新成功"
}
```

**限制说明**:
- ✅ 可以修改：价格、启用状态、时间段配置
- ❌ 不可修改：任务类型代码（type_code）、任务类型名称（type_name）
- ❌ 不可删除：系统预设任务类型无法删除

---

## 预设任务类型

系统预置6种任务类型：

| ID | 代码 | 名称 | 默认价格 | 默认时间段 |
|----|------|------|----------|------------|
| 7 | browse | 浏览任务 | 2京豆 | 08:00-12:00, 14:00-18:00 |
| 8 | search_browse | 关键词搜索浏览任务 | 3京豆 | 08:00-12:00, 14:00-18:00 |
| 9 | add_to_cart | 加入购物车任务 | 5京豆 | 08:00-12:00, 14:00-18:00 |
| 10 | follow_shop | 关注店铺任务 | 4京豆 | 08:00-12:00, 14:00-18:00 |
| 11 | follow_product | 收藏商品任务 | 4京豆 | 08:00-12:00, 14:00-18:00 |
| 12 | purchase | 购买商品任务 | 10京豆 | 08:00-12:00, 14:00-18:00 |

---

## 时间段限制说明

### 时间段配置规则

1. **最多两个时间段**: 每个任务类型最多可以配置两个时间段
2. **HH:MM格式**: 时间必须使用24小时制，格式为 `HH:MM`（如 `08:00`, `14:30`）
3. **可选配置**: 
   - 如果不配置时间段（time_slot1为null），则任务全天可创建
   - 如果只配置time_slot1，则只有一个允许时间段
   - 如果同时配置time_slot1和time_slot2，则有两个允许时间段

### 创建任务时的验证

当用户创建任务时，系统会：
1. 检查任务类型是否启用（is_active）
2. 检查当前时间是否在允许的时间段内
3. 如果不在允许时间段，返回错误信息

**错误示例**:
```json
{
  "code": 400,
  "msg": "当前时间不在允许创建任务的时间段内（08:00-09:00）"
}
```

---

## 使用示例

### PowerShell 示例

```powershell
# 1. 登录获取Token
$loginResp = Invoke-RestMethod -Uri "http://localhost:5001/api/auth/login" `
    -Method POST `
    -Body '{"username":"admin","password":"admin123"}' `
    -ContentType "application/json"
$token = $loginResp.data.access_token

# 2. 获取任务类型列表
$headers = @{
    "Authorization" = "Bearer $token"
}
$types = Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/types" `
    -Headers $headers

# 显示结果
$types.data.task_types | ForEach-Object {
    Write-Host "[$($_.id)] $($_.type_name) - $($_.jingdou_price)京豆"
    Write-Host "  时间段: $($_.time_slots -join ', ')"
}

# 3. 更新任务类型配置
$updateHeaders = @{
    "Authorization" = "Bearer $token"
    "Content-Type" = "application/json"
}
$updateBody = @{
    jingdou_price = 5
    is_active = $true
    time_slot1_start = "09:00"
    time_slot1_end = "11:00"
    time_slot2_start = "15:00"
    time_slot2_end = "17:00"
} | ConvertTo-Json

Invoke-RestMethod -Uri "http://localhost:5001/api/tasks/types/7" `
    -Method PUT `
    -Headers $updateHeaders `
    -Body $updateBody
```

### Python 示例

```python
import requests

# 1. 登录
login_resp = requests.post(
    'http://localhost:5001/api/auth/login',
    json={'username': 'admin', 'password': 'admin123'}
)
token = login_resp.json()['data']['access_token']

# 2. 获取任务类型列表
headers = {'Authorization': f'Bearer {token}'}
types_resp = requests.get(
    'http://localhost:5001/api/tasks/types',
    headers=headers
)
task_types = types_resp.json()['data']['task_types']

# 显示结果
for t in task_types:
    print(f"[{t['id']}] {t['type_name']} - {t['jingdou_price']}京豆")
    print(f"  时间段: {', '.join(t['time_slots'])}")
    print(f"  启用: {t['is_active']}")
    print()

# 3. 更新任务类型
update_resp = requests.put(
    'http://localhost:5001/api/tasks/types/7',
    headers={'Authorization': f'Bearer {token}', 'Content-Type': 'application/json'},
    json={
        'jingdou_price': 5,
        'is_active': True,
        'time_slot1_start': '09:00',
        'time_slot1_end': '11:00',
        'time_slot2_start': '15:00',
        'time_slot2_end': '17:00'
    }
)
print(update_resp.json()['msg'])
```

---

## 前端集成建议

### 1. 展示任务类型列表

```javascript
// 获取任务类型列表
async function getTaskTypes() {
  const response = await fetch('/api/tasks/types', {
    headers: {
      'Authorization': `Bearer ${token}`
    }
  });
  const data = await response.json();
  return data.data.task_types;
}

// 渲染到页面
const taskTypes = await getTaskTypes();
taskTypes.forEach(type => {
  console.log(`${type.type_name}:`);
  console.log(`  价格: ${type.jingdou_price}京豆`);
  console.log(`  时间段: ${type.time_slots.join(', ')}`);
  console.log(`  状态: ${type.is_active ? '启用' : '禁用'}`);
});
```

### 2. 任务类型配置表单

```html
<!-- 任务类型配置表单 -->
<form id="taskTypeForm">
  <div>
    <label>任务类型名称</label>
    <input type="text" readonly value="浏览任务" />
  </div>
  
  <div>
    <label>任务单价（京豆）</label>
    <input type="number" name="jingdou_price" value="3" />
  </div>
  
  <div>
    <label>是否启用</label>
    <input type="checkbox" name="is_active" checked />
  </div>
  
  <div>
    <label>时间段1</label>
    <input type="time" name="time_slot1_start" value="08:00" />
    -
    <input type="time" name="time_slot1_end" value="12:00" />
  </div>
  
  <div>
    <label>时间段2（可选）</label>
    <input type="time" name="time_slot2_start" value="14:00" />
    -
    <input type="time" name="time_slot2_end" value="18:00" />
  </div>
  
  <button type="submit">保存配置</button>
</form>
```

### 3. 时间段显示组件

```vue
<template>
  <div class="time-slots">
    <div v-if="taskType.has_time_limit" class="time-limit">
      <i class="icon-clock"></i>
      <span v-for="slot in taskType.time_slots" :key="slot">
        {{ slot }}
      </span>
    </div>
    <div v-else class="no-limit">
      <i class="icon-all-day"></i>
      全天可创建
    </div>
  </div>
</template>
```

---

## 注意事项

1. **系统预设类型不可删除**: 所有6种任务类型都是系统预设，无法删除
2. **名称不可修改**: 任务类型的名称和代码是固定的，不允许修改
3. **时间段格式**: 必须使用 `HH:MM` 格式，如 `08:00`, `14:30`
4. **时间段验证**: 创建任务时会实时验证当前时间是否在允许范围内
5. **管理员权限**: 只有管理员可以修改任务类型配置
6. **前端缓存**: 建议前端缓存任务类型列表，减少API调用

---

## 更新日志

### 2025-12-11
- ✅ 实现6种预设任务类型
- ✅ 添加时间段限制功能
- ✅ 禁止创建/删除任务类型
- ✅ 优化API响应格式
- ✅ 添加 `has_time_limit` 字段
- ✅ 添加 `time_slots` 数组字段（便于前端展示）
- ✅ 添加 `total` 字段（返回任务类型总数）

---

**文档版本**: v1.0  
**最后更新**: 2025-12-11  
**维护者**: AI Assistant
