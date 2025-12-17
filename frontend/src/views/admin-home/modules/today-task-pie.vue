<script setup lang="ts">
import { onMounted, onUnmounted, ref, watch, computed } from 'vue';
import { useEcharts } from '@/hooks/common/echarts';
import { fetchTodayTaskStats, type TodayTaskStats } from '@/service/api';

defineOptions({
  name: 'TodayTaskPie'
});

const props = defineProps<{
  statMode: 'count' | 'execute';
  taskType: string | null;
}>();

const stats = ref<TodayTaskStats | null>(null);
let refreshTimer: ReturnType<typeof setInterval> | null = null;

// 动态计算单位和标题
const unit = computed(() => (props.statMode === 'count' ? '个' : '次'));
const chartTitle = computed(() =>
  props.statMode === 'count' ? '今日任务数量统计' : '今日任务执行次数统计'
);

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: {
    trigger: 'item',
    formatter: `{b}: {c}${unit.value} ({d}%)`
  },
  legend: {
    bottom: '5%',
    left: 'center'
  },
  series: [
    {
      name: props.statMode === 'count' ? '任务数量' : '任务执行次数',
      type: 'pie',
      radius: ['40%', '70%'],
      avoidLabelOverlap: false,
      itemStyle: {
        borderRadius: 10,
        borderColor: '#fff',
        borderWidth: 2
      },
      label: {
        show: true,
        position: 'outside',
        formatter: `{b}\n{c}${unit.value}`
      },
      emphasis: {
        label: {
          show: true,
          fontSize: 14,
          fontWeight: 'bold'
        }
      },
      data: [] as { name: string; value: number; itemStyle: { color: string } }[]
    }
  ]
}));

async function loadData() {
  const { data, error } = await fetchTodayTaskStats({
    stat_mode: props.statMode,
    task_type: props.taskType || undefined
  });
  if (!error && data) {
    stats.value = data;
    const currentUnit = props.statMode === 'count' ? '个' : '次';
    updateOptions(opts => {
      opts.tooltip.formatter = `{b}: {c}${currentUnit} ({d}%)`;
      opts.series[0].name = props.statMode === 'count' ? '任务数量' : '任务执行次数';
      opts.series[0].label.formatter = `{b}\n{c}${currentUnit}`;
      // 确保数据存在且有效
      if (data.pending_value !== undefined && data.running_value !== undefined && data.completed_value !== undefined) {
        opts.series[0].data = [
          { name: '待执行', value: data.pending_value, itemStyle: { color: '#faad14' } },
          { name: '执行中', value: data.running_value, itemStyle: { color: '#1890ff' } },
          { name: '已完成', value: data.completed_value, itemStyle: { color: '#52c41a' } }
        ];
      } else {
        // 如果数据无效，则设置为空数组
        opts.series[0].data = [];
      }
      return opts;
    });
  }
}

// 监听参数变化
watch(
  () => [props.statMode, props.taskType],
  () => {
    loadData();
  }
);

onMounted(() => {
  loadData();
  // 10秒刷新一次
  refreshTimer = setInterval(loadData, 10000);
});

onUnmounted(() => {
  if (refreshTimer) {
    clearInterval(refreshTimer);
  }
});
</script>

<template>
  <NCard :title="chartTitle" :bordered="false" class="card-wrapper">
    <template #header-extra>
      <NSpace :size="8">
        <NTag v-if="stats" type="info">{{ statMode === 'count' ? '总任务' : '总次数' }}: {{ stats?.total_value?.toLocaleString() || 0 }} {{ unit }}</NTag>
        <NTag v-if="stats && statMode === 'execute'" type="default">任务数: {{ stats?.total_tasks || 0 }} 个</NTag>
      </NSpace>
    </template>
    <div ref="domRef" class="h-300px"></div>
    <div v-if="stats" class="flex justify-around mt-16px">
      <NStatistic label="待执行" :value="stats?.pending_value || 0">
        <template #suffix>
          <span class="text-12px text-gray-500">{{ unit }} ({{ (stats?.pending_percent || 0).toFixed(1) }}%)</span>
        </template>
      </NStatistic>
      <NStatistic label="执行中" :value="stats?.running_value || 0">
        <template #suffix>
          <span class="text-12px text-gray-500">{{ unit }} ({{ (stats?.running_percent || 0).toFixed(1) }}%)</span>
        </template>
      </NStatistic>
      <NStatistic label="已完成" :value="stats?.completed_value || 0">
        <template #suffix>
          <span class="text-12px text-gray-500">{{ unit }} ({{ (stats?.completed_percent || 0).toFixed(1) }}%)</span>
        </template>
      </NStatistic>
    </div>
  </NCard>
</template>

<style scoped></style>