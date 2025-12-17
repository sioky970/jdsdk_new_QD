<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import { NButton, NCard, NDescriptions, NDescriptionsItem, NInput, NModal, NSpace, NTag, useDialog, useMessage } from 'naive-ui';
import { fetchApiKey, generateApiKey, resetApiKey, deleteApiKey } from '@/service/api';

const message = useMessage();
const dialog = useDialog();
const router = useRouter();

// 状态数据
const loading = ref(false);
const apiKey = ref('');
const createdAt = ref('');
const lastUsedAt = ref('');
const showResetModal = ref(false);
const resetLoading = ref(false);

// 加载API Key信息
async function loadApiKey() {
  loading.value = true;
  const { data, error } = await fetchApiKey();
  if (!error && data) {
    apiKey.value = data.api_key || '';
    createdAt.value = data.created_at || '';
    lastUsedAt.value = data.last_used_at || '';
  }
  loading.value = false;
}

// 生成新的API Key
function handleGenerate() {
  dialog.warning({
    title: '确认生成',
    content: apiKey.value 
      ? '检测到已有API Key，生成新的API Key将使旧的API Key失效，确定继续吗？'
      : '确定要生成新的API Key吗？',
    positiveText: '确认生成',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { data, error } = await generateApiKey();
      if (!error && data) {
        message.success('API Key生成成功！请妥善保管');
        loadApiKey();
      }
    }
  });
}

// 重置API Key
function handleReset() {
  showResetModal.value = true;
}

async function confirmReset() {
  resetLoading.value = true;
  const { data, error } = await resetApiKey();
  if (!error && data) {
    message.success('API Key已重置');
    showResetModal.value = false;
    loadApiKey();
  }
  resetLoading.value = false;
}

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

// 复制API Key
function copyApiKey() {
  if (!apiKey.value) {
    message.warning('暂无API Key可复制');
    return;
  }
  
  navigator.clipboard.writeText(apiKey.value).then(() => {
    message.success('API Key已复制到剪贴板');
  }).catch(() => {
    message.error('复制失败，请手动复制');
  });
}

onMounted(() => {
  loadApiKey();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- API Key管理 -->
    <NCard title="API Key 管理" :bordered="false">
      <template #header-extra>
        <NSpace>
          <NButton v-if="!apiKey" type="primary" @click="handleGenerate">
            生成 API Key
          </NButton>
          <template v-if="apiKey">
            <NButton type="warning" @click="handleReset">
              重置
            </NButton>
            <NButton type="error" @click="handleDelete">
              删除
            </NButton>
          </template>
        </NSpace>
      </template>
      
      <div v-if="loading" class="text-center py-8 text-gray-400">
        加载中...
      </div>
      
      <div v-else-if="!apiKey" class="empty-state">
        <div class="empty-icon">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="64" height="64">
            <path d="M7 14c-1.66 0-3 1.34-3 3 0 1.31-1.16 2-2 2 .92 1.22 2.49 2 4 2 2.21 0 4-1.79 4-4 0-1.66-1.34-3-3-3zm13.71-9.37l-1.34-1.34c-.39-.39-1.02-.39-1.41 0L9 12.25 11.75 15l8.96-8.96c.39-.39.39-1.02 0-1.41z"/>
          </svg>
        </div>
        <h3 class="empty-title">暂无 API Key</h3>
        <p class="empty-desc">点击右上角"生成 API Key"按钮创建您的专属API Key</p>
      </div>
      
      <div v-else class="api-key-content">
        <NDescriptions :column="1" bordered label-placement="left">
          <NDescriptionsItem label="API Key">
            <div class="api-key-display">
              <NInput 
                :value="apiKey" 
                readonly 
                placeholder="API Key"
                class="api-key-input"
              >
                <template #suffix>
                  <NButton text @click="copyApiKey">
                    <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
                      <path d="M16 1H4c-1.1 0-2 .9-2 2v14h2V3h12V1zm3 4H8c-1.1 0-2 .9-2 2v14c0 1.1.9 2 2 2h11c1.1 0 2-.9 2-2V7c0-1.1-.9-2-2-2zm0 16H8V7h11v14z"/>
                    </svg>
                  </NButton>
                </template>
              </NInput>
            </div>
          </NDescriptionsItem>
          <NDescriptionsItem label="创建时间">
            {{ createdAt ? new Date(createdAt).toLocaleString('zh-CN') : '-' }}
          </NDescriptionsItem>
          <NDescriptionsItem label="最后使用时间">
            {{ lastUsedAt ? new Date(lastUsedAt).toLocaleString('zh-CN') : '从未使用' }}
          </NDescriptionsItem>
          <NDescriptionsItem label="状态">
            <NTag type="success">正常</NTag>
          </NDescriptionsItem>
        </NDescriptions>
      </div>
    </NCard>

    <!-- 使用说明 -->
    <NCard title="API 使用说明" :bordered="false">
      <template #header-extra>
        <NButton type="primary" @click="router.push('/api-docs')">
          <template #icon>
            <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="16" height="16">
              <path d="M14 2H6c-1.1 0-1.99.9-1.99 2L4 20c0 1.1.89 2 1.99 2H18c1.1 0 2-.9 2-2V8l-6-6zm2 16H8v-2h8v2zm0-4H8v-2h8v2zm-3-5V3.5L18.5 9H13z"/>
            </svg>
          </template>
          查看完整接口文档
        </NButton>
      </template>
      <div class="usage-guide">
        <h4 class="guide-title">接口调用方式</h4>
        <p class="guide-desc">在请求头中添加 API Key 进行身份验证：</p>
        <pre class="code-block">Authorization: Bearer YOUR_API_KEY</pre>
        
        <h4 class="guide-title">支持的API操作</h4>
        <ul class="api-list">
          <li><strong>任务管理</strong>：创建任务、取消任务、修改任务、查询任务列表</li>
          <li><strong>余额查询</strong>：查询京豆余额、充值记录</li>
          <li><strong>明细管理</strong>：查询京豆明细、导出明细报表</li>
        </ul>
        
        <h4 class="guide-title">安全提示</h4>
        <ul class="security-tips">
          <li>请妥善保管您的 API Key，不要泄露给他人</li>
          <li>API Key 具有与您账户相同的操作权限</li>
          <li>如发现 API Key 泄露，请立即重置</li>
          <li>建议定期更换 API Key 以提高安全性</li>
        </ul>
      </div>
    </NCard>

    <!-- 重置确认弹窗 -->
    <NModal
      v-model:show="showResetModal"
      preset="dialog"
      title="确认重置"
      type="warning"
      :mask-closable="false"
    >
      <template #default>
        <p>重置后，旧的 API Key 将立即失效，所有使用旧 API Key 的请求将无法访问。</p>
        <p class="mt-2 text-warning">确定要重置 API Key 吗？</p>
      </template>
      <template #action>
        <NSpace justify="end">
          <NButton @click="showResetModal = false">取消</NButton>
          <NButton type="warning" :loading="resetLoading" @click="confirmReset">
            确认重置
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.empty-state {
  display: flex;
  flex-direction: column;
  align-items: center;
  justify-content: center;
  padding: 80px 20px;
}

.empty-icon {
  color: rgba(var(--base-text-color), 0.3);
  margin-bottom: 24px;
}

.empty-title {
  font-size: 18px;
  font-weight: 600;
  color: rgb(var(--base-text-color));
  margin-bottom: 12px;
}

.empty-desc {
  font-size: 14px;
  color: rgba(var(--base-text-color), 0.6);
}

.api-key-content {
  max-width: 800px;
}

.api-key-display {
  width: 100%;
}

.api-key-input {
  font-family: 'Courier New', monospace;
}

.usage-guide {
  line-height: 1.8;
}

.guide-title {
  font-size: 16px;
  font-weight: 600;
  color: rgb(var(--base-text-color));
  margin: 20px 0 12px 0;
  padding-left: 8px;
  border-left: 3px solid rgb(var(--primary-color));
}

.guide-title:first-child {
  margin-top: 0;
}

.guide-desc {
  color: rgba(var(--base-text-color), 0.7);
  margin-bottom: 12px;
}

.code-block {
  background: #0d1117;
  padding: 12px 16px;
  border-radius: 4px;
  border-left: 3px solid rgb(var(--primary-color));
  border: 1px solid rgba(255, 255, 255, 0.1);
  font-family: 'Courier New', monospace;
  overflow-x: auto;
  color: rgb(var(--primary-color));
}

.api-list,
.security-tips {
  padding-left: 20px;
  color: rgba(var(--base-text-color), 0.8);
}

.api-list li,
.security-tips li {
  margin-bottom: 8px;
}

.api-list li strong {
  color: rgb(var(--base-text-color));
}

.text-warning {
  color: rgb(var(--warning-color));
  font-weight: 600;
}

.mt-2 {
  margin-top: 8px;
}
</style>
