<script setup lang="ts">
import { h, onMounted, ref } from 'vue';
import { NButton, NCard, NDataTable, NInput, NInputNumber, NModal, NSpace, NSwitch, NTag, NForm, NFormItem, NGrid, NGi, NInputGroup, NInputGroupLabel, useMessage, useDialog, type DataTableColumn } from 'naive-ui';
import { fetchTaskTypesAdmin, updateTaskType, fetchAnnouncement, updateAnnouncement, fetchDeviceAuthKey, updateDeviceAuthKey, type TaskTypeAdmin, type UpdateTaskTypeRequest } from '@/service/api';

defineOptions({
  name: 'SystemSettings'
});

const message = useMessage();
const dialog = useDialog();

// 任务类型数据
const taskTypes = ref<TaskTypeAdmin[]>([]);
const loading = ref(false);

// 公告数据
const announcement = ref('');
const announcementLoading = ref(false);
const announcementSaving = ref(false);

// 设备密钥数据
const deviceKey = ref('');
const deviceKeyLoading = ref(false);
const deviceKeySaving = ref(false);

// 编辑弹窗
const showEditModal = ref(false);
const editLoading = ref(false);
const editingType = ref<TaskTypeAdmin | null>(null);
const editForm = ref<UpdateTaskTypeRequest>({
  type_name: '',
  jingdou_price: 0,
  is_active: true,
  execute_multiplier: 1,
  time_slot1_start: '',
  time_slot1_end: '',
  time_slot2_start: '',
  time_slot2_end: ''
});

// 表格列定义
const columns: DataTableColumn<TaskTypeAdmin>[] = [
  { title: 'ID', key: 'id', width: 60 },
  { title: '类型代码', key: 'type_code', width: 120 },
  { 
    title: '类型名称', 
    key: 'type_name', 
    width: 140,
    render: (row) => h('span', { style: { fontWeight: 600 } }, row.type_name)
  },
  { 
    title: '京豆价格', 
    key: 'jingdou_price', 
    width: 90,
    render: (row) => h('span', { style: { color: '#2080f0', fontWeight: 600 } }, `${row.jingdou_price} 豆`)
  },
  { 
    title: '执行倍数', 
    key: 'execute_multiplier', 
    width: 90,
    render: (row) => h(NTag, { 
      type: row.execute_multiplier > 1 ? 'warning' : 'default',
      size: 'small'
    }, () => `×${row.execute_multiplier || 1}`)
  },
  { 
    title: '状态', 
    key: 'is_active', 
    width: 80,
    render: (row) => h(NTag, { 
      type: row.is_active ? 'success' : 'error',
      size: 'small'
    }, () => row.is_active ? '启用' : '停用')
  },
  { 
    title: '时间限制', 
    key: 'time_slots', 
    width: 180,
    render: (row) => {
      if (!row.has_time_limit || !row.time_slots?.length) {
        return h('span', { style: { color: '#909399' } }, '无限制');
      }
      return h(NSpace, { size: 'small' }, () => 
        row.time_slots.map(slot => h(NTag, { size: 'small', type: 'info' }, () => slot))
      );
    }
  },
  {
    title: '操作',
    key: 'actions',
    width: 100,
    fixed: 'right',
    render: (row) => h(NButton, { 
      size: 'small', 
      type: 'primary', 
      onClick: () => handleEdit(row) 
    }, () => '编辑')
  }
];

// 加载任务类型
async function loadTaskTypes() {
  loading.value = true;
  console.log('开始加载任务类型...');
  const { data, error } = await fetchTaskTypesAdmin();
  console.log('加载任务类型结果:', { data, error });
  loading.value = false;
  
  if (!error && data) {
    // 强制替换数组引用以触发响应式更新
    taskTypes.value = [...(data.task_types || [])];
    console.log('任务类型已更新:', taskTypes.value);
  }
}

// 加载公告
async function loadAnnouncement() {
  announcementLoading.value = true;
  const { data, error } = await fetchAnnouncement();
  if (!error && data) {
    announcement.value = data.announcement || '';
  }
  announcementLoading.value = false;
}

// 编辑任务类型
function handleEdit(row: TaskTypeAdmin) {
  editingType.value = row;
  editForm.value = {
    type_name: row.type_name,
    jingdou_price: row.jingdou_price,
    is_active: row.is_active,
    execute_multiplier: row.execute_multiplier || 1,
    time_slot1_start: row.time_slot1_start || '',
    time_slot1_end: row.time_slot1_end || '',
    time_slot2_start: row.time_slot2_start || '',
    time_slot2_end: row.time_slot2_end || ''
  };
  showEditModal.value = true;
}

// 保存任务类型
async function saveTaskType() {
  if (!editingType.value) return;
  
  editLoading.value = true;
  console.log('保存任务类型请求:', editingType.value.id, editForm.value);
  const { error, data } = await updateTaskType(editingType.value.id, editForm.value);
  console.log('保存任务类型响应:', { error, data });
  editLoading.value = false;
  
  if (!error) {
    message.success('任务类型配置已更新');
    showEditModal.value = false;
    await loadTaskTypes();
  } else {
    console.error('更新失败:', error);
    message.error('更新失败，请重试');
  }
}

// 快速切换启用状态
async function toggleActive(row: TaskTypeAdmin) {
  const { error } = await updateTaskType(row.id, { is_active: !row.is_active });
  if (!error) {
    message.success(row.is_active ? '已停用' : '已启用');
    await loadTaskTypes();
  } else {
    message.error('操作失败，请重试');
  }
}

// 保存公告
async function saveAnnouncement() {
  announcementSaving.value = true;
  const { error } = await updateAnnouncement(announcement.value);
  if (!error) {
    message.success('公告已更新');
  }
  announcementSaving.value = false;
}

// 确认清空公告
function confirmClearAnnouncement() {
  dialog.warning({
    title: '确认清空',
    content: '确定要清空系统公告吗？',
    positiveText: '确认清空',
    negativeText: '取消',
    onPositiveClick: async () => {
      announcement.value = '';
      await saveAnnouncement();
    }
  });
}

// 加载设备密钥
async function loadDeviceKey() {
  deviceKeyLoading.value = true;
  const { data, error } = await fetchDeviceAuthKey();
  if (!error && data) {
    deviceKey.value = data.device_key;
  }
  deviceKeyLoading.value = false;
}

// 保存设备密钥
async function saveDeviceKey() {
  if (!deviceKey.value || deviceKey.value.length < 6) {
    message.error('设备密钥长度必须大于6位');
    return;
  }
  
  deviceKeySaving.value = true;
  const { error } = await updateDeviceAuthKey(deviceKey.value);
  if (!error) {
    message.success('设备密钥更新成功');
  }
  deviceKeySaving.value = false;
}

// 确认重置设备密钥
function confirmResetDeviceKey() {
  dialog.warning({
    title: '确认重置',
    content: '确定要重置设备密钥为默认值吗？',
    positiveText: '确认重置',
    negativeText: '取消',
    onPositiveClick: async () => {
      deviceKey.value = 'KKNN778899';
      await saveDeviceKey();
    }
  });
}

onMounted(() => {
  loadTaskTypes();
  loadAnnouncement();
  loadDeviceKey();
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 页面标题 -->
    <NCard :bordered="false" class="card-wrapper">
      <div class="flex items-center gap-12px">
        <span class="text-20px font-bold">系统参数设置</span>
        <NTag type="warning" size="small">仅管理员</NTag>
      </div>
      <p class="text-14px text-gray-500 mt-8px">
        管理任务类型配置、执行倍数、价格和系统公告等核心参数
      </p>
    </NCard>

    <!-- 任务类型配置 -->
    <NCard title="任务类型配置" :bordered="false" class="card-wrapper">
      <template #header-extra>
        <NButton type="primary" size="small" @click="loadTaskTypes" :loading="loading">
          刷新
        </NButton>
      </template>

      <NDataTable
        :columns="columns"
        :data="taskTypes"
        :loading="loading"
        :bordered="false"
        :single-line="false"
        size="small"
        :scroll-x="900"
        :row-key="(row: TaskTypeAdmin) => row.id"
      />

      <div class="mt-12px text-12px text-gray-400">
        <p>• 执行倍数：设备执行一次任务时，系统自动乘以该倍数提交执行量</p>
        <p>• 时间限制：普通用户只能在指定时间段内创建该类型任务</p>
      </div>
    </NCard>

    <!-- 系统公告 -->
    <NCard title="系统公告" :bordered="false" class="card-wrapper">
      <template #header-extra>
        <NSpace>
          <NButton size="small" @click="loadAnnouncement" :loading="announcementLoading">
            刷新
          </NButton>
          <NButton type="error" size="small" @click="confirmClearAnnouncement" :disabled="!announcement">
            清空
          </NButton>
        </NSpace>
      </template>

      <NInput 
        v-model:value="announcement" 
        type="textarea" 
        :rows="4"
        placeholder="请输入系统公告内容，留空则不显示公告"
        :loading="announcementLoading"
      />

      <div class="mt-12px flex items-center justify-between">
        <span class="text-12px text-gray-400">公告将显示在用户首页和管理员首页顶部</span>
        <NButton type="primary" @click="saveAnnouncement" :loading="announcementSaving">
          保存公告
        </NButton>
      </div>
    </NCard>

    <!-- 设备认证密钥 -->
    <NCard title="设备认证密钥" :bordered="false" class="card-wrapper">
      <template #header-extra>
        <NSpace>
          <NButton size="small" @click="loadDeviceKey" :loading="deviceKeyLoading">
            刷新
          </NButton>
          <NButton type="warning" size="small" @click="confirmResetDeviceKey">
            重置为默认值
          </NButton>
        </NSpace>
      </template>

      <div class="mb-16px">
        <NTag type="info" size="small">
          <template #icon>
            <span class="i-mdi-information"></span>
          </template>
          设备端API认证密钥，用于iOS设备请求任务和代理
        </NTag>
      </div>

      <NInputGroup>
        <NInputGroupLabel>当前密钥</NInputGroupLabel>
        <NInput 
          v-model:value="deviceKey" 
          type="password"
          show-password-on="click"
          placeholder="请输入设备密钥（至少6位）"
          :loading="deviceKeyLoading"
          :minlength="6"
        />
        <NButton type="primary" @click="saveDeviceKey" :loading="deviceKeySaving">
          保存
        </NButton>
      </NInputGroup>

      <div class="mt-12px space-y-8px">
        <div class="flex items-start gap-8px text-12px text-gray-400">
          <span class="i-mdi-alert-circle-outline mt-2px flex-shrink-0"></span>
          <span>密钥长度必须大于6位，建议使用数字和字母组合</span>
        </div>
        <div class="flex items-start gap-8px text-12px text-orange-500">
          <span class="i-mdi-alert-outline mt-2px flex-shrink-0"></span>
          <span>修改密钥后需同步更新iOS脚本配置文件中的 DEVICE_AUTH_KEY，否则设备无法连接</span>
        </div>
        <div class="flex items-start gap-8px text-12px text-gray-400">
          <span class="i-mdi-file-code-outline mt-2px flex-shrink-0"></span>
          <span>iOS脚本配置文件路径：<code class="bg-gray-100 dark:bg-gray-800 px-4px rounded">jdios/config.py</code></span>
        </div>
      </div>
    </NCard>

    <!-- 编辑任务类型弹窗 -->
    <NModal
      v-model:show="showEditModal"
      preset="card"
      :title="`编辑任务类型 - ${editingType?.type_name}`"
      style="width: 600px"
      :mask-closable="false"
    >
      <NForm label-placement="left" label-width="100px">
        <NGrid :cols="2" :x-gap="16">
          <NGi>
            <NFormItem label="类型代码">
              <NInput :value="editingType?.type_code" disabled />
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="类型名称">
              <NInput v-model:value="editForm.type_name" placeholder="请输入类型名称" />
            </NFormItem>
          </NGi>
        </NGrid>

        <NGrid :cols="2" :x-gap="16">
          <NGi>
            <NFormItem label="京豆价格">
              <NInputNumber 
                v-model:value="editForm.jingdou_price" 
                :min="1" 
                :max="10000"
                style="width: 100%"
              >
                <template #suffix>京豆/次</template>
              </NInputNumber>
            </NFormItem>
          </NGi>
          <NGi>
            <NFormItem label="执行倍数">
              <NInputNumber 
                v-model:value="editForm.execute_multiplier" 
                :min="1" 
                :max="100"
                style="width: 100%"
              >
                <template #suffix>倍</template>
              </NInputNumber>
            </NFormItem>
          </NGi>
        </NGrid>

        <NFormItem label="启用状态">
          <NSwitch v-model:value="editForm.is_active">
            <template #checked>启用</template>
            <template #unchecked>停用</template>
          </NSwitch>
        </NFormItem>

        <NFormItem label="时间段1">
          <NGrid :cols="2" :x-gap="8">
            <NGi>
              <NInput v-model:value="editForm.time_slot1_start" placeholder="开始 HH:MM" />
            </NGi>
            <NGi>
              <NInput v-model:value="editForm.time_slot1_end" placeholder="结束 HH:MM" />
            </NGi>
          </NGrid>
        </NFormItem>

        <NFormItem label="时间段2">
          <NGrid :cols="2" :x-gap="8">
            <NGi>
              <NInput v-model:value="editForm.time_slot2_start" placeholder="开始 HH:MM" />
            </NGi>
            <NGi>
              <NInput v-model:value="editForm.time_slot2_end" placeholder="结束 HH:MM" />
            </NGi>
          </NGrid>
        </NFormItem>

        <div class="text-12px text-gray-400 mb-16px">
          提示：时间格式为 HH:MM（如 08:00），留空表示不限制时间
        </div>
      </NForm>

      <template #footer>
        <NSpace justify="end">
          <NButton @click="showEditModal = false">取消</NButton>
          <NButton type="primary" :loading="editLoading" @click="saveTaskType">
            保存
          </NButton>
        </NSpace>
      </template>
    </NModal>
  </div>
</template>

<style scoped>
.card-wrapper {
  border-radius: 8px;
}
</style>
