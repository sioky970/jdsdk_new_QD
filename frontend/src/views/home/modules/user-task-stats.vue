<script setup lang="ts">
import { computed, onMounted, onUnmounted, ref } from 'vue';
import { fetchUserTodayStats } from '@/service/api';
import type { UserTodayStats } from '@/service/api';

const stats = ref<UserTodayStats | null>(null);
const loading = ref(true);
let refreshTimer: ReturnType<typeof setInterval> | null = null;

async function loadStats() {
  loading.value = true;
  const { data, error } = await fetchUserTodayStats();
  if (!error && data) {
    stats.value = data;
  }
  loading.value = false;
}

// 暴露方法供父组件调用
defineExpose({
  refresh: loadStats
});

const pieData = computed(() => {
  if (!stats.value) return [];
  const { today_completed, today_pending, today_failed, today_partial_completed } = stats.value;
  return [
    { name: '已完成', value: today_completed, color: '#18a058' },
    { name: '待执行', value: today_pending, color: '#2080f0' },
    { name: '部分完成', value: today_partial_completed, color: '#f0a020' },
    { name: '失败', value: today_failed, color: '#d03050' }
  ].filter(item => item.value > 0);
});

// 执行次数进度数据
const executeProgressData = computed(() => {
  if (!stats.value) return [];
  const { total_execute_count, total_executed_count } = stats.value;
  const pending = total_execute_count - total_executed_count;
  return [
    { name: '已执行', value: total_executed_count, color: '#18a058' },
    { name: '待执行', value: pending > 0 ? pending : 0, color: '#2080f0' }
  ].filter(item => item.value > 0);
});

onMounted(() => {
  loadStats();
  // 每1分钟自动刷新
  refreshTimer = setInterval(() => {
    loadStats();
  }, 60000);
});

onUnmounted(() => {
  // 清理定时器
  if (refreshTimer) {
    clearInterval(refreshTimer);
    refreshTimer = null;
  }
});
</script>

<template>
  <NCard title="今日任务统计" :bordered="false" class="card-wrapper">
    <NSpin :show="loading">
      <template v-if="stats">
        <NGrid :cols="4" :x-gap="12" :y-gap="12">
          <NGi>
            <NStatistic label="今日任务数" :value="stats.today_total" />
          </NGi>
          <NGi>
            <NStatistic label="已完成" :value="stats.today_completed">
              <template #suffix>
                <span class="text-success text-sm"> / {{ stats.today_total }}</span>
              </template>
            </NStatistic>
          </NGi>
          <NGi>
            <NStatistic label="总执行次数" :value="stats.total_execute_count" />
          </NGi>
          <NGi>
            <NStatistic label="已执行次数" :value="stats.total_executed_count">
              <template #suffix>
                <span class="text-success text-sm"> / {{ stats.total_execute_count }}</span>
              </template>
            </NStatistic>
          </NGi>
        </NGrid>

        <NDivider />

        <!-- 任务完成率和执行完成率 -->
        <div class="flex items-center justify-around">
          <div class="text-center">
            <NProgress
              type="circle"
              :percentage="Number(stats.completion_rate)"
              :stroke-width="10"
              style="width: 100px"
              :indicator-text-color="Number(stats.completion_rate) >= 80 ? '#18a058' : '#2080f0'"
            >
              <span class="text-lg font-bold">{{ stats.completion_rate }}%</span>
            </NProgress>
            <div class="text-sm text-gray-400 mt-2">任务完成率</div>
          </div>
          <div class="text-center">
            <NProgress
              type="circle"
              :percentage="Number(stats.execute_completion_rate)"
              :stroke-width="10"
              style="width: 100px"
              :indicator-text-color="Number(stats.execute_completion_rate) >= 80 ? '#18a058' : '#f0a020'"
            >
              <span class="text-lg font-bold">{{ stats.execute_completion_rate }}%</span>
            </NProgress>
            <div class="text-sm text-gray-400 mt-2">执行完成率</div>
          </div>
        </div>

        <NDivider />

        <!-- 任务状态分布 -->
        <div class="flex items-center justify-between mb-4">
          <span class="text-sm text-gray-400">任务状态分布</span>
        </div>
        <div class="flex flex-wrap gap-3">
          <div v-for="item in pieData" :key="item.name" class="flex items-center">
            <span
              class="inline-block w-3 h-3 rounded-full mr-2"
              :style="{ backgroundColor: item.color }"
            />
            <span class="text-sm">{{ item.name }}: {{ item.value }}</span>
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
.text-success {
  color: #18a058;
}
</style>
