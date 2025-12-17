<script setup lang="ts">
import { onMounted, onUnmounted, ref, computed } from 'vue';
import { fetchJingdouStats } from '@/service/api';
import type { JingdouStatsResponse } from '@/service/api';

const stats = ref<JingdouStatsResponse | null>(null);
const loading = ref(true);
let refreshTimer: ReturnType<typeof setInterval> | null = null;

async function loadStats() {
  loading.value = true;
  const { data, error } = await fetchJingdouStats();
  if (!error && data) {
    stats.value = data;
  }
  loading.value = false;
}

// 暴露方法供父组件调用
defineExpose({
  refresh: loadStats
});

// 余额状态颜色
const balanceColor = computed(() => {
  if (!stats.value) return '#2080f0';
  const days = parseFloat(stats.value.estimated_days_after_future);
  if (days <= 3) return '#d03050'; // 红色 - 危险
  if (days <= 7) return '#f0a020'; // 橙色 - 警告
  return '#18a058'; // 绿色 - 安全
});

// 余额状态文字
const balanceStatus = computed(() => {
  if (!stats.value) return '';
  const days = parseFloat(stats.value.estimated_days_after_future);
  if (days <= 3) return '余额紧张';
  if (days <= 7) return '余额充足';
  return '余额充裕';
});

onMounted(() => {
  loadStats();
  // 每2分钟自动刷新
  refreshTimer = setInterval(() => {
    loadStats();
  }, 120000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
});
</script>

<template>
  <NCard :bordered="false" class="jingdou-card">
    <NSpin :show="loading">
      <template v-if="stats">
        <div class="jingdou-container">
          <!-- 左侧：核心余额信息 -->
          <div class="balance-main">
            <div class="balance-header">
              <NIcon size="24" class="balance-icon">
                <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24">
                  <path fill="currentColor" d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm1.41 16.09V20h-2.67v-1.93c-1.71-.36-3.16-1.46-3.27-3.4h1.96c.1 1.05.82 1.87 2.65 1.87 1.96 0 2.4-.98 2.4-1.59 0-.83-.44-1.61-2.67-2.14-2.48-.6-4.18-1.62-4.18-3.67 0-1.72 1.39-2.84 3.11-3.21V4h2.67v1.95c1.86.45 2.79 1.86 2.85 3.39H14.3c-.05-1.11-.64-1.87-2.22-1.87-1.5 0-2.4.68-2.4 1.64 0 .84.65 1.39 2.67 1.91s4.18 1.39 4.18 3.91c-.01 1.83-1.38 2.83-3.12 3.16z"/>
                </svg>
              </NIcon>
              <span class="balance-title">京豆余额</span>
              <NTag :type="balanceColor === '#d03050' ? 'error' : balanceColor === '#f0a020' ? 'warning' : 'success'" size="small">
                {{ balanceStatus }}
              </NTag>
            </div>
            <div class="balance-value" :style="{ color: balanceColor }">
              {{ stats.jingdou_balance.toLocaleString() }}
            </div>
            <div class="balance-subtitle">
              可用余额: {{ stats.available_balance.toLocaleString() }} (扣除待执行任务)
            </div>
          </div>

          <!-- 右侧：统计信息 -->
          <div class="stats-grid">
            <!-- 预计可用天数 -->
            <div class="stat-item highlight">
              <div class="stat-value" :style="{ color: balanceColor }">
                {{ stats.estimated_days_after_future }}
              </div>
              <div class="stat-label">预计可用天数</div>
            </div>

            <!-- 日均消耗 -->
            <div class="stat-item">
              <div class="stat-value">{{ stats.daily_avg_consumed }}</div>
              <div class="stat-label">日均消耗</div>
            </div>

            <!-- 今日消耗 -->
            <div class="stat-item">
              <div class="stat-value">{{ stats.today_consumed.toLocaleString() }}</div>
              <div class="stat-label">今日消耗</div>
            </div>

            <!-- 待执行消耗 -->
            <div class="stat-item">
              <div class="stat-value">{{ stats.future_consumed.toLocaleString() }}</div>
              <div class="stat-label">待执行任务</div>
            </div>

            <!-- 近7天消耗 -->
            <div class="stat-item">
              <div class="stat-value">{{ stats.past_7_days_consumed.toLocaleString() }}</div>
              <div class="stat-label">近7天消耗</div>
            </div>

            <!-- 近30天消耗 -->
            <div class="stat-item">
              <div class="stat-value">{{ stats.past_30_days_consumed.toLocaleString() }}</div>
              <div class="stat-label">近30天消耗</div>
            </div>
          </div>
        </div>
      </template>
      <template v-else>
        <NEmpty description="暂无数据" />
      </template>
    </NSpin>
  </NCard>
</template>

<style scoped>
.jingdou-card {
  background: linear-gradient(135deg, rgba(32, 128, 240, 0.1) 0%, rgba(24, 160, 88, 0.1) 100%);
  border: 1px solid rgba(255, 255, 255, 0.1);
}

.jingdou-container {
  display: flex;
  align-items: center;
  gap: 40px;
}

.balance-main {
  flex-shrink: 0;
  min-width: 200px;
}

.balance-header {
  display: flex;
  align-items: center;
  gap: 8px;
  margin-bottom: 8px;
}

.balance-icon {
  color: #f0a020;
}

.balance-title {
  font-size: 14px;
  color: #999;
}

.balance-value {
  font-size: 42px;
  font-weight: 700;
  line-height: 1.2;
  font-family: 'Roboto', monospace;
}

.balance-subtitle {
  font-size: 12px;
  color: #999;
  margin-top: 4px;
}

.stats-grid {
  flex: 1;
  display: grid;
  grid-template-columns: repeat(6, 1fr);
  gap: 16px;
}

.stat-item {
  text-align: center;
  padding: 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
  border: 1px solid rgba(255, 255, 255, 0.05);
}

.stat-item.highlight {
  background: rgba(32, 128, 240, 0.1);
  border-color: rgba(32, 128, 240, 0.2);
}

.stat-value {
  font-size: 20px;
  font-weight: 600;
  color: #fff;
  margin-bottom: 4px;
}

.stat-label {
  font-size: 12px;
  color: #999;
}

/* 响应式 */
@media (max-width: 1200px) {
  .stats-grid {
    grid-template-columns: repeat(3, 1fr);
  }
}

@media (max-width: 768px) {
  .jingdou-container {
    flex-direction: column;
    gap: 20px;
  }

  .balance-main {
    text-align: center;
    width: 100%;
  }

  .balance-header {
    justify-content: center;
  }

  .stats-grid {
    grid-template-columns: repeat(2, 1fr);
    width: 100%;
  }
}
</style>
