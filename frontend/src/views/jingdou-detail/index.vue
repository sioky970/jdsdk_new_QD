<script setup lang="ts">
import { computed, h, onMounted, ref } from 'vue';
import { NButton, NCard, NDataTable, NDatePicker, NSelect, NSpace, NStatistic, NModal, NForm, NFormItem, NInputNumber, NInput, NRadioGroup, NRadio, useMessage } from 'naive-ui';
import type { DataTableColumn } from 'naive-ui';
import * as XLSX from 'xlsx';
import { fetchJingdouRecords, searchAdminUsers, adjustUserJingdou } from '@/service/api';
import type { AdminUser } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';

const message = useMessage();
const authStore = useAuthStore();

// 是否为管理员
const isAdmin = computed(() => authStore.userInfo.role === 'admin');

// 状态数据
const loading = ref(false);
const records = ref<any[]>([]);
const total = ref(0);

// 筛选条件
const filters = ref({
  type: null as string | null, // 类型筛选
  dateRange: null as [number, number] | null, // 日期范围
  page: 1,
  perPage: 20
});

// 类型选项
const typeOptions = [
  { label: '全部类型', value: null },
  { label: '任务消耗', value: 'task_consume' },
  { label: '任务退还', value: 'task_refund' },
  { label: '充值', value: 'recharge' },
  { label: '提现', value: 'withdraw' },
  { label: '管理员调整', value: 'admin_adjust' }
];

// 管理员：用户筛选
const selectedUserId = ref<number | null>(null);
const userSearchKeyword = ref('');
const userOptions = ref<{ label: string; value: number }[]>([]);
const userSearchLoading = ref(false);

// 管理员：调整京豆弹窗
const showAdjustModal = ref(false);
const adjustLoading = ref(false);
const adjustForm = ref({
  userId: null as number | null,
  amount: 0,
  operationType: 'recharge' as 'recharge' | 'deduct',
  remark: ''
});
const adjustUserOptions = ref<{ label: string; value: number }[]>([]);

// 动态计算列（管理员显示归属用户列）
const columns = computed(() => {
  const baseColumns: DataTableColumn[] = [
    {
      title: '时间',
      key: 'created_at',
      width: 180,
      render(row: any) {
        return new Date(row.created_at).toLocaleString('zh-CN');
      }
    },
    {
      title: '类型',
      key: 'type',
      width: 120,
      render(row: any) {
        const typeMap: Record<string, { label: string; color: string }> = {
          task_consume: { label: '任务消耗', color: '#d03050' },
          task_refund: { label: '任务退还', color: '#18a058' },
          recharge: { label: '充值', color: '#2080f0' },
          withdraw: { label: '提现', color: '#f0a020' },
          admin_adjust: { label: '管理员调整', color: '#a78bfa' }
        };
        const info = typeMap[row.type] || { label: row.type, color: '#909399' };
        return h('span', { style: { color: info.color, fontWeight: 600 } }, info.label);
      }
    },
    {
      title: '变动金额',
      key: 'amount',
      width: 120,
      render(row: any) {
        const isIncrease = row.amount > 0;
        const color = isIncrease ? '#18a058' : '#d03050';
        const sign = isIncrease ? '+' : '';
        return h('span', { style: { color, fontWeight: 700, fontSize: '15px' } }, `${sign}${row.amount}`);
      }
    },
    {
      title: '变动后余额',
      key: 'balance_after',
      width: 120,
      render(row: any) {
        return h('span', { style: { fontWeight: 600, color: '#2080f0' } }, String(row.balance_after));
      }
    },
    {
      title: '关联任务ID',
      key: 'task_id',
      width: 120,
      render(row: any) {
        return row.task_id ? `#${row.task_id}` : '-';
      }
    },
    {
      title: '备注',
      key: 'remark',
      minWidth: 200,
      ellipsis: {
        tooltip: true
      }
    }
  ];

  // 管理员专属：在类型列后面插入“归属用户”列
  if (isAdmin.value) {
    const userColumn: DataTableColumn = {
      title: '归属用户',
      key: 'username',
      width: 140,
      render(row: any) {
        if (!row.username) return '-';
        const displayName = row.nickname && row.nickname !== row.username 
          ? `${row.nickname} (${row.username})` 
          : row.username;
        return h('span', { style: { color: '#8b5cf6', fontWeight: 500 } }, displayName);
      }
    };
    // 插入到类型列后面（索引2）
    baseColumns.splice(2, 0, userColumn);
  }

  return baseColumns;
});

// 分页信息
const pagination = computed(() => ({
  page: filters.value.page,
  pageSize: filters.value.perPage,
  itemCount: total.value,
  showSizePicker: true,
  pageSizes: [10, 20, 50, 100],
  prefix: () => `共 ${total.value} 条`
}));

// 是否还有更多数据
const hasMore = computed(() => {
  const totalPages = Math.ceil(total.value / filters.value.perPage);
  return filters.value.page < totalPages;
});

// 计算总计数据
const summary = computed(() => {
  // 计算当前列表中的变动金额合计
  const totalAmount = records.value.reduce((sum, record) => sum + (record.amount || 0), 0);
  
  // 获取类型名称
  const getTypeName = () => {
    if (!filters.value.type) return '全部类型';
    const typeMap: Record<string, string> = {
      task_consume: '任务消耗',
      task_refund: '任务退还',
      recharge: '充值',
      withdraw: '提现',
      admin_adjust: '管理员调整'
    };
    return typeMap[filters.value.type] || filters.value.type;
  };
  
  // 获取日期范围文本
  const getDateRangeText = () => {
    if (!filters.value.dateRange) return '全部时间';
    const startDate = new Date(filters.value.dateRange[0]).toLocaleDateString('zh-CN');
    const endDate = new Date(filters.value.dateRange[1]).toLocaleDateString('zh-CN');
    return `${startDate} ~ ${endDate}`;
  };
  
  return {
    totalAmount,
    typeName: getTypeName(),
    dateRangeText: getDateRangeText(),
    recordCount: records.value.length
  };
});

// 加载京豆明细
async function loadRecords(append = false) {
  loading.value = true;
  
  const params: Record<string, any> = {
    page: filters.value.page,
    per_page: filters.value.perPage
  };
  
  if (filters.value.type) {
    params.type = filters.value.type;
  }
  
  if (filters.value.dateRange) {
    params.start_date = new Date(filters.value.dateRange[0]).toISOString().split('T')[0];
    params.end_date = new Date(filters.value.dateRange[1]).toISOString().split('T')[0];
  }

  // 管理员：支持按用户筛选
  if (isAdmin.value && selectedUserId.value) {
    params.user_id = selectedUserId.value;
  }
  
  try {
    const { data, error } = await fetchJingdouRecords(params);
    if (!error && data) {
      if (append) {
        records.value = [...records.value, ...(data.records || [])];
      } else {
        records.value = data.records || [];
      }
      total.value = data.total;
    } else {
      message.error(error?.message || '加载失败');
    }
  } catch (error) {
    message.error('加载失败');
  } finally {
    loading.value = false;
  }
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
      label: `${u.username} (${u.nickname || u.username}) - 余额:${u.jingdou_balance}`,
      value: u.id
    }));
  }
  userSearchLoading.value = false;
}

// 管理员：调整用户搜索
async function handleAdjustUserSearch(keyword: string) {
  if (!keyword || keyword.length < 1) {
    adjustUserOptions.value = [];
    return;
  }
  const { data, error } = await searchAdminUsers({ keyword });
  if (!error && data) {
    adjustUserOptions.value = (data.users || []).map((u: AdminUser) => ({
      label: `${u.username} (${u.nickname || u.username}) - 余额:${u.jingdou_balance}`,
      value: u.id
    }));
  }
}

// 管理员：打开调整京豆弹窗
function handleOpenAdjust() {
  adjustForm.value = {
    userId: selectedUserId.value,
    amount: 0,
    operationType: 'recharge',
    remark: ''
  };
  adjustUserOptions.value = [...userOptions.value];
  showAdjustModal.value = true;
}

// 管理员：提交京豆调整
async function submitAdjust() {
  if (!adjustForm.value.userId) {
    message.error('请选择用户');
    return;
  }
  if (!adjustForm.value.amount || adjustForm.value.amount <= 0) {
    message.error('请输入有效的金额');
    return;
  }
  
  adjustLoading.value = true;
  const amount = adjustForm.value.operationType === 'deduct' ? -adjustForm.value.amount : adjustForm.value.amount;
  
  const { error } = await adjustUserJingdou(adjustForm.value.userId, {
    amount,
    operation_type: adjustForm.value.operationType,
    remark: adjustForm.value.remark || (adjustForm.value.operationType === 'recharge' ? '管理员充值' : '管理员扣除')
  });
  
  if (!error) {
    message.success('京豆调整成功');
    showAdjustModal.value = false;
    loadRecords();
  }
  adjustLoading.value = false;
}

// 加载更多数据（无限滚动）
async function loadMore() {
  if (!loading.value && hasMore.value) {
    filters.value.page++;
    await loadRecords(true);
  }
}

// 容器滚动事件处理
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

onMounted(() => {
  loadRecords();
});

// 搜索
function handleSearch() {
  filters.value.page = 1;
  records.value = []; // 清空现有数据
  loadRecords();
}

// 重置筛选
function handleReset() {
  filters.value = {
    type: null,
    dateRange: null,
    page: 1,
    perPage: 20
  };
  selectedUserId.value = null;
  userSearchKeyword.value = '';
  records.value = []; // 清空现有数据
  loadRecords();
}

// 导出Excel
function handleExport() {
  if (records.value.length === 0) {
    message.warning('暂无数据可导出');
    return;
  }
  
  // 准备导出数据
  const exportData = records.value.map(record => {
    const typeMap: Record<string, string> = {
      task_consume: '任务消耗',
      task_refund: '任务退还',
      recharge: '充值',
      withdraw: '提现',
      admin_adjust: '管理员调整'
    };
    
    return {
      '时间': new Date(record.created_at).toLocaleString('zh-CN'),
      '类型': typeMap[record.type] || record.type,
      '变动金额': record.amount,
      '变动后余额': record.balance_after,
      '关联任务ID': record.task_id || '-',
      '备注': record.remark || ''
    };
  });
  
  // 创建工作簿
  const workbook = XLSX.utils.book_new();
  const worksheet = XLSX.utils.json_to_sheet([]);
  
  // 添加标题行（顶部表头）
  const headerData = [
    ['京豆明细报表'],
    ['查询类型', summary.value.typeName],
    ['日期范围', summary.value.dateRangeText],
    ['记录数量', `${summary.value.recordCount} 条`],
    ['变动金额合计', summary.value.totalAmount],
    [], // 空行
  ];
  
  // 将标题数据写入worksheet
  XLSX.utils.sheet_add_aoa(worksheet, headerData, { origin: 'A1' });
  
  // 添加数据表格（从第7行开始）
  XLSX.utils.sheet_add_json(worksheet, exportData, { origin: 'A7', skipHeader: false });
  
  // 设置列宽
  worksheet['!cols'] = [
    { wch: 20 }, // 时间
    { wch: 12 }, // 类型
    { wch: 12 }, // 变动金额
    { wch: 12 }, // 变动后余额
    { wch: 15 }, // 关联任务ID
    { wch: 30 }  // 备注
  ];
  
  // 合并标题单元格
  worksheet['!merges'] = [
    { s: { r: 0, c: 0 }, e: { r: 0, c: 5 } }  // 合并第一行的6列作为标题
  ];
  
  // 设置标题样式（加粗、居中）
  if (!worksheet['A1'].s) worksheet['A1'].s = {};
  worksheet['A1'].s = {
    font: { bold: true, sz: 16 },
    alignment: { horizontal: 'center', vertical: 'center' }
  };
  
  XLSX.utils.book_append_sheet(workbook, worksheet, '京豆明细');
  
  // 导出文件
  const fileName = `京豆明细_${new Date().toLocaleDateString('zh-CN').replace(/\//g, '-')}.xlsx`;
  XLSX.writeFile(workbook, fileName);
  
  message.success(`已导出 ${records.value.length} 条记录`);
}

</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 筛选区域 -->
    <NCard title="京豆明细" :bordered="false">
      <template #header-extra>
        <NSpace>
          <NButton @click="handleSearch">查询</NButton>
          <NButton @click="handleReset">重置</NButton>
          <!-- 管理员：调整京豆按钮 -->
          <NButton v-if="isAdmin" type="warning" @click="handleOpenAdjust">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
                <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
              </svg>
            </template>
            调整京豆
          </NButton>
          <NButton type="primary" @click="handleExport">
            <template #icon>
              <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor" width="18" height="18">
                <path d="M19 9h-4V3H9v6H5l7 7 7-7zM5 18v2h14v-2H5z"/>
              </svg>
            </template>
            导出Excel
          </NButton>
        </NSpace>
      </template>
      
      <NSpace wrap :size="16">
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-400">类型:</span>
          <NSelect
            v-model:value="filters.type"
            :options="typeOptions"
            placeholder="全部类型"
            clearable
            style="width: 150px"
          />
        </div>
        
        <div class="flex items-center gap-2">
          <span class="text-sm text-gray-400">日期范围:</span>
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
            style="width: 220px"
            @search="handleUserSearch"
          />
        </div>
      </NSpace>
    </NCard>

    <!-- 总计信息 -->
    <NCard :bordered="false" v-if="records.length > 0">
      <div class="summary-container">
        <div class="summary-item">
          <span class="summary-label">查询类型：</span>
          <span class="summary-value-highlight">{{ summary.typeName }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">日期范围：</span>
          <span class="summary-value-highlight">{{ summary.dateRangeText }}</span>
        </div>
        <div class="summary-item">
          <span class="summary-label">记录数量：</span>
          <span class="summary-value-highlight">{{ summary.recordCount }} 条</span>
        </div>
        <div class="summary-item summary-highlight">
          <span class="summary-label">变动金额合计：</span>
          <NStatistic 
            :value="summary.totalAmount" 
            :value-style="{ 
              color: summary.totalAmount >= 0 ? '#18a058' : '#d03050',
              fontWeight: 700,
              fontSize: '20px'
            }"
          >
            <template #prefix>
              <span v-if="summary.totalAmount > 0">+</span>
            </template>
          </NStatistic>
        </div>
      </div>
    </NCard>

    <!-- 数据表格 -->
    <NCard :bordered="false" class="flex-1-hidden">
      <div class="table-container" @scroll="handleScroll">
        <NDataTable
          :columns="columns"
          :data="records"
          :loading="loading"
          :scroll-x="900"
        />
        
        <!-- 加载更多提示 -->
        <div v-if="hasMore" class="load-more-trigger">
          <div v-if="loading" class="text-center py-4 text-gray-400">
            加载中...
          </div>
        </div>
        
        <!-- 底部空白区域 -->
        <div v-if="!hasMore && records.length > 0" class="bottom-spacer">
          <div class="text-center text-gray-400 text-sm">
            已加载全部 {{ total }} 条数据
          </div>
        </div>
      </div>
    </NCard>

    <!-- 管理员：调整京豆弹窗 -->
    <NModal
      v-model:show="showAdjustModal"
      preset="card"
      title="调整用户京豆"
      style="width: 500px"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="100px">
        <NFormItem label="选择用户" required>
          <NSelect
            v-model:value="adjustForm.userId"
            filterable
            remote
            :options="adjustUserOptions"
            placeholder="搜索用户"
            style="width: 100%"
            @search="handleAdjustUserSearch"
          />
        </NFormItem>
        <NFormItem label="操作类型" required>
          <NRadioGroup v-model:value="adjustForm.operationType">
            <NRadio value="recharge">充值（增加）</NRadio>
            <NRadio value="deduct">扣除（减少）</NRadio>
          </NRadioGroup>
        </NFormItem>
        <NFormItem label="金额" required>
          <NInputNumber
            v-model:value="adjustForm.amount"
            :min="1"
            :max="1000000"
            placeholder="请输入京豆数量"
            style="width: 100%"
          >
            <template #suffix>京豆</template>
          </NInputNumber>
        </NFormItem>
        <NFormItem label="备注">
          <NInput
            v-model:value="adjustForm.remark"
            type="textarea"
            :rows="2"
            placeholder="请输入备注信息"
          />
        </NFormItem>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showAdjustModal = false">取消</NButton>
          <NButton type="primary" :loading="adjustLoading" @click="submitAdjust">
            {{ adjustForm.operationType === 'recharge' ? '确认充值' : '确认扣除' }}
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.summary-container {
  display: grid;
  grid-template-columns: repeat(auto-fit, minmax(200px, 1fr));
  gap: 20px;
  padding: 8px 0;
}

.summary-item {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.summary-label {
  font-size: 14px;
  color: #909399;
  font-weight: 500;
}

.summary-value {
  font-size: 16px;
  color: #303133;
  font-weight: 600;
}

.summary-value-highlight {
  font-size: 18px;
  color: #2080f0;
  font-weight: 700;
}

.summary-highlight {
  background: linear-gradient(135deg, rgba(24, 160, 88, 0.05) 0%, rgba(208, 48, 80, 0.05) 100%);
  padding: 12px;
  border-radius: 8px;
  border: 1px solid rgba(24, 160, 88, 0.1);
}

.summary-highlight .summary-label {
  color: #606266;
}

.table-container {
  position: relative;
  max-height: calc(100vh - 400px);
  overflow-y: auto;
  padding-bottom: 20px;
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
  padding: 60px 0 100px;
  display: flex;
  align-items: center;
  justify-content: center;
}
</style>
