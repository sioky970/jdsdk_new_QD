<script setup lang="ts">
import { onMounted, ref, computed, watch } from 'vue';
import { useEcharts } from '@/hooks/common/echarts';
import { fetchTaskPressure, type TaskPressure } from '@/service/api';

defineOptions({
  name: 'TaskPressure'
});

const props = defineProps<{
  statMode: 'count' | 'execute';
  taskType: string | null;
}>();

const pressure = ref<TaskPressure | null>(null);

// 动态计算单位
const unit = computed(() => (props.statMode === 'count' ? '个' : '次'));

const pressureType = computed(() => {
  if (!pressure.value) return 'default';
  const level = pressure.value.pressure_level;
  if (level === '低') return 'success';
  if (level === '中') return 'warning';
  if (level === '高' || level === '超载') return 'error';
  return 'info';
});

const { domRef, updateOptions } = useEcharts(() => ({
  tooltip: {
    trigger: 'axis',
    axisPointer: {
      type: 'shadow'
    }
  },
  grid: {
    left: '3%',
    right: '4%',
    bottom: '3%',
    containLabel: true
  },
  xAxis: {
    type: 'category',
    data: [] as string[],
    axisLabel: {
      rotate: 45
    }
  },
  yAxis: {
    type: 'value',
    name: props.statMode === 'count' ? '任务数' : '执行次数'
  },
  series: [
    {
      name: '待执行任务',
      type: 'bar',
      data: [] as number[],
      itemStyle: {
        color: '#1890ff'
      },
      label: {
        show: true,
        position: 'top'
      }
    }
  ]
}));

async function loadData() {
  const { data, error } = await fetchTaskPressure({
    stat_mode: props.statMode,
    task_type: props.taskType || undefined
  });
  if (!error && data) {
    pressure.value = data;
    updateOptions(opts => {
      opts.yAxis.name = props.statMode === 'count' ? '任务数' : '执行次数';
      // 确保数据存在
      if (data.future_tasks) {
        opts.xAxis.data = data.future_tasks.map(t => t.day?.substring(5) || ''); // MM-DD格式
        opts.series[0].data = data.future_tasks.map(t => t.count || 0);
      } else {
        opts.xAxis.data = [];
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
});
</script>

<template>
  <NCard title="任务执行压力" :bordered="false" class="card-wrapper">
    <template #header-extra>
      <NTag v-if="pressure" :type="pressureType">
        压力等级: {{ pressure?.pressure_level || '未知' }}
      </NTag>
    </template>
    
    <div class="mb-16px">
      <NAlert v-if="pressure" :type="pressureType" :show-icon="true">
        <template #header>执行能力分析</template>
        <div class="flex gap-24px">
          <span>昨日完成: <strong>{{ pressure?.yesterday_completed || 0 }}</strong> {{ unit }}</span>
          <span>3日平均: <strong>{{ (pressure?.avg_3days_completed || 0).toFixed(1) }}</strong> {{ unit }}/天</span>
          <span>待执行: <strong>{{ pressure?.total_future_pending || 0 }}</strong> {{ unit }}</span>
        </div>
      </NAlert>
    </div>

    <div class="text-14px mb-8px font-500">未来7天待执行{{ statMode === 'count' ? '任务' : '次数' }}分布</div>
    <div ref="domRef" class="h-250px"></div>
  </NCard>
</template>

<style scoped></style>