<script setup lang="ts">
import { ref } from 'vue';
import { NCard, NCollapse, NCollapseItem, NTable, NTag, NCode, NDivider, NAlert, NSpace, NButton } from 'naive-ui';

// 当前展开的接口
const expandedNames = ref<string[]>(['task-create']);

// 复制代码
function copyCode(code: string) {
  navigator.clipboard.writeText(code).then(() => {
    window.$message?.success('代码已复制到剪贴板');
  }).catch(() => {
    window.$message?.error('复制失败，请手动复制');
  });
}

// API接口列表
const apiList = [
  {
    key: 'task-create',
    title: '创建单个任务',
    method: 'POST',
    path: '/api/openapi/tasks',
    desc: '使用API Key创建单个任务，支持所有任务类型',
    params: [
      { name: 'task_type', type: 'string', required: true, desc: '任务类型代码（通过 GET /api/openapi/task-types 获取）' },
      { name: 'sku', type: 'string', required: true, desc: '商品SKU' },
      { name: 'shop_name', type: 'string', required: false, desc: '店铺名称' },
      { name: 'keyword', type: 'string', required: false, desc: '搜索关键词（search_browse类型必填）' },
      { name: 'start_time', type: 'string', required: true, desc: '开始执行时间（ISO 8601格式，如：2025-12-12T12:00:00Z）' },
      { name: 'execute_count', type: 'int', required: true, desc: '执行次数' },
      { name: 'priority', type: 'int', required: false, desc: '优先级（默认0）' },
      { name: 'remark', type: 'string', required: false, desc: '备注信息' }
    ],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "task_id": 220,
    "task_type": "search_browse",
    "sku": "100012345678",
    "status": "waiting",
    "consume_jingdou": 15,
    "balance": 985,
    "created_at": "2025-12-12T12:00:00+08:00",
    "message": "任务创建成功"
  }
}`,
    example: `curl -X POST "http://localhost:5001/api/openapi/tasks" \\
  -H "X-API-KEY: sk_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "task_type": "search_browse",
    "sku": "100012345678",
    "shop_name": "测试店铺",
    "keyword": "手机壳",
    "start_time": "2025-12-12T12:00:00Z",
    "execute_count": 5,
    "priority": 1,
    "remark": "API测试任务"
  }'`
  },
  {
    key: 'task-batch',
    title: '批量创建任务',
    method: 'POST',
    path: '/api/openapi/tasks/batch',
    desc: '批量创建多个任务，单次最多100个',
    params: [
      { name: 'tasks', type: 'array', required: true, desc: '任务数组，每个元素结构与单个创建相同' }
    ],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "total_submitted": 2,
    "success_count": 2,
    "failed_count": 0,
    "total_consume": 25,
    "balance": 975,
    "created_tasks": [
      {"task_id": 221, "sku": "111111", "consume_jingdou": 10},
      {"task_id": 222, "sku": "222222", "consume_jingdou": 15}
    ],
    "failed_tasks": [],
    "message": "批量创建完成：成功 2 个，失败 0 个"
  }
}`,
    example: `curl -X POST "http://localhost:5001/api/openapi/tasks/batch" \\
  -H "X-API-KEY: sk_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "tasks": [
      {
        "task_type": "add_to_cart",
        "sku": "111111",
        "shop_name": "店铺1",
        "start_time": "2025-12-12T12:00:00Z",
        "execute_count": 2
      },
      {
        "task_type": "add_to_cart",
        "sku": "222222",
        "shop_name": "店铺2",
        "start_time": "2025-12-12T12:00:00Z",
        "execute_count": 3
      }
    ]
  }'`
  },
  {
    key: 'task-list',
    title: '查询任务列表',
    method: 'GET',
    path: '/api/openapi/tasks',
    desc: '查询任务列表，支持多条件检索和分页',
    params: [
      { name: 'page', type: 'int', required: false, desc: '页码（默认1）' },
      { name: 'page_size', type: 'int', required: false, desc: '每页数量（默认20，最大100）' },
      { name: 'status', type: 'string', required: false, desc: '任务状态：waiting/running/completed/failed/cancelled' },
      { name: 'task_type', type: 'string', required: false, desc: '任务类型' },
      { name: 'sku', type: 'string', required: false, desc: '商品SKU（模糊搜索）' },
      { name: 'shop_name', type: 'string', required: false, desc: '店铺名称（模糊搜索）' },
      { name: 'keyword', type: 'string', required: false, desc: '关键词（模糊搜索）' },
      { name: 'start_date', type: 'string', required: false, desc: '开始日期（YYYY-MM-DD）' },
      { name: 'end_date', type: 'string', required: false, desc: '结束日期（YYYY-MM-DD）' }
    ],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "items": [
      {
        "id": 220,
        "task_type": "search_browse",
        "sku": "100012345678",
        "shop_name": "测试店铺",
        "keyword": "手机壳",
        "start_time": "2025-12-12T20:00:00+08:00",
        "execute_count": 5,
        "executed_count": 0,
        "priority": 1,
        "status": "waiting",
        "consume_jingdou": 15,
        "remark": "API测试任务",
        "created_at": "2025-12-12T11:53:18+08:00",
        "updated_at": "2025-12-12T11:53:18+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1,
    "pages": 1
  }
}`,
    example: `curl -X GET "http://localhost:5001/api/openapi/tasks?status=waiting&page=1&page_size=10" \\
  -H "X-API-KEY: sk_your_api_key"`
  },
  {
    key: 'task-detail',
    title: '查询任务详情',
    method: 'GET',
    path: '/api/openapi/tasks/:id',
    desc: '根据任务ID查询任务详细信息',
    params: [
      { name: 'id', type: 'int', required: true, desc: '任务ID（URL路径参数）' }
    ],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "id": 220,
    "task_type": "search_browse",
    "sku": "100012345678",
    "shop_name": "测试店铺",
    "keyword": "手机壳",
    "start_time": "2025-12-12T20:00:00+08:00",
    "execute_count": 5,
    "executed_count": 0,
    "priority": 1,
    "status": "waiting",
    "consume_jingdou": 15,
    "remark": "API测试任务",
    "created_at": "2025-12-12T11:53:18+08:00",
    "updated_at": "2025-12-12T11:53:18+08:00"
  }
}`,
    example: `curl -X GET "http://localhost:5001/api/openapi/tasks/220" \\
  -H "X-API-KEY: sk_your_api_key"`
  },
  {
    key: 'task-update',
    title: '修改任务',
    method: 'PUT',
    path: '/api/openapi/tasks/:id',
    desc: '修改指定任务（只能修改等待中的任务）',
    params: [
      { name: 'id', type: 'int', required: true, desc: '任务ID（URL路径参数）' },
      { name: 'shop_name', type: 'string', required: false, desc: '店铺名称' },
      { name: 'keyword', type: 'string', required: false, desc: '关键词（仅search_browse类型可修改）' },
      { name: 'priority', type: 'int', required: false, desc: '优先级' },
      { name: 'remark', type: 'string', required: false, desc: '备注' }
    ],
    response: `{
  "code": 0,
  "msg": "任务更新成功",
  "data": {
    "task_id": 220,
    "shop_name": "修改后的店铺",
    "keyword": "新关键词",
    "priority": 5,
    "remark": "修改测试",
    "updated_at": "2025-12-12T12:00:00+08:00"
  }
}`,
    example: `curl -X PUT "http://localhost:5001/api/openapi/tasks/220" \\
  -H "X-API-KEY: sk_your_api_key" \\
  -H "Content-Type: application/json" \\
  -d '{
    "shop_name": "修改后的店铺",
    "priority": 5,
    "remark": "修改测试"
  }'`
  },
  {
    key: 'task-cancel',
    title: '取消任务',
    method: 'POST',
    path: '/api/openapi/tasks/:id/cancel',
    desc: '取消指定任务（只能取消等待中的任务，京豆会自动退还）',
    params: [
      { name: 'id', type: 'int', required: true, desc: '任务ID（URL路径参数）' }
    ],
    response: `{
  "code": 0,
  "msg": "任务取消成功，京豆已退还",
  "data": {
    "task_id": 220,
    "refund_jingdou": 15,
    "balance": 1000
  }
}`,
    example: `curl -X POST "http://localhost:5001/api/openapi/tasks/220/cancel" \\
  -H "X-API-KEY: sk_your_api_key"`
  },
  {
    key: 'task-types',
    title: '获取任务类型',
    method: 'GET',
    path: '/api/openapi/task-types',
    desc: '获取所有可用的任务类型及其价格信息',
    params: [],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "task_types": [
      {
        "type_code": "search_browse",
        "type_name": "关键词搜索浏览",
        "jingdou_price": 3,
        "has_time_limit": true,
        "time_slots": ["09:00-12:00", "14:00-18:00"]
      },
      {
        "type_code": "add_to_cart",
        "type_name": "加入购物车",
        "jingdou_price": 5,
        "has_time_limit": false,
        "time_slots": []
      }
    ],
    "total": 6
  }
}`,
    example: `curl -X GET "http://localhost:5001/api/openapi/task-types" \\
  -H "X-API-KEY: sk_your_api_key"`
  },
  {
    key: 'balance',
    title: '查询京豆余额',
    method: 'GET',
    path: '/api/openapi/balance',
    desc: '查询当前账户的京豆余额',
    params: [],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "user_id": 2,
    "username": "admin",
    "jingdou_balance": 1000,
    "last_updated": "2025-12-12T12:00:00+08:00"
  }
}`,
    example: `curl -X GET "http://localhost:5001/api/openapi/balance" \\
  -H "X-API-KEY: sk_your_api_key"`
  },
  {
    key: 'jingdou-records',
    title: '查询京豆明细',
    method: 'GET',
    path: '/api/openapi/jingdou/records',
    desc: '查询京豆变动明细记录',
    params: [
      { name: 'page', type: 'int', required: false, desc: '页码（默认1）' },
      { name: 'page_size', type: 'int', required: false, desc: '每页数量（默认20，最大100）' },
      { name: 'type', type: 'string', required: false, desc: '类型：task/recharge/refund/admin' },
      { name: 'start_date', type: 'string', required: false, desc: '开始日期（YYYY-MM-DD）' },
      { name: 'end_date', type: 'string', required: false, desc: '结束日期（YYYY-MM-DD）' }
    ],
    response: `{
  "code": 0,
  "msg": "操作成功",
  "data": {
    "items": [
      {
        "id": 100,
        "amount": -15,
        "balance": 985,
        "operation_type": "task",
        "related_id": 220,
        "remark": "API创建任务扣除 - SKU:100012345678",
        "created_at": "2025-12-12T11:53:18+08:00"
      }
    ],
    "page": 1,
    "page_size": 20,
    "total": 1,
    "pages": 1
  }
}`,
    example: `curl -X GET "http://localhost:5001/api/openapi/jingdou/records?page=1&page_size=10" \\
  -H "X-API-KEY: sk_your_api_key"`
  }
];

// 错误码列表
const errorCodes = [
  { code: 400, desc: '请求参数错误', example: '任务类型不存在、参数格式错误、余额不足等' },
  { code: 401, desc: '认证失败', example: '未提供API密钥、API密钥无效' },
  { code: 403, desc: '权限不足', example: '用户已被禁用、无权操作该资源' },
  { code: 404, desc: '资源不存在', example: '任务不存在、用户不存在' },
  { code: 429, desc: '请求过于频繁', example: '超出频率限制（每秒最多2次）' },
  { code: 500, desc: '服务器内部错误', example: '系统异常，请稍后重试' }
];

// 获取方法标签类型
function getMethodType(method: string): 'success' | 'warning' | 'info' | 'error' {
  const map: Record<string, 'success' | 'warning' | 'info' | 'error'> = {
    'GET': 'success',
    'POST': 'info',
    'PUT': 'warning',
    'DELETE': 'error'
  };
  return map[method] || 'info';
}
</script>

<template>
  <div class="api-docs-container">
    <!-- 概述 -->
    <NCard title="API 接口文档" :bordered="false" class="mb-4">
      <NAlert type="info" :show-icon="true" class="mb-4">
        <template #header>认证方式</template>
        所有接口需要在请求头中携带 <code>X-API-KEY</code> 进行身份验证
      </NAlert>
      
      <div class="auth-example">
        <div class="code-header">
          <span>请求头示例</span>
          <NButton text size="small" @click="copyCode('X-API-KEY: sk_your_api_key_here')">
            复制
          </NButton>
        </div>
        <pre class="code-block">X-API-KEY: sk_your_api_key_here</pre>
      </div>

      <NDivider />

      <NSpace vertical :size="12">
        <div class="info-item">
          <strong>基础URL：</strong>
          <code>http://your-domain:5001/api/openapi</code>
        </div>
        <div class="info-item">
          <strong>频率限制：</strong>
          <NTag type="warning">每秒最多 2 次请求</NTag>
        </div>
        <div class="info-item">
          <strong>数据格式：</strong>
          <code>Content-Type: application/json</code>
        </div>
      </NSpace>
    </NCard>

    <!-- 接口列表 -->
    <NCard title="接口列表" :bordered="false" class="mb-4">
      <NCollapse v-model:expanded-names="expandedNames" accordion>
        <NCollapseItem v-for="api in apiList" :key="api.key" :name="api.key">
          <template #header>
            <div class="api-header">
              <NTag :type="getMethodType(api.method)" size="small" class="method-tag">
                {{ api.method }}
              </NTag>
              <span class="api-path">{{ api.path }}</span>
              <span class="api-title">{{ api.title }}</span>
            </div>
          </template>

          <div class="api-content">
            <p class="api-desc">{{ api.desc }}</p>

            <!-- 请求参数 -->
            <div v-if="api.params.length > 0" class="section">
              <h4 class="section-title">请求参数</h4>
              <NTable :bordered="true" :single-line="false" size="small">
                <thead>
                  <tr>
                    <th style="width: 150px;">参数名</th>
                    <th style="width: 80px;">类型</th>
                    <th style="width: 60px;">必填</th>
                    <th>说明</th>
                  </tr>
                </thead>
                <tbody>
                  <tr v-for="param in api.params" :key="param.name">
                    <td><code>{{ param.name }}</code></td>
                    <td>{{ param.type }}</td>
                    <td>
                      <NTag v-if="param.required" type="error" size="small">是</NTag>
                      <NTag v-else type="default" size="small">否</NTag>
                    </td>
                    <td>{{ param.desc }}</td>
                  </tr>
                </tbody>
              </NTable>
            </div>
            <div v-else class="section">
              <h4 class="section-title">请求参数</h4>
              <p class="no-params">无需请求参数</p>
            </div>

            <!-- 响应示例 -->
            <div class="section">
              <h4 class="section-title">成功响应示例</h4>
              <div class="code-header">
                <span>JSON</span>
                <NButton text size="small" @click="copyCode(api.response)">复制</NButton>
              </div>
              <pre class="code-block response-block">{{ api.response }}</pre>
            </div>

            <!-- 调用示例 -->
            <div class="section">
              <h4 class="section-title">调用示例 (cURL)</h4>
              <div class="code-header">
                <span>Shell</span>
                <NButton text size="small" @click="copyCode(api.example)">复制</NButton>
              </div>
              <pre class="code-block example-block">{{ api.example }}</pre>
            </div>
          </div>
        </NCollapseItem>
      </NCollapse>
    </NCard>

    <!-- 错误码说明 -->
    <NCard title="错误码说明" :bordered="false" class="mb-4">
      <NTable :bordered="true" :single-line="false" size="small">
        <thead>
          <tr>
            <th style="width: 100px;">错误码</th>
            <th style="width: 150px;">说明</th>
            <th>示例场景</th>
          </tr>
        </thead>
        <tbody>
          <tr v-for="err in errorCodes" :key="err.code">
            <td><NTag type="error">{{ err.code }}</NTag></td>
            <td>{{ err.desc }}</td>
            <td>{{ err.example }}</td>
          </tr>
        </tbody>
      </NTable>

      <NDivider />

      <h4 class="section-title">错误响应格式</h4>
      <pre class="code-block">{
  "code": 400,
  "msg": "京豆余额不足：创建此任务需要 50 京豆，您当前余额为 30 京豆，请先充值"
}</pre>
    </NCard>

    <!-- 注意事项 -->
    <NCard title="注意事项" :bordered="false">
      <NAlert type="warning" :show-icon="true" class="mb-3">
        <template #header>安全提示</template>
        <ul class="tips-list">
          <li>请妥善保管您的 API Key，不要泄露给他人</li>
          <li>API Key 具有与您账户相同的操作权限</li>
          <li>如发现 API Key 泄露，请立即重置</li>
        </ul>
      </NAlert>

      <NAlert type="info" :show-icon="true">
        <template #header>使用限制</template>
        <ul class="tips-list">
          <li>频率限制：每秒最多 2 次请求，超出将返回 429 错误</li>
          <li>批量创建：单次最多 100 个任务</li>
          <li>任务修改/取消：仅支持"等待中"状态的任务</li>
          <li>某些任务类型有时间段限制，请先查询任务类型确认</li>
        </ul>
      </NAlert>
    </NCard>
  </div>
</template>

<style scoped>
.api-docs-container {
  padding: 0 0 20px;
}

.mb-4 {
  margin-bottom: 16px;
}

.mb-3 {
  margin-bottom: 12px;
}

.auth-example {
  margin-top: 16px;
}

.code-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 8px 12px;
  background: rgb(var(--container-bg-color));
  border-radius: 4px 4px 0 0;
  border: 1px solid rgba(var(--base-text-color), 0.1);
  border-bottom: none;
  font-size: 12px;
  color: rgba(var(--base-text-color), 0.6);
}

.code-block {
  margin: 0;
  padding: 16px;
  background: #1e1e1e;
  color: #d4d4d4;
  border-radius: 0 0 4px 4px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  line-height: 1.5;
  overflow-x: auto;
  white-space: pre-wrap;
  word-break: break-all;
}

.info-item {
  display: flex;
  align-items: center;
  gap: 8px;
  color: rgb(var(--base-text-color));
}

.info-item code {
  background: rgba(var(--primary-color), 0.1);
  padding: 2px 8px;
  border-radius: 4px;
  font-family: 'Consolas', monospace;
  color: rgb(var(--primary-color));
}

.api-header {
  display: flex;
  align-items: center;
  gap: 12px;
  flex-wrap: wrap;
}

.method-tag {
  min-width: 50px;
  text-align: center;
}

.api-path {
  font-family: 'Consolas', monospace;
  color: rgb(var(--primary-color));
  font-weight: 500;
}

.api-title {
  color: rgb(var(--base-text-color));
}

.api-content {
  padding: 16px 0;
}

.api-desc {
  color: rgba(var(--base-text-color), 0.7);
  margin-bottom: 20px;
  padding-left: 4px;
}

.section {
  margin-bottom: 24px;
}

.section-title {
  font-size: 14px;
  font-weight: 600;
  color: rgb(var(--base-text-color));
  margin-bottom: 12px;
  padding-left: 8px;
  border-left: 3px solid rgb(var(--primary-color));
}

.no-params {
  color: rgba(var(--base-text-color), 0.5);
  padding-left: 4px;
}

.response-block {
  max-height: 300px;
}

.example-block {
  max-height: 200px;
}

.tips-list {
  margin: 0;
  padding-left: 20px;
  color: rgba(var(--base-text-color), 0.8);
}

.tips-list li {
  margin-bottom: 4px;
}

.tips-list li:last-child {
  margin-bottom: 0;
}

/* 表格样式适配主题 */
:deep(.n-table) {
  font-size: 13px;
}

:deep(.n-table code) {
  background: rgba(var(--primary-color), 0.1);
  padding: 2px 6px;
  border-radius: 3px;
  font-family: 'Consolas', monospace;
  color: rgb(var(--primary-color));
  font-size: 12px;
}

/* 折叠面板样式 */
:deep(.n-collapse-item__header-main) {
  font-weight: 500;
}

/* 代码块暗色主题 */
.code-block {
  background: #0d1117;
  border: 1px solid rgba(255, 255, 255, 0.1);
}
</style>
