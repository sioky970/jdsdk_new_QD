<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { fetchLoginAnnouncement, fetchActiveTaskTypes } from '@/service/api';
import TodayTaskPie from './modules/today-task-pie.vue';
import TaskPressure from './modules/task-pressure.vue';
import FinanceStats from './modules/finance-stats.vue';

defineOptions({
  name: 'AdminHome'
});

const appStore = useAppStore();
const gap = computed(() => (appStore.isMobile ? 0 : 16));

// 公告内容
const announcement = ref('');

// 统计筛选参数
const statMode = ref<'count' | 'execute'>('execute');
const taskType = ref<string | null>(null);

// 统计模式选项
const statModeOptions = [
  { label: '执行次数', value: 'execute' },
  { label: '任务数量', value: 'count' }
];

// 任务类型选项
const taskTypeOptions = ref<Array<{ label: string; value: string }>>([]);

// 获取公告
async function loadAnnouncement() {
  const { data, error } = await fetchLoginAnnouncement();
  if (!error && data?.announcement) {
    announcement.value = data.announcement;
  }
}

// 加载任务类型列表
async function loadTaskTypes() {
  const { data, error } = await fetchActiveTaskTypes();
  if (!error && data?.task_types) {
    taskTypeOptions.value = data.task_types.map(t => ({
      label: t.type_name,
      value: t.type_code
    }));
  }
}

onMounted(() => {
  loadAnnouncement();
  loadTaskTypes();
});
</script>

<template>
  <NSpace vertical :size="16">
    <!-- 系统公告 -->
    <NAlert v-if="announcement" title="系统公告" type="info">
      {{ announcement }}
    </NAlert>

    <!-- 管理员专属头部 -->
    <NCard :bordered="false" class="card-wrapper">
      <div class="flex items-center gap-16px">
        <NIcon size="48" color="#1890ff">
          <svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor">
            <path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/>
          </svg>
        </NIcon>
        <div>
          <h2 class="text-20px font-600 m-0">管理员控制台</h2>
          <p class="text-14px text-gray-500 m-0 mt-4px">实时监控系统运行状态，掌握任务执行与财务情况</p>
        </div>
      </div>
    </NCard>

    <!-- 统计筛选控制 -->
    <NCard :bordered="false" class="card-wrapper">
      <NSpace align="center" :size="24">
        <div class="flex items-center gap-8px">
          <span class="text-14px text-gray-500">统计维度:</span>
          <NRadioGroup v-model:value="statMode" size="small">
            <NRadioButton v-for="opt in statModeOptions" :key="opt.value" :value="opt.value">
              {{ opt.label }}
            </NRadioButton>
          </NRadioGroup>
        </div>
        <div class="flex items-center gap-8px">
          <span class="text-14px text-gray-500">任务类型:</span>
          <NSelect
            v-model:value="taskType"
            :options="taskTypeOptions"
            placeholder="全部类型"
            clearable
            style="width: 180px"
          />
        </div>
      </NSpace>
    </NCard>

    <!-- 今日任务统计 + 任务压力 -->
    <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
      <NGi span="24 s:24 m:12">
        <TodayTaskPie :stat-mode="statMode" :task-type="taskType" />
      </NGi>
      <NGi span="24 s:24 m:12">
        <TaskPressure :stat-mode="statMode" :task-type="taskType" />
      </NGi>
    </NGrid>

    <!-- 财务统计 -->
    <FinanceStats />
  </NSpace>
</template>

<style scoped></style>
