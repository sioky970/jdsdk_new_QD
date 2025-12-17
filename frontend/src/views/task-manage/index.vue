<script setup lang="ts">
import { computed, h, onMounted, onUnmounted, ref, watch } from 'vue';
import { useRoute, useRouter } from 'vue-router';
import type { DataTableSortState } from 'naive-ui';
import {
  NButton,
  NCard,
  NDataTable,
  NDatePicker,
  NDescriptions,
  NDescriptionsItem,
  NForm,
  NFormItem,
  NInput,
  NInputNumber,
  NModal,
  NSelect,
  NSpace,
  NTag,
  NSlider,
  NPopconfirm,
  NSpin,
  useDialog,
  useMessage
} from 'naive-ui';
import {
  fetchUserTasks,
  fetchTaskStatusOptions,
  cancelUserTask,
  updateUserTask,
  createTask,
  batchCreateTasks,
  fetchActiveTaskTypes,
  type CreateTaskRequest,
  type BatchCreateTaskRequest
} from '@/service/api';
import { updateTaskPriority, searchAdminUsers } from '@/service/api/admin';
import type { UserTask, StatusOption, UpdateTaskRequest, AdminUser, TaskTypeItem } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';

const message = useMessage();
const dialog = useDialog();
const authStore = useAuthStore();
const route = useRoute();
const router = useRouter();

// 是否为管理员
const isAdmin = computed(() => authStore.userInfo.role === 'admin');

// 状态数据
const loading = ref(false);
const tasks = ref<UserTask[]>([]);
const total = ref(0);
const statusOptions = ref<StatusOption[]>([]);
const currentTime = ref(new Date());
let timeUpdateInterval: number | null = null;

// 筛选条件
const filters = ref({
  status: 'waiting,running', // 默认显示待执行和执行中的任务
  dateRange: null as [number, number] | null,
  page: 1,
  perPage: 20
});

// 排序状态
const sortState = ref<{ columnKey: string; order: 'ascend' | 'descend' | false }>({
  columnKey: 'start_time',
  order: 'ascend' // 默认按开始时间升序
});

// 编辑弹窗
const showEditModal = ref(false);
const editingTask = ref<UserTask | null>(null);
const editForm = ref<UpdateTaskRequest>({});
const editLoading = ref(false);

// 管理员：优先级调整弹窗
const showPriorityModal = ref(false);
const priorityTask = ref<UserTask | null>(null);
const newPriority = ref(0);
const priorityLoading = ref(false);

// 管理员：用户筛选
const selectedUserId = ref<number | null>(null);
const userSearchKeyword = ref('');
const userOptions = ref<{ label: string; value: number }[]>([]);
const userSearchLoading = ref(false);

// 创建任务弹窗
const showCreateModal = ref(false);
const taskTypes = ref<TaskTypeItem[]>([]);
const taskTypeLoading = ref(false);
const isBatchMode = ref(false); // 批量模式
const batchTaskText = ref(''); // 批量任务文本
const createForm = ref<CreateTaskRequest>({
  task_type: '',
  sku: '',
  shop_name: '',
  keyword: '',
  start_time: '',
  execute_count: 1,
  priority: 0,
  remark: ''
});
const createLoading = ref(false);
const selectedDateTime = ref<number | null>(null);

// 分页信息
const pagination = computed(() => ({
  page: filters.value.page,
  pageSize: filters.value.perPage,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50],
  prefix: () => `共 ${total.value} 条`
}));

// 是否还有更多数据
const hasMore = computed(() => {
  const totalPages = Math.ceil(total.value / filters.value.perPage);
  return filters.value.page < totalPages;
});

// 当前筛选状态的显示文本
const currentStatusText = computed(() => {
  if (!filters.value.status) return '全部状态';
  
  const statuses = filters.value.status.split(',');
  const statusMap: Record<string, string> = {
    'waiting': '待开始',
    'running': '执行中',
    'completed': '已完成',
    'partial_completed': '部分完成',
    'failed': '失败',
    'cancelled': '已取消'
  };
  
  const labels = statuses.map(s => statusMap[s] || s).filter(Boolean);
  return labels.join('、');
});

// 加载任务列表
async function loadTasks(append = false) {
  loading.value = true;
  
  const params: Record<string, any> = {
    page: filters.value.page,
    per_page: filters.value.perPage
  };
  
  if (filters.value.status) {
    params.status = filters.value.status;
  }
  
  if (filters.value.dateRange) {
    params.start_date = new Date(filters.value.dateRange[0]).toISOString().split('T')[0];
    params.end_date = new Date(filters.value.dateRange[1]).toISOString().split('T')[0];
  }

  // 添加排序参数
  if (sortState.value.order) {
    params.sort_by = sortState.value.columnKey;
    params.sort_order = sortState.value.order === 'ascend' ? 'asc' : 'desc';
  }

  // 管理员：支持按用户筛选
  if (isAdmin.value && selectedUserId.value) {
    params.user_id = selectedUserId.value;
  }
  
  const { data, error } = await fetchUserTasks(params);
  if (!error && data) {
    if (append) {
      tasks.value = [...tasks.value, ...(data.tasks || [])];
    } else {
      tasks.value = data.tasks || [];
    }
    total.value = data.total;
  }
  loading.value = false;
}

// 管理员：搜索用户
async function handleUserSearch(keyword: string) {
  if (!keyword || keyword.length < 1) {
    userOptions.value = [];
    return;
  }
  userSearchLoading.value = true;
  const { data, error } = await searchAdminUsers({ keyword });
  if (!error && data) {
    userOptions.value = (data.users || []).map((u: AdminUser) => ({
      label: `${u.username} (${u.nickname || u.username})`,
      value: u.id
    }));
  }
  userSearchLoading.value = false;
}

// 管理员：打开优先级调整弹窗
function handlePriority(task: UserTask) {
  priorityTask.value = task;
  newPriority.value = task.priority || 0;
  showPriorityModal.value = true;
}

// 管理员：提交优先级调整
async function submitPriority() {
  if (!priorityTask.value) return;
  
  priorityLoading.value = true;
  const { error } = await updateTaskPriority(priorityTask.value.id, newPriority.value);
  
  if (!error) {
    message.success('优先级调整成功');
    showPriorityModal.value = false;
    loadTasks();
  }
  priorityLoading.value = false;
}

// 加载任务类型
async function loadTaskTypes() {
  taskTypeLoading.value = true;
  const { data, error } = await fetchActiveTaskTypes();
  if (!error && data) {
    taskTypes.value = data.task_types || [];
  }
  taskTypeLoading.value = false;
}

// 打开创建任务弹窗
function openCreateModal(batch = false) {
  isBatchMode.value = batch;
  batchTaskText.value = '';
  createForm.value = {
    task_type: taskTypes.value.length > 0 ? taskTypes.value[0].type_code : '',
    sku: '',
    shop_name: '',
    keyword: '',
    start_time: '',
    execute_count: 1,
    priority: 0,
    remark: ''
  };
  selectedDateTime.value = null;
  showCreateModal.value = true;
}

// 计算当前任务类型信息
const currentTaskType = computed(() => {
  return taskTypes.value.find(t => t.type_code === createForm.value.task_type);
});

// 是否是关键词搜索任务
const isSearchBrowseTask = computed(() => {
  return createForm.value.task_type === 'search_browse';
});

// 计算消耗
const estimatedConsume = computed(() => {
  if (!currentTaskType.value) return 0;
  return currentTaskType.value.jingdou_price * createForm.value.execute_count;
});

// 单个任务表单验证
const isSingleFormValid = computed(() => {
  if (!createForm.value.task_type || !createForm.value.sku.trim() || !selectedDateTime.value) {
    return false;
  }
  if (isSearchBrowseTask.value) {
    return createForm.value.keyword.trim() !== '' && createForm.value.shop_name.trim() !== '';
  }
  return true;
});

// 解析批量任务文本
function parseBatchTasks(): CreateTaskRequest[] {
  const lines = batchTaskText.value.split('\n').filter(line => line.trim());
  const tasks: CreateTaskRequest[] = [];
  
  for (const line of lines) {
    const parts = line.split('|').map(p => p.trim());
    if (parts.length < 2) continue;
    
    const task: CreateTaskRequest = {
      task_type: createForm.value.task_type,
      sku: parts[0],
      shop_name: parts[1] || '',
      keyword: parts[2] || '',
      start_time: createForm.value.start_time,
      execute_count: createForm.value.execute_count,
      priority: createForm.value.priority || 0,
      remark: createForm.value.remark || ''
    };
    
    // 如果不是关键词搜索任务，清空关键词
    if (task.task_type !== 'search_browse') {
      task.keyword = '';
    }
    
    tasks.push(task);
  }
  
  return tasks;
}

// 批量任务预览
const batchTasksPreview = computed(() => {
  return parseBatchTasks();
});

// 批量任务总消耗
const batchTotalConsume = computed(() => {
  if (!currentTaskType.value) return 0;
  return batchTasksPreview.value.length * currentTaskType.value.jingdou_price * createForm.value.execute_count;
});

// 时间选择器变化
function handleDateTimeChange(value: number | null) {
  if (value) {
    createForm.value.start_time = new Date(value).toISOString();
  }
}

// 提交创建任务
async function submitCreateTask() {
  if (isBatchMode.value) {
    await submitBatchCreate();
  } else {
    await submitSingleCreate();
  }
}

// 提交单个任务创建
async function submitSingleCreate() {
  if (!isSingleFormValid.value) {
    message.error('请填写完整信息');
    return;
  }
  
  createLoading.value = true;
  const { data, error } = await createTask(createForm.value);
  
  if (!error && data) {
    message.success(`任务创建成功，消耗 ${data.consume_jingdou} 京豆`);
    showCreateModal.value = false;
    loadTasks();
  }
  createLoading.value = false;
}

// 提交批量创建
async function submitBatchCreate() {
  const tasks = parseBatchTasks();
  
  if (tasks.length === 0) {
    message.error('请输入批量任务数据');
    return;
  }
  
  if (tasks.length > 100) {
    message.error('单次最多创建100个任务');
    return;
  }
  
  if (!selectedDateTime.value) {
    message.error('请选择开始时间');
    return;
  }
  
  createLoading.value = true;
  const { data, error } = await batchCreateTasks({ tasks });
  
  if (!error && data) {
    if (data.fail_count > 0) {
      message.warning(
        `批量创建完成：成功 ${data.success_count} 个，失败 ${data.fail_count} 个，共消耗 ${data.total_consume} 京豆`
      );
    } else {
      message.success(`批量创建成功，创建 ${data.success_count} 个任务，共消耗 ${data.total_consume} 京豆`);
    }
    showCreateModal.value = false;
    loadTasks();
  }
  createLoading.value = false;
}

// 加载更多数据（无限滚动）
async function loadMore() {
  if (!loading.value && hasMore.value) {
    filters.value.page++;
    await loadTasks(true);
  }
}

// 加载状态选项
async function loadStatusOptions() {
  const { data, error } = await fetchTaskStatusOptions();
  if (!error && data) {
    statusOptions.value = data.options || [];
  }
}

// 搜索
function handleSearch() {
  filters.value.page = 1;
  tasks.value = []; // 清空现有数据
  loadTasks();
}

// 重置筛选
function handleReset() {
  filters.value = {
    status: 'waiting,running',
    dateRange: null,
    page: 1,
    perPage: 20
  };
  sortState.value = {
    columnKey: 'start_time',
    order: 'ascend'
  };
  selectedUserId.value = null;
  userSearchKeyword.value = '';
  tasks.value = []; // 清空现有数据
  loadTasks();
}

// 分页变化（保留但不再使用）
function handlePageChange(page: number) {
  filters.value.page = page;
  loadTasks();
}

function handlePageSizeChange(pageSize: number) {
  filters.value.perPage = pageSize;
  filters.value.page = 1;
  loadTasks();
}

// 排序变化
function handleSorterChange(sorter: DataTableSortState | DataTableSortState[] | null) {
  if (!sorter || Array.isArray(sorter)) {
    sortState.value = { columnKey: 'created_at', order: 'descend' };
  } else {
    sortState.value = {
      columnKey: sorter.columnKey as string,
      order: sorter.order as 'ascend' | 'descend' | false
    };
  }
  filters.value.page = 1;
  tasks.value = []; // 清空现有数据
  loadTasks();
}

// 处理滚动事件
function handleScroll(e: Event) {
  const target = e.target as HTMLElement;
  const scrollTop = target.scrollTop;
  const scrollHeight = target.scrollHeight;
  const clientHeight = target.clientHeight;
  
  // 当滚动到距离底部50px时触发加载更多
  if (scrollHeight - scrollTop - clientHeight < 50) {
    loadMore();
  }
}

// 取消任务
function handleCancel(task: UserTask) {
  dialog.warning({
    title: '确认取消',
    content: `确定要取消任务 #${task.id} 吗？取消后将退还 ${task.consume_jingdou} 京豆。`,
    positiveText: '确认取消',
    negativeText: '再想想',
    onPositiveClick: async () => {
      const { data, error } = await cancelUserTask(task.id);
      if (!error && data) {
        message.success(`任务已取消，退还 ${data.refund_jingdou} 京豆`);
        loadTasks();
      }
    }
  });
}

// 编辑任务
function handleEdit(task: UserTask) {
  editingTask.value = task;
  editForm.value = {
    shop_name: task.shop_name,
    keyword: task.keyword,
    start_time: task.start_time,
    execute_count: task.execute_count
  };
  showEditModal.value = true;
}

// 提交编辑
async function submitEdit() {
  if (!editingTask.value) return;
  
  // 验证次数不能减少
  if (editForm.value.execute_count && editForm.value.execute_count < editingTask.value.execute_count) {
    message.error('执行次数不能减少');
    return;
  }
  
  editLoading.value = true;
  const { data, error } = await updateUserTask(editingTask.value.id, editForm.value);
  
  if (!error && data) {
    if (data.additional_jingdou > 0) {
      message.success(`任务修改成功，额外消耗 ${data.additional_jingdou} 京豆`);
    } else {
      message.success('任务修改成功');
    }
    showEditModal.value = false;
    loadTasks();
  }
  editLoading.value = false;
}

// 状态标签类型
function getStatusTagType(status: string) {
  const map: Record<string, 'success' | 'info' | 'warning' | 'error' | 'default'> = {
    completed: 'success',
    waiting: 'info',
    running: 'warning',
    partial_completed: 'warning',
    failed: 'error',
    cancelled: 'default'
  };
  return map[status] || 'default';
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

// 计算倒计时或已开始状态（使用当前时间）
function getTimeCountdown(startTimeStr: string, taskStatus: string) {
  if (!startTimeStr) return null;
  
  // 已取消的任务不显示倒计时
  if (taskStatus === 'cancelled') return null;
  
  const startTime = new Date(startTimeStr);
  const diff = startTime.getTime() - currentTime.value.getTime();
  
  // 如果已经开始
  if (diff <= 0) {
    return { type: 'started', text: '已到达开始时间' };
  }
  
  // 计算倒计时（精确到秒）
  const days = Math.floor(diff / (1000 * 60 * 60 * 24));
  const hours = Math.floor((diff % (1000 * 60 * 60 * 24)) / (1000 * 60 * 60));
  const minutes = Math.floor((diff % (1000 * 60 * 60)) / (1000 * 60));
  const seconds = Math.floor((diff % (1000 * 60)) / 1000);
  
  let countdown = '';
  if (days > 0) {
    countdown = `${days}天${hours}小时${minutes}分`;
  } else if (hours > 0) {
    countdown = `${hours}小时${minutes}分${seconds}秒`;
  } else if (minutes > 0) {
    countdown = `${minutes}分${seconds}秒`;
  } else if (seconds > 0) {
    countdown = `${seconds}秒`;
  } else {
    countdown = '即将开始';
  }
  
  return { type: 'countdown', text: countdown };
}

onMounted(async () => {
  loadStatusOptions();
  loadTaskTypes(); // 加载任务类型列表
  
  // 检查 URL 参数中是否有 user_id
  const userIdParam = route.query.user_id;
  if (isAdmin.value && userIdParam) {
    const userId = Number(userIdParam);
    if (!isNaN(userId)) {
      selectedUserId.value = userId;
      // 搜索用户信息以显示用户名
      const { data } = await searchAdminUsers({ keyword: '' });
      if (data?.users) {
        const user = data.users.find((u: AdminUser) => u.id === userId);
        if (user) {
          userOptions.value = [{
            label: `${user.username} (${user.nickname || user.username})`,
            value: user.id
          }];
        }
      }
    }
  }
  
  // 检查 URL 参数中是否有 status
  const statusParam = route.query.status;
  if (statusParam && typeof statusParam === 'string') {
    filters.value.status = statusParam;
  }
  
  loadTasks();
  
  // 启动定时器，每秒更新时间
  timeUpdateInterval = window.setInterval(() => {
    currentTime.value = new Date();
  }, 1000);
});

onUnmounted(() => {
  // 清除定时器
  if (timeUpdateInterval !== null) {
    clearInterval(timeUpdateInterval);
  }
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 筛选区域 -->
    <NCard title="任务管理" :bordered="false">
      <template #header-extra>
        <NSpace>
          <NButton type="primary" @click="() => openCreateModal(false)">
            <template #icon>
              <span class="i-carbon-add"></span>
            </template>
            创建任务
          </NButton>
          <NButton type="info" @click="() => openCreateModal(true)">
            <template #icon>
              <span class="i-carbon-document-multiple-01"></span>
            </template>
            批量创建
          </NButton>
          <NButton type="primary" @click="handleSearch">查询</NButton>
          <NButton @click="handleReset">重置</NButton>
        </NSpace>
      </template>
      
      <NSpace wrap :size="16">
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-400">状态:</span>
          <NSelect
            v-model:value="filters.status"
            :options="statusOptions.map(o => ({ label: o.label, value: o.value }))"
            placeholder="全部状态"
            clearable
            style="width: 180px"
          >
            <template #label>
              <span>{{ currentStatusText }}</span>
            </template>
          </NSelect>
        </div>
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-400">时间:</span>
          <NDatePicker
            v-model:value="filters.dateRange"
            type="daterange"
            clearable
            style="width: 260px"
          />
        </div>
        <!-- 管理员：用户筛选 -->
        <div v-if="isAdmin" class="flex items-center gap-2">
          <span class="text-sm text-gray-400">用户:</span>
          <NSelect
            v-model:value="selectedUserId"
            filterable
            remote
            :loading="userSearchLoading"
            :options="userOptions"
            placeholder="搜索用户"
            clearable
            style="width: 180px"
            @search="handleUserSearch"
          />
        </div>
      </NSpace>
    </NCard>

    <!-- 任务列表 -->
    <NCard :bordered="false" class="flex-1-hidden">
      <div class="table-container" @scroll="handleScroll">
        <NDataTable
          :loading="loading"
          :columns="[
            { title: 'ID', key: 'id', width: 60 },
            { title: '任务类型', key: 'task_type_name', width: 120 },
            { title: 'SKU', key: 'sku', width: 140, ellipsis: { tooltip: true } },
            { title: '店铺', key: 'shop_name', width: 120, ellipsis: { tooltip: true } },
                      { title: '关键词', key: 'keyword', width: 120, ellipsis: { tooltip: true }, render: (row: UserTask) => row.keyword || '-' },
            { 
              title: '状态', 
              key: 'status', 
              width: 100,
              render: (row: UserTask) => h(NTag, { type: getStatusTagType(row.status), size: 'small' }, () => row.status_text)
            },
            { 
              title: '进度', 
              key: 'progress', 
              width: 100,
              render: (row: UserTask) => `${row.executed_count} / ${row.execute_count}`
            },
            { title: '消耗京豆', key: 'consume_jingdou', width: 90 },
            // 管理员：显示优先级列
            ...(isAdmin ? [{
              title: '优先级',
              key: 'priority',
              width: 80,
              render: (row: UserTask) => h(NTag, { type: row.priority > 0 ? 'warning' : 'default', size: 'small' }, () => row.priority || 0)
            }] : []),
            { 
              title: '开始时间', 
              key: 'start_time', 
              width: 180,
              align: 'center',
              sorter: true,
              sortOrder: sortState.columnKey === 'start_time' ? sortState.order : false,
              render: (row: UserTask) => {
                const countdown = getTimeCountdown(row.start_time, row.status);
                return h('div', { class: 'flex flex-col gap-1 items-center' }, [
                  h('div', { class: 'text-sm' }, formatTime(row.start_time)),
                  countdown ? h(
                    NTag,
                    { 
                      type: countdown.type === 'started' ? 'success' : 'info',
                      size: 'small'
                    },
                    () => countdown.text
                  ) : null
                ]);
              }
            },
            { 
              title: '创建时间', 
              key: 'created_at', 
              width: 160,
              sorter: true,
              sortOrder: sortState.columnKey === 'created_at' ? sortState.order : false,
              render: (row: UserTask) => formatTime(row.created_at)
            },
            {
              title: '操作',
              key: 'actions',
              width: isAdmin ? 180 : 140,
              fixed: 'right',
              render: (row: UserTask) => h(NSpace, { size: 'small' }, () => [
                row.can_edit ? h(NButton, { size: 'small', type: 'info', onClick: () => handleEdit(row) }, () => '编辑') : null,
                row.can_cancel ? h(NButton, { size: 'small', type: 'error', onClick: () => handleCancel(row) }, () => '取消') : null,
                // 管理员：优先级调整按钮
                isAdmin ? h(NButton, { size: 'small', type: 'warning', onClick: () => handlePriority(row) }, () => '优先级') : null,
                !row.can_edit && !row.can_cancel && !isAdmin ? h('span', { class: 'text-gray-400 text-xs' }, '-') : null
              ])
            }
          ]"
          :data="tasks"
          :row-key="(row: UserTask) => row.id"
          :scroll-x="1340"
          @update:sorter="handleSorterChange"
        />
        
        <!-- 加载更多提示 -->
        <div v-if="hasMore" class="load-more-trigger">
          <div v-if="loading" class="text-center py-4 text-gray-400">
            加载中...
          </div>
        </div>
        
        <!-- 底部空白区域 -->
        <div v-if="!hasMore && tasks.length > 0" class="bottom-spacer">
          <div class="text-center text-gray-400 text-sm">
            已加载全部 {{ total }} 条数据
          </div>
        </div>
      </div>
    </NCard>

    <!-- 编辑弹窗 -->
    <NModal
      v-model:show="showEditModal"
      preset="card"
      title="编辑任务"
      style="width: 500px"
      :mask-closable="false"
    >
      <template v-if="editingTask">
        <NDescriptions :column="1" label-placement="left" bordered class="mb-4">
          <NDescriptionsItem label="任务ID">#{{ editingTask.id }}</NDescriptionsItem>
          <NDescriptionsItem label="任务类型">{{ editingTask.task_type_name }}</NDescriptionsItem>
          <NDescriptionsItem label="SKU">{{ editingTask.sku }}</NDescriptionsItem>
        </NDescriptions>

        <NForm label-placement="left" label-width="80px">
          <NFormItem label="店铺名称">
            <NInput v-model:value="editForm.shop_name" placeholder="店铺名称" />
          </NFormItem>
          <NFormItem v-if="editingTask.task_type === 'search_browse'" label="关键词">
            <NInput v-model:value="editForm.keyword" placeholder="搜索关键词" />
          </NFormItem>
          <NFormItem label="执行次数">
            <NInputNumber
              v-model:value="editForm.execute_count"
              :min="editingTask.execute_count"
              :max="10000"
              style="width: 100%"
            />
            <template #feedback>
              <span class="text-warning text-xs">只能增加次数，不能减少（增加次数将额外扣除京豆）</span>
            </template>
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

    <!-- 管理员：优先级调整弹窗 -->
    <NModal
      v-model:show="showPriorityModal"
      preset="card"
      title="调整任务优先级"
      style="width: 450px"
      :mask-closable="false"
    >
      <template v-if="priorityTask">
        <NDescriptions :column="1" label-placement="left" bordered class="mb-4">
          <NDescriptionsItem label="任务ID">#{{ priorityTask.id }}</NDescriptionsItem>
          <NDescriptionsItem label="任务类型">{{ priorityTask.task_type_name }}</NDescriptionsItem>
          <NDescriptionsItem label="SKU">{{ priorityTask.sku }}</NDescriptionsItem>
          <NDescriptionsItem label="当前优先级">
            <NTag :type="priorityTask.priority > 0 ? 'warning' : 'default'" size="small">
              {{ priorityTask.priority || 0 }}
            </NTag>
          </NDescriptionsItem>
        </NDescriptions>

        <NForm label-placement="left" label-width="100px">
          <NFormItem label="新优先级">
            <div class="w-full">
              <NSlider
                v-model:value="newPriority"
                :min="0"
                :max="100"
                :step="1"
                :marks="{ 0: '普通', 50: '中', 100: '最高' }"
              />
              <div class="text-center mt-2">
                <NInputNumber v-model:value="newPriority" :min="0" :max="100" style="width: 120px" />
              </div>
            </div>
          </NFormItem>
          <NFormItem>
            <div class="text-gray-400 text-xs">
              优先级越高，任务将更早被执行。范围：0-100
            </div>
          </NFormItem>
        </NForm>
      </template>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showPriorityModal = false">取消</NButton>
          <NButton type="primary" :loading="priorityLoading" @click="submitPriority">确认调整</NButton>
        </NSpace>
      </template>
    </NModal>

    <!-- 创建任务弹窗 -->
    <NModal
      v-model:show="showCreateModal"
      preset="card"
      :title="isBatchMode ? '批量创建任务' : '创建任务'"
      style="width: 600px"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="90px">
        <NFormItem label="任务类型" required>
          <NSelect
            v-model:value="createForm.task_type"
            :options="taskTypes.map(t => ({ label: `${t.type_name} (${t.jingdou_price}京豆/次)`, value: t.type_code }))"
            :loading="taskTypeLoading"
            placeholder="请选择任务类型"
          />
        </NFormItem>

        <!-- 单个创建模式 -->
        <template v-if="!isBatchMode">
          <NFormItem label="SKU编号" required>
            <NInput
              v-model:value="createForm.sku"
              placeholder="请输入SKU编号"
              clearable
            />
          </NFormItem>
          <NFormItem label="店铺名称" :required="isSearchBrowseTask">
            <NInput
              v-model:value="createForm.shop_name"
              :placeholder="isSearchBrowseTask ? '关键词搜索任务必填' : '选填'"
              clearable
            />
          </NFormItem>
          <NFormItem v-if="isSearchBrowseTask" label="搜索关键词" required>
            <NInput
              v-model:value="createForm.keyword"
              placeholder="请输入搜索关键词"
              clearable
            />
          </NFormItem>
          <NFormItem label="备注">
            <NInput
              v-model:value="createForm.remark"
              type="textarea"
              placeholder="选填"
              :rows="2"
            />
          </NFormItem>
        </template>

        <!-- 批量创建模式 -->
        <template v-else>
          <NFormItem label="批量数据" required>
            <NInput
              v-model:value="batchTaskText"
              type="textarea"
              :placeholder="isSearchBrowseTask ? 'SKU|店铺名称|关键词\n每行一个任务，用 | 分隔' : 'SKU|店铺名称\n每行一个任务，用 | 分隔'"
              :rows="10"
            />
          </NFormItem>
          <NFormItem>
            <div class="text-xs text-gray-400">
              <div>格式说明：</div>
              <div v-if="isSearchBrowseTask">关键词搜索：SKU|店铺名称|关键词</div>
              <div v-else>其他任务：SKU|店铺名称</div>
              <div class="mt-1">已解析 <span class="text-primary font-bold">{{ batchTasksPreview.length }}</span> 个任务</div>
            </div>
          </NFormItem>
        </template>

        <!-- 公共参数 -->
        <NFormItem label="执行次数" required>
          <NInputNumber
            v-model:value="createForm.execute_count"
            :min="1"
            :max="10000"
            style="width: 100%"
          />
        </NFormItem>
        
        <NFormItem label="开始时间" required>
          <NDatePicker
            v-model:value="selectedDateTime"
            type="datetime"
            clearable
            :is-date-disabled="(ts: number) => ts < Date.now() - 86400000"
            placeholder="选择任务开始执行时间"
            style="width: 100%"
            @update:value="handleDateTimeChange"
          />
        </NFormItem>

        <NFormItem v-if="isAdmin" label="优先级">
          <NInputNumber
            v-model:value="createForm.priority"
            :min="0"
            :max="100"
            style="width: 100%"
            placeholder="0-100，默认0"
          />
        </NFormItem>

        <!-- 消耗预览 -->
        <NFormItem label="预计消耗">
          <div class="w-full">
            <div class="flex items-center gap-2">
              <span class="text-lg font-bold text-primary">
                {{ isBatchMode ? batchTotalConsume : estimatedConsume }}
              </span>
              <span class="text-gray-400">京豆</span>
            </div>
            <div v-if="isBatchMode" class="text-xs text-gray-400 mt-1">
              {{ batchTasksPreview.length }} 个任务 × {{ createForm.execute_count }} 次 × {{ currentTaskType?.jingdou_price || 0 }} 京豆/次
            </div>
            <div v-else class="text-xs text-gray-400 mt-1">
              {{ createForm.execute_count }} 次 × {{ currentTaskType?.jingdou_price || 0 }} 京豆/次
            </div>
          </div>
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showCreateModal = false">取消</NButton>
          <NButton
            type="primary"
            :loading="createLoading"
            :disabled="isBatchMode ? batchTasksPreview.length === 0 : !isSingleFormValid"
            @click="submitCreateTask"
          >
            {{ isBatchMode ? `创建 ${batchTasksPreview.length} 个任务` : '创建任务' }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.text-warning {
  color: #f0a020;
}

.text-primary {
  color: #18a058;
}

.table-container {
  position: relative;
  max-height: calc(100vh - 300px);
  overflow-y: auto;
}

.table-container :deep(.n-data-table-wrapper) {
  max-height: none !important;
}

.table-container :deep(.n-data-table-base-table-body-wrapper) {
  overflow-y: visible !important;
}

.load-more-trigger {
  min-height: 50px;
  display: flex;
  align-items: center;
  justify-content: center;
}

.bottom-spacer {
  padding: 40px 0 60px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
