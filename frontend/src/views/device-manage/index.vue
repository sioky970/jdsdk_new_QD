<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import {
  NButton,
  NCard,
  NEmpty,
  NGrid,
  NGi,
  NPopconfirm,
  NSpace,
  NSpin,
  NStatistic,
  NTag,
  NTooltip,
  useDialog,
  useMessage
} from 'naive-ui';
import {
  fetchDevices,
  fetchDeviceStatistics,
  clearAllDevices
} from '@/service/api';
import type { Device, DeviceStatistics } from '@/service/api';

defineOptions({
  name: 'DeviceManage'
});

const message = useMessage();
const dialog = useDialog();

// 状态数据
const loading = ref(false);
const devices = ref<Device[]>([]);
const total = ref(0);
const statistics = ref<DeviceStatistics | null>(null);
const statsLoading = ref(false);

// 分页
const page = ref(1);
const pageSize = ref(100); // 卡片形式一次加载更多

// 自动刷新定时器
let refreshTimer: ReturnType<typeof setInterval> | null = null;

// 状态配置
const statusConfig: Record<string, { color: string; bgColor: string; label: string }> = {
  online: { color: 'rgb(var(--success-color))', bgColor: 'rgba(var(--success-color), 0.1)', label: '在线' },
  working: { color: 'rgb(var(--warning-color))', bgColor: 'rgba(var(--warning-color), 0.1)', label: '工作中' },
  idle: { color: 'rgb(var(--primary-color))', bgColor: 'rgba(var(--primary-color), 0.1)', label: '空闲' },
  offline: { color: 'rgba(var(--base-text-color), 0.5)', bgColor: 'rgba(var(--base-text-color), 0.05)', label: '离线' }
};

// 获取设备状态样式
function getStatusStyle(status: string) {
  const config = statusConfig[status] || statusConfig.offline;
  return {
    color: config.color,
    backgroundColor: config.bgColor,
    padding: '2px 8px',
    borderRadius: '4px',
    fontSize: '12px'
  };
}

// 计算在线时长
function getOnlineDuration(lastHeartbeat: string | null): string {
  if (!lastHeartbeat) return '-';
  const now = new Date();
  const last = new Date(lastHeartbeat);
  const diff = now.getTime() - last.getTime();
  
  if (diff < 0) return '刚刚上线';
  
  const seconds = Math.floor(diff / 1000);
  if (seconds < 60) return `${seconds}秒前`;
  
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes}分钟前`;
  
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours}小时前`;
  
  const days = Math.floor(hours / 24);
  return `${days}天前`;
}

// 格式化时间
function formatTime(time: string | null): string {
  if (!time) return '-';
  const date = new Date(time);
  return date.toLocaleString('zh-CN', {
    month: '2-digit',
    day: '2-digit',
    hour: '2-digit',
    minute: '2-digit'
  });
}

// 判断设备类型（优先使用 device_type，兼容旧字段 os_info）
function getDeviceType(device: Device): 'android' | 'ios' {
  // 优先使用新字段
  if (device.device_type === 'ios' || device.device_type === 'android') {
    return device.device_type;
  }
  // 兼容旧字段 os_info
  if (device.os_info) {
    const lower = device.os_info.toLowerCase();
    if (lower.includes('ios') || lower.includes('iphone') || lower.includes('ipad') || lower.includes('apple')) {
      return 'ios';
    }
  }
  return 'android';
}

// 加载设备列表
async function loadDevices() {
  loading.value = true;
  const { data, error } = await fetchDevices({
    page: page.value,
    page_size: pageSize.value
  });
  loading.value = false;

  if (!error && data) {
    devices.value = data.items || [];
    total.value = data.total || 0;
  }
}

// 加载统计数据
async function loadStatistics() {
  statsLoading.value = true;
  const { data, error } = await fetchDeviceStatistics();
  statsLoading.value = false;

  if (!error && data) {
    statistics.value = data;
  }
}

// 清空所有设备
function handleClearAll() {
  dialog.warning({
    title: '危险操作',
    content: '确定要清空所有设备记录吗？此操作不可恢复！',
    positiveText: '确认清空',
    negativeText: '取消',
    onPositiveClick: async () => {
      const { data, error } = await clearAllDevices();
      if (!error) {
        message.success(`已清空 ${data?.deleted_count || 0} 台设备`);
        loadDevices();
        loadStatistics();
      }
    }
  });
}

// 刷新
function handleRefresh() {
  loadDevices();
  loadStatistics();
}

// 启动自动刷新
function startAutoRefresh() {
  refreshTimer = setInterval(() => {
    loadDevices();
    loadStatistics();
  }, 10000); // 每10秒刷新一次
}

onMounted(() => {
  loadDevices();
  loadStatistics();
  startAutoRefresh();
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<template>
  <div class="min-h-500px flex-col-stretch gap-16px overflow-hidden lt-sm:overflow-auto">
    <!-- 页面标题 -->
    <NCard :bordered="false" class="card-wrapper">
      <div class="flex items-center justify-between">
        <div>
          <div class="flex items-center gap-12px">
            <span class="text-20px font-bold">设备管理</span>
            <NTag type="warning" size="small">仅管理员</NTag>
            <NTag type="info" size="small">自动刷新</NTag>
          </div>
          <p class="text-14px text-gray-500 mt-8px">
            查看设备状态、地理位置、任务执行情况，设备主动访问服务器后自动注册
          </p>
        </div>
        <NSpace>
          <NButton type="primary" @click="handleRefresh" :loading="loading">
            刷新
          </NButton>
          <NPopconfirm @positive-click="handleClearAll">
            <template #trigger>
              <NButton type="error">清空所有</NButton>
            </template>
            确定要清空所有设备吗？
          </NPopconfirm>
        </NSpace>
      </div>
    </NCard>

    <!-- 统计卡片 -->
    <NCard :bordered="false" class="card-wrapper">
      <NGrid :cols="5" :x-gap="16" responsive="screen" item-responsive>
        <NGi>
          <NStatistic label="设备总数" :value="statistics?.total_devices || 0">
            <template #suffix>台</template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="在线设备" :value="statistics?.online_devices || 0">
            <template #prefix>
              <span class="text-green-500">●</span>
            </template>
            <template #suffix>台</template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="工作中" :value="statistics?.working_devices || 0">
            <template #prefix>
              <span class="text-yellow-500">▶</span>
            </template>
            <template #suffix>台</template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="空闲" :value="statistics?.idle_devices || 0">
            <template #prefix>
              <span class="text-blue-500">○</span>
            </template>
            <template #suffix>台</template>
          </NStatistic>
        </NGi>
        <NGi>
          <NStatistic label="离线设备" :value="statistics?.offline_devices || 0">
            <template #prefix>
              <span class="text-gray-400">○</span>
            </template>
            <template #suffix>台</template>
          </NStatistic>
        </NGi>
      </NGrid>
    </NCard>

    <!-- 设备卡片列表 -->
    <NCard title="设备列表" :bordered="false" class="card-wrapper flex-1">
      <template #header-extra>
        <span class="text-gray-500 text-14px">共 {{ total }} 台设备</span>
      </template>
      
      <NSpin :show="loading">
        <div v-if="devices.length === 0 && !loading" class="py-40px">
          <NEmpty description="暂无设备，等待设备主动访问服务器" />
        </div>
        
        <NGrid v-else :cols="4" :x-gap="12" :y-gap="12" responsive="screen" item-responsive>
          <NGi v-for="device in devices" :key="device.id" span="4 m:2 l:1">
            <NCard 
              :bordered="true" 
              class="device-card"
              :class="{ 'device-blocked': device.is_blocked, 'device-offline': device.status === 'offline' }"
            >
              <!-- 左右布局 -->
              <div class="device-layout">
                <!-- 左侧：设备图标 -->
                <div class="device-left">
                  <div class="device-icon" :class="[`status-${device.status}`, `type-${getDeviceType(device)}`]">
                    <!-- Android 图标 -->
                    <svg v-if="getDeviceType(device) === 'android'" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M6 18c0 .55.45 1 1 1h1v3.5c0 .83.67 1.5 1.5 1.5s1.5-.67 1.5-1.5V19h2v3.5c0 .83.67 1.5 1.5 1.5s1.5-.67 1.5-1.5V19h1c.55 0 1-.45 1-1V8H6v10zM3.5 8C2.67 8 2 8.67 2 9.5v7c0 .83.67 1.5 1.5 1.5S5 17.33 5 16.5v-7C5 8.67 4.33 8 3.5 8zm17 0c-.83 0-1.5.67-1.5 1.5v7c0 .83.67 1.5 1.5 1.5s1.5-.67 1.5-1.5v-7c0-.83-.67-1.5-1.5-1.5zm-4.97-5.84l1.3-1.3c.2-.2.2-.51 0-.71-.2-.2-.51-.2-.71 0l-1.48 1.48C13.85 1.23 12.95 1 12 1c-.96 0-1.86.23-2.66.63L7.85.15c-.2-.2-.51-.2-.71 0-.2.2-.2.51 0 .71l1.31 1.31C6.97 3.26 6 5.01 6 7h12c0-1.99-.97-3.75-2.47-4.84zM10 5H9V4h1v1zm5 0h-1V4h1v1z"/>
                    </svg>
                    <!-- iOS 图标 -->
                    <svg v-else xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
                      <path d="M18.71 19.5c-.83 1.24-1.71 2.45-3.05 2.47-1.34.03-1.77-.79-3.29-.79-1.53 0-2 .77-3.27.82-1.31.05-2.3-1.32-3.14-2.53C4.25 17 2.94 12.45 4.7 9.39c.87-1.52 2.43-2.48 4.12-2.51 1.28-.02 2.5.87 3.29.87.78 0 2.26-1.07 3.81-.91.65.03 2.47.26 3.64 1.98-.09.06-2.17 1.28-2.15 3.81.03 3.02 2.65 4.03 2.68 4.04-.03.07-.42 1.44-1.38 2.83M13 3.5c.73-.83 1.94-1.46 2.94-1.5.13 1.17-.34 2.35-1.04 3.19-.69.85-1.83 1.51-2.95 1.42-.15-1.15.41-2.35 1.05-3.11z"/>
                    </svg>
                  </div>
                  <span class="status-badge" :style="getStatusStyle(device.status)">
                    {{ statusConfig[device.status]?.label }}
                  </span>

                </div>
                
                <!-- 右侧：信息文字 -->
                <div class="device-right">
                  <!-- 设备名称和ID -->
                  <div class="device-header">
                    <div class="device-name" :title="device.device_name || device.device_id">
                      {{ device.device_name || device.device_id }}
                    </div>
                    <NTooltip>
                      <template #trigger>
                        <span class="device-id">#{{ device.id }}</span>
                      </template>
                      设备ID: {{ device.device_id }}
                    </NTooltip>
                  </div>
                  
                  <!-- 任务执行统计 -->
                  <div class="stats-row">
                    <div class="stat-item">
                      <span class="stat-value">{{ device.task_count || 0 }}</span>
                      <span class="stat-label">总执行</span>
                    </div>
                    <div class="stat-divider"></div>
                    <div class="stat-item">
                      <span class="stat-value">{{ (device.hourly_rate || 0).toFixed(1) }}</span>
                      <span class="stat-label">次/小时</span>
                    </div>
                    <div class="stat-divider"></div>
                    <div class="stat-item">
                      <span class="stat-value">{{ Math.round(device.daily_estimate || 0) }}</span>
                      <span class="stat-label">预估/天</span>
                    </div>
                  </div>
                  
                  <!-- 设备信息 -->
                  <div class="device-info-grid">
                    <div class="info-row">
                      <span class="info-label">型号</span>
                      <span class="info-value">{{ device.device_model || device.os_info || '-' }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">系统</span>
                      <span class="info-value">{{ device.os_version || '-' }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">位置</span>
                      <span class="info-value">{{ device.location || '未知' }}</span>
                    </div>
                    <div class="info-row">
                      <span class="info-label">活跃</span>
                      <span class="info-value">{{ getOnlineDuration(device.last_heartbeat) }}</span>
                    </div>
                  </div>
                </div>
              </div>
            </NCard>
          </NGi>
        </NGrid>
      </NSpin>
    </NCard>
  </div>
</template>

<style scoped>
.card-wrapper {
  border-radius: 8px;
}

.device-card {
  transition: all 0.3s ease;
  height: 100%;
}

.device-card:hover {
  box-shadow: 0 4px 12px rgba(0, 0, 0, 0.1);
  transform: translateY(-2px);
}

.device-blocked {
  border-color: rgb(var(--error-color));
  background-color: rgba(var(--error-color), 0.05);
}

.device-offline {
  opacity: 0.7;
}

/* 左右布局 */
.device-layout {
  display: flex;
  gap: 12px;
}

/* 左侧图标区域 */
.device-left {
  display: flex;
  flex-direction: column;
  align-items: center;
  min-width: 52px;
}

.device-icon {
  width: 44px;
  height: 44px;
  border-radius: 10px;
  display: flex;
  align-items: center;
  justify-content: center;
  margin-bottom: 6px;
}

.device-icon svg {
  width: 26px;
  height: 26px;
}

/* Android 图标样式 */
.device-icon.type-android.status-online {
  background: linear-gradient(135deg, rgba(61, 220, 132, 0.2), rgba(61, 220, 132, 0.1));
  color: #3DDC84;
}

.device-icon.type-android.status-working {
  background: linear-gradient(135deg, rgba(61, 220, 132, 0.25), rgba(var(--warning-color), 0.15));
  color: #3DDC84;
}

.device-icon.type-android.status-idle {
  background: linear-gradient(135deg, rgba(61, 220, 132, 0.15), rgba(61, 220, 132, 0.08));
  color: #3DDC84;
}

.device-icon.type-android.status-offline {
  background: rgba(var(--base-text-color), 0.06);
  color: rgba(var(--base-text-color), 0.4);
}

/* iOS 图标样式 */
.device-icon.type-ios.status-online {
  background: linear-gradient(135deg, rgba(var(--base-text-color), 0.12), rgba(var(--base-text-color), 0.06));
  color: rgb(var(--base-text-color));
}

.device-icon.type-ios.status-working {
  background: linear-gradient(135deg, rgba(var(--base-text-color), 0.15), rgba(var(--warning-color), 0.1));
  color: rgb(var(--base-text-color));
}

.device-icon.type-ios.status-idle {
  background: linear-gradient(135deg, rgba(var(--base-text-color), 0.1), rgba(var(--base-text-color), 0.05));
  color: rgba(var(--base-text-color), 0.8);
}

.device-icon.type-ios.status-offline {
  background: rgba(var(--base-text-color), 0.06);
  color: rgba(var(--base-text-color), 0.4);
}

.status-badge {
  font-size: 10px;
  font-weight: 500;
  padding: 2px 6px !important;
}

/* 右侧信息区域 */
.device-right {
  flex: 1;
  min-width: 0;
  display: flex;
  flex-direction: column;
  gap: 6px;
}

.device-header {
  display: flex;
  align-items: center;
  justify-content: space-between;
  gap: 8px;
}

.device-name {
  font-size: 15px;
  font-weight: 600;
  color: rgb(var(--base-text-color));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  flex: 1;
}

.device-id {
  font-size: 11px;
  color: rgba(var(--base-text-color), 0.4);
  cursor: help;
  flex-shrink: 0;
}

/* 统计行 */
.stats-row {
  display: flex;
  align-items: center;
  background: rgba(var(--base-text-color), 0.04);
  border-radius: 6px;
  padding: 6px 10px;
  gap: 8px;
}

.stat-item {
  display: flex;
  flex-direction: column;
  align-items: center;
  flex: 1;
}

.stat-value {
  font-size: 15px;
  font-weight: 700;
  color: rgb(var(--primary-color));
  line-height: 1.2;
}

.stat-label {
  font-size: 10px;
  color: rgba(var(--base-text-color), 0.5);
  margin-top: 2px;
}

.stat-divider {
  width: 1px;
  height: 22px;
  background: rgba(var(--base-text-color), 0.1);
}

/* 设备信息网格 */
.device-info-grid {
  display: grid;
  grid-template-columns: 1fr 1fr;
  gap: 4px 12px;
}

.info-row {
  display: flex;
  align-items: center;
  gap: 4px;
}

.info-label {
  font-size: 11px;
  color: rgba(var(--base-text-color), 0.5);
  min-width: 28px;
  flex-shrink: 0;
}

.info-value {
  font-size: 12px;
  color: rgb(var(--base-text-color));
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
  font-weight: 500;
}

</style>
