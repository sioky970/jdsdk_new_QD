<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { useRouter } from 'vue-router';
import {
  NButton,
  NCard,
  NDataTable,
  NDescriptions,
  NDescriptionsItem,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NPopconfirm,
  NSelect,
  NSpace,
  NSwitch,
  NTag,
  useDialog,
  useMessage
} from 'naive-ui';
import type { DataTableColumn } from 'naive-ui';
import {
  fetchAdminUsers,
  createAdminUser,
  updateAdminUser,
  deleteAdminUser,
  adjustUserJingdou,
  getUserApiKey,
  resetUserApiKey,
  deleteUserApiKey
} from '@/service/api';
import type { AdminUser, CreateUserRequest, UpdateUserRequest } from '@/service/api';

const message = useMessage();
const dialog = useDialog();
const router = useRouter();

// 状态数据
const loading = ref(false);
const users = ref<AdminUser[]>([]);
const total = ref(0);

// 筛选条件
const filters = ref({
  search: '',
  page: 1,
  perPage: 20
});

// 分页信息
const pagination = computed(() => ({
  page: filters.value.page,
  pageSize: filters.value.perPage,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  prefix: () => `共 ${total.value} 条`
}));

// 创建用户弹窗
const showCreateModal = ref(false);
const createLoading = ref(false);
const createForm = ref<CreateUserRequest>({
  username: '',
  password: '',
  nickname: '',
  role: 'common',
  jingdou_balance: 0
});

// 编辑用户弹窗
const showEditModal = ref(false);
const editLoading = ref(false);
const editingUser = ref<AdminUser | null>(null);
const editForm = ref<UpdateUserRequest>({});
const showPasswordField = ref(false);

// API Key弹窗
const showApiKeyModal = ref(false);
const apiKeyLoading = ref(false);
const apiKeyUser = ref<AdminUser | null>(null);
const apiKeyInfo = ref<{
  api_key: string | null;
  created_at: string | null;
  last_used_at: string | null;
}>({
  api_key: null,
  created_at: null,
  last_used_at: null
});

// 京豆调整弹窗
const showJingdouModal = ref(false);
const jingdouLoading = ref(false);
const jingdouUser = ref<AdminUser | null>(null);
const jingdouForm = ref({
  amount: 0,
  operationType: 'recharge' as 'recharge' | 'deduct',
  remark: ''
});

// 角色选项
const roleOptions = [
  { label: '普通用户', value: 'common' },
  { label: '管理员', value: 'admin' }
];

// 表格列定义
const columns: DataTableColumn<AdminUser>[] = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '用户名', key: 'username', width: 100 },
  { 
    title: '昵称', 
    key: 'nickname', 
    width: 100,
    render: (row) => row.nickname || '-'
  },
  { 
    title: '角色', 
    key: 'role', 
    width: 80,
    render: (row) => h(NTag, { 
      type: row.role === 'admin' ? 'warning' : 'info',
      size: 'small'
    }, () => row.role === 'admin' ? '管理员' : '普通用户')
  },
  { 
    title: '京豆余额', 
    key: 'jingdou_balance', 
    width: 90,
    render: (row) => h('span', { style: { color: '#2080f0', fontWeight: 600 } }, row.jingdou_balance)
  },
  { 
    title: '未完成任务', 
    key: 'pending_task_count', 
    width: 90,
    render: (row) => h('span', { 
      style: { color: (row.pending_task_count || 0) > 0 ? '#f0a020' : '#18a058' } 
    }, row.pending_task_count || 0)
  },
  { 
    title: '未完成次数', 
    key: 'pending_execute_count', 
    width: 90,
    render: (row) => h('span', { 
      style: { color: (row.pending_execute_count || 0) > 0 ? '#f0a020' : '#18a058' } 
    }, row.pending_execute_count || 0)
  },
  { 
    title: '未完成京豆', 
    key: 'pending_jingdou', 
    width: 90,
    render: (row) => h('span', { 
      style: { color: (row.pending_jingdou || 0) > 0 ? '#f0a020' : '#18a058' } 
    }, row.pending_jingdou || 0)
  },
  { 
    title: '任务未完成%', 
    key: 'pending_task_percent', 
    width: 100,
    render: (row) => {
      const percent = row.pending_task_percent || 0;
      return h('span', { 
        style: { color: percent > 50 ? '#d03050' : percent > 20 ? '#f0a020' : '#18a058' } 
      }, percent.toFixed(1) + '%');
    }
  },
  { 
    title: '次数未完成%', 
    key: 'pending_execute_percent', 
    width: 100,
    render: (row) => {
      const percent = row.pending_execute_percent || 0;
      return h('span', { 
        style: { color: percent > 50 ? '#d03050' : percent > 20 ? '#f0a020' : '#18a058' } 
      }, percent.toFixed(1) + '%');
    }
  },
  { 
    title: '历史消耗', 
    key: 'history_consumed_jingdou', 
    width: 90,
    render: (row) => h('span', { style: { color: '#909399' } }, row.history_consumed_jingdou || 0)
  },
  { 
    title: '状态', 
    key: 'is_active', 
    width: 70,
    render: (row) => h(NTag, { 
      type: row.is_active ? 'success' : 'error',
      size: 'small'
    }, () => row.is_active ? '启用' : '禁用')
  },
  {
    title: '操作',
    key: 'actions',
    width: 310,
    fixed: 'right',
    render: (row) => h(NSpace, { size: 'small' }, () => [
      h(NButton, { size: 'small', type: 'info', onClick: () => handleEdit(row) }, () => '编辑'),
      h(NButton, { size: 'small', type: 'primary', onClick: () => handleViewTasks(row) }, () => '查看任务'),
      h(NButton, { size: 'small', type: 'default', onClick: () => handleApiKey(row) }, () => 'API Key'),
      h(NButton, { size: 'small', type: 'warning', onClick: () => handleJingdou(row) }, () => '调整京豆'),
      row.role !== 'admin' ? h(
        NPopconfirm,
        { 
          onPositiveClick: () => handleDelete(row),
          negativeText: '取消',
          positiveText: '确认删除'
        },
        {
          trigger: () => h(NButton, { size: 'small', type: 'error' }, () => '删除'),
          default: () => `确定要删除用户 ${row.username} 吗？`
        }
      ) : null
    ])
  }
];

// 查看用户任务
function handleViewTasks(user: AdminUser) {
  router.push({
    path: '/task-manage',
    query: {
      user_id: user.id,
      status: 'waiting,running,completed'
    }
  });
}

// 格式化时间
function formatTime(timeStr: string) {
  if (!timeStr) return '-';
  const date = new Date(timeStr);
  return date.toLocaleString('zh-CN', {
    year: 'numeric',
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
}

// 加载用户列表
async function loadUsers() {
  loading.value = true;
  
  const params: Record<string, any> = {
    page: filters.value.page,
    per_page: filters.value.perPage
  };
  
  if (filters.value.search) {
    params.search = filters.value.search;
  }
  
  const { data, error } = await fetchAdminUsers(params);
  if (!error && data) {
    users.value = data.items || [];
    total.value = data.total;
  }
  loading.value = false;
}

// 搜索
function handleSearch() {
  filters.value.page = 1;
  loadUsers();
}

// 重置
function handleReset() {
  filters.value = {
    search: '',
    page: 1,
    perPage: 20
  };
  loadUsers();
}

// 分页变化
function handlePageChange(page: number) {
  filters.value.page = page;
  loadUsers();
}

function handlePageSizeChange(pageSize: number) {
  filters.value.perPage = pageSize;
  filters.value.page = 1;
  loadUsers();
}

// 打开创建用户弹窗
function handleCreate() {
  createForm.value = {
    username: '',
    password: '',
    nickname: '',
    role: 'common',
    jingdou_balance: 0
  };
  showCreateModal.value = true;
}

// 提交创建用户
async function submitCreate() {
  if (!createForm.value.username) {
    message.error('请输入用户名');
    return;
  }
  if (!createForm.value.password || createForm.value.password.length < 6) {
    message.error('密码长度至少6位');
    return;
  }
  
  createLoading.value = true;
  const { error } = await createAdminUser(createForm.value);
  
  if (!error) {
    message.success('用户创建成功');
    showCreateModal.value = false;
    loadUsers();
  }
  createLoading.value = false;
}

// 打开编辑用户弹窗
function handleEdit(user: AdminUser) {
  editingUser.value = user;
  editForm.value = {
    nickname: user.nickname || '',
    role: user.role,
    is_active: user.is_active,
    jingdou_balance: user.jingdou_balance,
    password: undefined
  };
  showPasswordField.value = false;
  showEditModal.value = true;
}

// 提交编辑用户
async function submitEdit() {
  if (!editingUser.value) return;
  
  // 如果密码字段为空，不发送
  const submitData = { ...editForm.value };
  if (!submitData.password) {
    delete submitData.password;
  }
  
  editLoading.value = true;
  const { error } = await updateAdminUser(editingUser.value.id, submitData);
  
  if (!error) {
    message.success('用户信息更新成功');
    showEditModal.value = false;
    loadUsers();
  }
  editLoading.value = false;
}

// 删除用户
async function handleDelete(user: AdminUser) {
  const { error } = await deleteAdminUser(user.id);
  if (!error) {
    message.success('用户删除成功');
    loadUsers();
  }
}

// 打开京豆调整弹窗
function handleJingdou(user: AdminUser) {
  jingdouUser.value = user;
  jingdouForm.value = {
    amount: 0,
    operationType: 'recharge',
    remark: ''
  };
  showJingdouModal.value = true;
}

// 提交京豆调整
async function submitJingdou() {
  if (!jingdouUser.value) return;
  if (!jingdouForm.value.amount || jingdouForm.value.amount <= 0) {
    message.error('请输入有效的金额');
    return;
  }
  
  jingdouLoading.value = true;
  const amount = jingdouForm.value.operationType === 'deduct' 
    ? -jingdouForm.value.amount 
    : jingdouForm.value.amount;
  
  const { error } = await adjustUserJingdou(jingdouUser.value.id, {
    amount,
    operation_type: jingdouForm.value.operationType,
    remark: jingdouForm.value.remark || (jingdouForm.value.operationType === 'recharge' ? '管理员充值' : '管理员扣除')
  });
  
  if (!error) {
    message.success('京豆调整成功');
    showJingdouModal.value = false;
    loadUsers();
  }
  jingdouLoading.value = false;
}

// 打开API Key弹窗
async function handleApiKey(user: AdminUser) {
  apiKeyUser.value = user;
  apiKeyLoading.value = true;
  showApiKeyModal.value = true;
  
  const { data, error } = await getUserApiKey(user.id);
  if (!error && data) {
    apiKeyInfo.value = data;
  } else {
    apiKeyInfo.value = { api_key: null, created_at: null, last_used_at: null };
  }
  apiKeyLoading.value = false;
}

// 重置用户API Key
async function handleResetApiKey() {
  if (!apiKeyUser.value) return;
  
  dialog.warning({
    title: '确认重置',
    content: `确定要为用户 ${apiKeyUser.value.username} 重置API Key吗？原有的API Key将失效。`,
    positiveText: '确认重置',
    negativeText: '取消',
    onPositiveClick: async () => {
      if (!apiKeyUser.value) return;
      apiKeyLoading.value = true;
      const { data, error } = await resetUserApiKey(apiKeyUser.value.id);
      if (!error && data) {
        apiKeyInfo.value = {
          api_key: data.api_key,
          created_at: data.created_at,
          last_used_at: null
        };
        message.success('API Key重置成功');
      }
      apiKeyLoading.value = false;
    }
  });
}

// 删除用户API Key
async function handleDeleteApiKey() {
  if (!apiKeyUser.value) return;
  
  dialog.warning({
    title: '确认删除',
    content: `确定要删除用户 ${apiKeyUser.value.username} 的API Key吗？删除后该用户将无法使用API Key访问接口。`,
    positiveText: '确认删除',
    negativeText: '取消',
    onPositiveClick: async () => {
      if (!apiKeyUser.value) return;
      const { error } = await deleteUserApiKey(apiKeyUser.value.id);
      if (!error) {
        apiKeyInfo.value = { api_key: null, created_at: null, last_used_at: null };
        message.success('API Key已删除');
      }
    }
  });
}

// 复制API Key
function copyApiKey() {
  if (apiKeyInfo.value.api_key) {
    navigator.clipboard.writeText(apiKeyInfo.value.api_key);
    message.success('API Key已复制到剪贴板');
  }
}

onMounted(() => {
  loadUsers();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 筛选区域 -->
    <NCard title="用户管理" :bordered="false">
      <template #header-extra>
        <NSpace>
          <NButton type="primary" @click="handleCreate">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
                <path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/>
              </svg>
            </template>
            新建用户
          </NButton>
        </NSpace>
      </template>
      
      <NSpace wrap :size="16">
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-400">搜索:</span>
          <NInput
            v-model:value="filters.search"
            placeholder="输入用户名搜索"
            clearable
            style="width: 200px"
            @keyup.enter="handleSearch"
          />
        </div>
        <NButton type="primary" @click="handleSearch">查询</NButton>
        <NButton @click="handleReset">重置</NButton>
      </NSpace>
    </NCard>

    <!-- 用户列表 -->
    <NCard :bordered="false" class="flex-1-hidden">
      <NDataTable
        :loading="loading"
        :columns="columns"
        :data="users"
        :row-key="(row: AdminUser) => row.id"
        :scroll-x="1600"
        :pagination="pagination"
        @update:page="handlePageChange"
        @update:page-size="handlePageSizeChange"
      />
    </NCard>

    <!-- 创建用户弹窗 -->
    <NModal
      v-model:show="showCreateModal"
      preset="card"
      title="新建用户"
      style="width: 500px"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="100px">
        <NFormItem label="用户名" required>
          <NInput v-model:value="createForm.username" placeholder="请输入用户名" />
        </NFormItem>
        <NFormItem label="密码" required>
          <NInput v-model:value="createForm.password" type="password" placeholder="请输入密码（至少6位）" />
        </NFormItem>
        <NFormItem label="昵称">
          <NInput v-model:value="createForm.nickname" placeholder="请输入昵称（可选）" />
        </NFormItem>
        <NFormItem label="角色">
          <NSelect v-model:value="createForm.role" :options="roleOptions" style="width: 100%" />
        </NFormItem>
        <NFormItem label="初始京豆">
          <NInputNumber v-model:value="createForm.jingdou_balance" :min="0" style="width: 100%">
            <template #suffix>京豆</template>
          </NInputNumber>
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showCreateModal = false">取消</NButton>
          <NButton type="primary" :loading="createLoading" @click="submitCreate">创建</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 编辑用户弹窗 -->
    <NModal
      v-model:show="showEditModal"
      preset="card"
      title="编辑用户"
      style="width: 500px"
      :mask-closable="false"
    >
      <template v-if="editingUser">
        <NDescriptions :column="1" label-placement="left" bordered class="mb-4">
          <NDescriptionsItem label="用户ID">#{{ editingUser.id }}</NDescriptionsItem>
          <NDescriptionsItem label="用户名">{{ editingUser.username }}</NDescriptionsItem>
        </NDescriptions>

        <NForm label-placement="left" label-width="100px">
          <NFormItem label="昵称">
            <NInput v-model:value="editForm.nickname" placeholder="请输入昵称" />
          </NFormItem>
          <NFormItem label="角色">
            <NSelect v-model:value="editForm.role" :options="roleOptions" style="width: 100%" />
          </NFormItem>
          <NFormItem label="账户状态">
            <NSwitch v-model:value="editForm.is_active">
              <template #checked>启用</template>
              <template #unchecked>禁用</template>
            </NSwitch>
          </NFormItem>
          <NFormItem label="京豆余额">
            <NInputNumber v-model:value="editForm.jingdou_balance" :min="0" style="width: 100%">
              <template #suffix>京豆</template>
            </NInputNumber>
          </NFormItem>
          <NFormItem label="修改密码">
            <div style="width: 100%">
              <NButton v-if="!showPasswordField" size="small" @click="showPasswordField = true">
                点击修改密码
              </NButton>
              <NInput 
                v-else
                v-model:value="editForm.password" 
                type="password" 
                placeholder="输入新密码（至少6位，留空不修改）" 
              />
            </div>
          </NFormItem>
        </NForm>
      </template>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEditModal = false">取消</NButton>
          <NButton type="primary" :loading="editLoading" @click="submitEdit">保存</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 京豆调整弹窗 -->
    <NModal
      v-model:show="showJingdouModal"
      preset="card"
      title="调整京豆"
      style="width: 500px"
      :mask-closable="false"
    >
      <template v-if="jingdouUser">
        <NDescriptions :column="1" label-placement="left" bordered class="mb-4">
          <NDescriptionsItem label="用户">{{ jingdouUser.username }}</NDescriptionsItem>
          <NDescriptionsItem label="当前余额">
            <span style="color: #2080f0; font-weight: 600">{{ jingdouUser.jingdou_balance }} 京豆</span>
          </NDescriptionsItem>
        </NDescriptions>

        <NForm label-placement="left" label-width="100px">
          <NFormItem label="操作类型" required>
            <NSelect
              v-model:value="jingdouForm.operationType"
              :options="[
                { label: '充值（增加）', value: 'recharge' },
                { label: '扣除（减少）', value: 'deduct' }
              ]"
              style="width: 100%"
            />
          </NFormItem>
          <NFormItem label="金额" required>
            <NInputNumber v-model:value="jingdouForm.amount" :min="1" :max="1000000" style="width: 100%">
              <template #suffix>京豆</template>
            </NInputNumber>
          </NFormItem>
          <NFormItem label="备注">
            <NInput v-model:value="jingdouForm.remark" type="textarea" :rows="2" placeholder="请输入备注信息" />
          </NFormItem>
        </NForm>
      </template>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showJingdouModal = false">取消</NButton>
          <NButton type="primary" :loading="jingdouLoading" @click="submitJingdou">
            {{ jingdouForm.operationType === 'recharge' ? '确认充值' : '确认扣除' }}
          </NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- API Key管理弹窗 -->
    <NModal
      v-model:show="showApiKeyModal"
      preset="card"
      title="API Key 管理"
      style="width: 600px"
      :mask-closable="false"
    >
      <template v-if="apiKeyUser">
        <NDescriptions :column="1" label-placement="left" bordered class="mb-4">
          <NDescriptionsItem label="用户">{{ apiKeyUser.username }}</NDescriptionsItem>
        </NDescriptions>

        <div v-if="apiKeyLoading" class="py-8 text-center">
          加载中...
        </div>
        <div v-else>
          <NDescriptions :column="1" label-placement="left" bordered>
            <NDescriptionsItem label="API Key">
              <div v-if="apiKeyInfo.api_key" class="flex items-center gap-2">
                <code class="api-key-code">
                  {{ apiKeyInfo.api_key }}
                </code>
                <NButton size="small" @click="copyApiKey">复制</NButton>
              </div>
              <span v-else class="text-gray-400">未生成</span>
            </NDescriptionsItem>
            <NDescriptionsItem label="创建时间">
              {{ apiKeyInfo.created_at ? formatTime(apiKeyInfo.created_at) : '-' }}
            </NDescriptionsItem>
            <NDescriptionsItem label="最后使用">
              {{ apiKeyInfo.last_used_at ? formatTime(apiKeyInfo.last_used_at) : '从未使用' }}
            </NDescriptionsItem>
          </NDescriptions>
        </div>
      </template>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showApiKeyModal = false">关闭</NButton>
          <NButton type="warning" :loading="apiKeyLoading" @click="handleResetApiKey">
            {{ apiKeyInfo.api_key ? '重置 API Key' : '生成 API Key' }}
          </NButton>
          <NButton 
            v-if="apiKeyInfo.api_key" 
            type="error" 
            :loading="apiKeyLoading" 
            @click="handleDeleteApiKey"
          >
            删除 API Key
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.api-key-code {
  background: rgba(99, 125, 255, 0.15);
  color: #63e2b7;
  padding: 6px 12px;
  border-radius: 6px;
  font-family: 'Consolas', 'Monaco', 'Courier New', monospace;
  font-size: 13px;
  word-break: break-all;
  border: 1px solid rgba(99, 125, 255, 0.3);
}
</style>
