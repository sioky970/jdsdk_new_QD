<script setup lang="ts">
import { onMounted, ref } from 'vue';
import { fetchFinanceStats, type FinanceStats } from '@/service/api';

defineOptions({
  name: 'FinanceStats'
});

const stats = ref<FinanceStats | null>(null);

async function loadData() {
  const { data, error } = await fetchFinanceStats();
  if (!error && data) {
    stats.value = data;
  }
}

onMounted(() => {
  loadData();
});
</script>

<template>
  <NCard title="财务统计" :bordered="false" class="card-wrapper">
    <template v-if="stats">
      <!-- 统计卡片 -->
      <div class="grid grid-cols-2 gap-16px mb-24px">
        <NCard size="small" :bordered="true">
          <NStatistic label="日均充值京豆(30天)" :value="stats.avg_daily_recharge.toFixed(0)">
            <template #prefix>
              <span class="text-success">+</span>
            </template>
          </NStatistic>
        </NCard>
        <NCard size="small" :bordered="true">
          <NStatistic label="日均消耗京豆(30天)" :value="stats.avg_daily_consume.toFixed(0)">
            <template #prefix>
              <span class="text-error">-</span>
            </template>
          </NStatistic>
        </NCard>
        <NCard size="small" :bordered="true">
          <NStatistic label="今日充值" :value="stats.today_recharge">
            <template #prefix>
              <span class="text-success">+</span>
            </template>
          </NStatistic>
        </NCard>
        <NCard size="small" :bordered="true">
          <NStatistic label="今日消耗" :value="stats.today_consume">
            <template #prefix>
              <span class="text-error">-</span>
            </template>
          </NStatistic>
        </NCard>
      </div>

      <!-- 低余额用户列表 -->
      <div class="text-14px mb-8px font-500">京豆余额最低用户 (Top 10)</div>
      <NDataTable
        :columns="[
          { title: '排名', key: 'rank', width: 60 },
          { title: '用户名', key: 'username' },
          { title: '昵称', key: 'nickname' },
          { title: '京豆余额', key: 'jingdou_balance', align: 'right' }
        ]"
        :data="(stats.low_balance_users || []).map((u, i) => ({ ...u, rank: i + 1 }))"
        :bordered="false"
        size="small"
        :pagination="false"
      />
    </template>
    <NSpin v-else size="large" class="w-full h-200px flex-center" />
  </NCard>
</template>

<style scoped>
.text-success {
  color: #52c41a;
}
.text-error {
  color: #f5222d;
}
</style>