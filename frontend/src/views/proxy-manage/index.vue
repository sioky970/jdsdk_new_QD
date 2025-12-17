<script setup lang="ts">
import { h, ref, reactive, computed, onMounted } from 'vue';
import type { FormInst, DataTableColumns } from 'naive-ui';
import { NTag, NSpace, NButton, NPopover, NIcon } from 'naive-ui';
import { QrCode } from '@vicons/ionicons5';
import { request } from '@/service/request';
import QRCode from 'qrcode';

defineOptions({
  name: 'ProxyManage'
});

interface Proxy {
  id: number;
  ip: string;
  port: number;
  username: string;
  password: string;
  province: string;
  city: string;
  isp: string;
  remark: string;
  qrcode_url: string;
  usage_count: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
}

interface ProxyStatistics {
  total_count: number;
  active_count: number;
  inactive_count: number;
  total_usage: number;
  avg_usage: number;
}

const loading = ref(false);
const proxyList = ref<Proxy[]>([]);
const total = ref(0);
const checkedRowKeys = ref<number[]>([]); // 选中的行
const statistics = ref<ProxyStatistics>({
  total_count: 0,
  active_count: 0,
  inactive_count: 0,
  total_usage: 0,
  avg_usage: 0
});

// 查询参数
const queryParams = reactive({
  page: 1,
  page_size: 20,
  keyword: '',
  is_active: ''
});

// 批量导入对话框
const showBatchImportModal = ref(false);
const batchImportText = ref('');
const importing = ref(false);

// 新增/编辑对话框
const showEditModal = ref(false);
const editForm = ref<Partial<Proxy>>({});
const editFormRef = ref<FormInst | null>(null);
const isEdit = ref(false);

// 是否还有更多数据
const hasMore = computed(() => {
  const totalPages = Math.ceil(total.value / queryParams.page_size);
  return queryParams.page < totalPages;
});

// 加载代理列表
async function loadProxyList(append = false) {
  loading.value = true;
  try {
    const params: any = {
      page: queryParams.page,
      page_size: queryParams.page_size
    };
    if (queryParams.keyword) {
      params.keyword = queryParams.keyword;
    }
    if (queryParams.is_active !== '') {
      params.is_active = queryParams.is_active;
    }

    const res = await request<any>({
      url: '/api/proxies',
      method: 'GET',
      params
    });

    if (res.data) {
      if (append) {
        proxyList.value = [...proxyList.value, ...(res.data.proxies || [])];
      } else {
        proxyList.value = res.data.proxies || [];
      }
      total.value = res.data.total || 0;
    }
  } catch (error) {
    window.$message?.error('加载代理列表失败');
  } finally {
    loading.value = false;
  }
}

// 加载统计信息
async function loadStatistics() {
  try {
    const res = await request<any>({
      url: '/api/proxies/statistics',
      method: 'GET'
    });

    if (res.data) {
      statistics.value = res.data;
    }
  } catch (error) {
    console.error('加载统计信息失败', error);
  }
}

// 搜索
function handleSearch() {
  queryParams.page = 1;
  proxyList.value = []; // 清空现有数据
  loadProxyList();
}

// 重置
function handleReset() {
  queryParams.page = 1;
  queryParams.keyword = '';
  queryParams.is_active = '';
  proxyList.value = []; // 清空现有数据
  loadProxyList();
}

// 加载更多数据（无限滚动）
async function loadMore() {
  if (!loading.value && hasMore.value) {
    queryParams.page++;
    await loadProxyList(true);
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

// 打开批量导入对话框
function openBatchImport() {
  batchImportText.value = '';
  showBatchImportModal.value = true;
}

// 批量导入
async function handleBatchImport() {
  if (!batchImportText.value.trim()) {
    window.$message?.warning('请输入代理数据');
    return;
  }

  importing.value = true;
  try {
    const res = await request<any>({
      url: '/api/proxies/batch-import',
      method: 'POST',
      data: {
        proxy_list: batchImportText.value
      }
    });

    if (res.data) {
      window.$message?.success(
        `批量导入完成：成功 ${res.data.success_count} 条，失败 ${res.data.fail_count} 条`
      );
      if (res.data.errors && res.data.errors.length > 0) {
        console.warn('导入错误:', res.data.errors);
      }
      showBatchImportModal.value = false;
      queryParams.page = 1;
      proxyList.value = []; // 清空现有数据
      loadProxyList();
      loadStatistics();
    }
  } catch (error: any) {
    window.$message?.error(error.message || '批量导入失败');
  } finally {
    importing.value = false;
  }
}

// 打开新增对话框
function openAdd() {
  isEdit.value = false;
  editForm.value = {
    ip: '',
    port: undefined,
    username: '',
    password: '',
    remark: ''
  };
  showEditModal.value = true;
}

// 打开编辑对话框
function openEdit(row: Proxy) {
  isEdit.value = true;
  editForm.value = { ...row };
  showEditModal.value = true;
}

// 保存代理
async function handleSave() {
  try {
    await editFormRef.value?.validate();

    const url = isEdit.value ? `/api/proxies/${editForm.value.id}` : '/api/proxies';
    const method = isEdit.value ? 'PUT' : 'POST';

    await request({
      url,
      method,
      data: editForm.value
    });

    window.$message?.success(isEdit.value ? '更新成功' : '创建成功');
    showEditModal.value = false;
    queryParams.page = 1;
    proxyList.value = []; // 清空现有数据
    loadProxyList();
    loadStatistics();
  } catch (error: any) {
    if (error.message) {
      window.$message?.error(error.message);
    }
  }
}

// 删除代理
async function handleDelete(row: Proxy) {
  window.$dialog?.warning({
    title: '确认删除',
    content: `确定要删除代理 ${row.ip}:${row.port} 吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        await request({
          url: `/api/proxies/${row.id}`,
          method: 'DELETE'
        });

        window.$message?.success('删除成功');
        queryParams.page = 1;
        proxyList.value = []; // 清空现有数据
        loadProxyList();
        loadStatistics();
      } catch (error: any) {
        window.$message?.error(error.message || '删除失败');
      }
    }
  });
}

// 切换激活状态
async function toggleActive(row: Proxy) {
  try {
    await request({
      url: `/api/proxies/${row.id}`,
      method: 'PUT',
      data: {
        is_active: !row.is_active
      }
    });

    window.$message?.success('状态更新成功');
    queryParams.page = 1;
    proxyList.value = []; // 清空现有数据
    loadProxyList();
    loadStatistics();
  } catch (error: any) {
    window.$message?.error(error.message || '状态更新失败');
  }
}

// 批量删除
async function handleBatchDelete() {
  if (checkedRowKeys.value.length === 0) {
    window.$message?.warning('请选择要删除的代理');
    return;
  }

  window.$dialog?.warning({
    title: '确认删除',
    content: `确定要删除选中的 ${checkedRowKeys.value.length} 个代理吗？`,
    positiveText: '确定',
    negativeText: '取消',
    onPositiveClick: async () => {
      try {
        const res = await request<any>({
          url: '/api/proxies/batch-delete',
          method: 'POST',
          data: {
            ids: checkedRowKeys.value
          }
        });

        window.$message?.success(res.msg || '批量删除成功');
        checkedRowKeys.value = []; // 清空选中
        queryParams.page = 1;
        proxyList.value = [];
        loadProxyList();
        loadStatistics();
      } catch (error: any) {
        window.$message?.error(error.message || '批量删除失败');
      }
    }
  });
}

// 表单验证规则
const rules = {
  ip: { required: true, message: '请输入IP地址', trigger: 'blur' },
  port: { required: true, type: 'number', message: '请输入端口', trigger: 'blur' },
  username: { required: true, message: '请输入用户名', trigger: 'blur' },
  password: { required: true, message: '请输入密码', trigger: 'blur' }
};

// 生成二维码图片
async function generateQRCodeImage(url: string): Promise<string> {
  try {
    // 使用 qrcode 库生成 base64 图片
    const qrCodeDataURL = await QRCode.toDataURL(url, {
      width: 200,
      margin: 1,
      color: {
        dark: '#000000',
        light: '#FFFFFF'
      }
    });
    return qrCodeDataURL;
  } catch (error) {
    console.error('生成二维码失败:', error);
    return '';
  }
}

// 移除分页处理函数，改用无限滚动
// function handlePageChange(page: number) {
//   queryParams.page = page;
//   loadProxyList();
// }
// function handlePageSizeChange(pageSize: number) {
//   queryParams.page_size = pageSize;
//   queryParams.page = 1;
//   loadProxyList();
// }

// 表格列定义
const columns: DataTableColumns<Proxy> = [
  { type: 'selection' },
  { title: 'ID', key: 'id', width: 60 },
  { title: 'IP地址', key: 'ip', width: 140 },
  { title: '端口', key: 'port', width: 80 },
  { title: '用户名', key: 'username', width: 120 },
  { title: '密码', key: 'password', width: 100 },
  {
    title: '地理位置',
    key: 'location',
    width: 180,
    render: row => `${row.province || ''} ${row.city || ''}`
  },
  { title: 'ISP', key: 'isp', width: 100 },
  { title: '使用次数', key: 'usage_count', width: 90 },
  {
    title: '二维码',
    key: 'qrcode',
    width: 80,
    render: row => {
      if (!row.qrcode_url) return null;
      
      return h(
        NPopover,
        { trigger: 'hover', placement: 'left' },
        {
          trigger: () =>
            h(
              NIcon,
              { size: 24, style: { cursor: 'pointer', color: '#18a058' } },
              { default: () => h(QrCode) }
            ),
          default: () =>
            h('div', { class: 'qrcode-popover' }, [
              h('img', {
                src: '',
                alt: 'QR Code',
                style: { width: '200px', height: '200px' },
                onVnodeMounted: async (vnode: any) => {
                  const qrImage = await generateQRCodeImage(row.qrcode_url);
                  if (vnode.el && qrImage) {
                    vnode.el.src = qrImage;
                  }
                }
              }),
              h('p', { style: { marginTop: '8px', fontSize: '12px', color: '#666', textAlign: 'center' } }, 
                `${row.ip}:${row.port}`
              )
            ])
        }
      );
    }
  },
  { title: '备注', key: 'remark', width: 150, ellipsis: { tooltip: true } },
  {
    title: '状态',
    key: 'is_active',
    width: 80,
    render: row =>
      h(
        NTag,
        { type: row.is_active ? 'success' : 'error' },
        { default: () => (row.is_active ? '激活' : '禁用') }
      )
  },
  {
    title: '操作',
    key: 'actions',
    width: 200,
    fixed: 'right',
    render: row =>
      h(
        NSpace,
        {},
        {
          default: () => [
            h(
              NButton,
              {
                size: 'small',
                onClick: () => openEdit(row)
              },
              { default: () => '编辑' }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: row.is_active ? 'warning' : 'success',
                onClick: () => toggleActive(row)
              },
              { default: () => (row.is_active ? '禁用' : '启用') }
            ),
            h(
              NButton,
              {
                size: 'small',
                type: 'error',
                onClick: () => handleDelete(row)
              },
              { default: () => '删除' }
            )
          ]
        }
      )
  }
];

onMounted(() => {
  loadProxyList();
  loadStatistics();
});
</script>

<template>
  <div class="h-full overflow-hidden flex flex-col gap-4">
    <!-- 统计卡片 -->
    <NCard title="代理池统计" :bordered="false">
      <NGrid :cols="5" :x-gap="16">
        <NGridItem>
          <NStatistic label="总代理数" :value="statistics.total_count" />
        </NGridItem>
        <NGridItem>
          <NStatistic label="激活中" :value="statistics.active_count" />
        </NGridItem>
        <NGridItem>
          <NStatistic label="已禁用" :value="statistics.inactive_count" />
        </NGridItem>
        <NGridItem>
          <NStatistic label="总使用次数" :value="statistics.total_usage" />
        </NGridItem>
        <NGridItem>
          <NStatistic label="平均使用" :value="statistics.avg_usage.toFixed(2)" />
        </NGridItem>
      </NGrid>
    </NCard>

    <!-- 主内容卡片 -->
    <NCard title="SK5代理池管理" :bordered="false" class="flex-1 overflow-hidden flex flex-col">
      <!-- 搜索区域 -->
      <div class="mb-4 flex gap-4">
        <NInput
          v-model:value="queryParams.keyword"
          placeholder="搜索IP或备注"
          clearable
          style="width: 200px"
          @keyup.enter="handleSearch"
        />
        <NSelect
          v-model:value="queryParams.is_active"
          placeholder="激活状态"
          clearable
          style="width: 120px"
          :options="[
            { label: '全部', value: '' },
            { label: '激活', value: 'true' },
            { label: '禁用', value: 'false' }
          ]"
        />
        <NButton type="primary" @click="handleSearch">搜索</NButton>
        <NButton @click="handleReset">重置</NButton>
        <div class="flex-1" />
        <NButton
          v-if="checkedRowKeys.length > 0"
          type="error"
          @click="handleBatchDelete"
        >
          批量删除 ({{ checkedRowKeys.length }})
        </NButton>
        <NButton type="info" @click="openBatchImport">批量导入</NButton>
        <NButton type="primary" @click="openAdd">新增代理</NButton>
      </div>

      <!-- 表格 -->
      <div class="table-container" @scroll="handleScroll">
        <NDataTable
          v-model:checked-row-keys="checkedRowKeys"
          :columns="columns"
          :data="proxyList"
          :loading="loading"
          :scroll-x="1500"
          :single-line="false"
          :row-key="(row: Proxy) => row.id"
        />

        <!-- 加载更多提示 -->
        <div v-if="hasMore" class="load-more-trigger">
          <div v-if="loading" class="text-center py-4 text-gray-400">
            加载中...
          </div>
        </div>

        <!-- 底部空白区域 -->
        <div v-if="!hasMore && proxyList.length > 0" class="bottom-spacer">
          <div class="text-center text-gray-400 text-sm">
            已加载全部 {{ total }} 条数据
          </div>
        </div>
      </div>
    </NCard>

    <!-- 批量导入对话框 -->
    <NModal v-model:show="showBatchImportModal" preset="card" title="批量导入代理" style="width: 600px">
      <NAlert type="info" class="mb-4">
        请按照格式输入：IP|端口|用户名|密码，每行一条<br />
        示例：42.101.12.24|11011|chtJZ0530135|3678
      </NAlert>
      <NInput
        v-model:value="batchImportText"
        type="textarea"
        placeholder="请输入代理数据，每行一条"
        :rows="10"
      />
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showBatchImportModal = false">取消</NButton>
          <NButton type="primary" :loading="importing" @click="handleBatchImport">导入</NButton>
        </div>
      </template>
    </NModal>

    <!-- 新增/编辑对话框 -->
    <NModal
      v-model:show="showEditModal"
      preset="card"
      :title="isEdit ? '编辑代理' : '新增代理'"
      style="width: 500px"
    >
      <NForm ref="editFormRef" :model="editForm" :rules="rules" label-placement="left" label-width="80">
        <NFormItem label="IP地址" path="ip">
          <NInput v-model:value="editForm.ip" placeholder="请输入IP地址" />
        </NFormItem>
        <NFormItem label="端口" path="port">
          <NInputNumber
            v-model:value="editForm.port"
            placeholder="请输入端口"
            :min="1"
            :max="65535"
            style="width: 100%"
          />
        </NFormItem>
        <NFormItem label="用户名" path="username">
          <NInput v-model:value="editForm.username" placeholder="请输入用户名" />
        </NFormItem>
        <NFormItem label="密码" path="password">
          <NInput v-model:value="editForm.password" type="password" show-password-on="click" placeholder="请输入密码" />
        </NFormItem>
        <NFormItem label="备注">
          <NInput v-model:value="editForm.remark" type="textarea" :rows="3" placeholder="请输入备注" />
        </NFormItem>
      </NForm>
      <template #footer>
        <div class="flex justify-end gap-2">
          <NButton @click="showEditModal = false">取消</NButton>
          <NButton type="primary" @click="handleSave">保存</NButton>
        </div>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
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
  padding: 20px 0;
}
</style>
