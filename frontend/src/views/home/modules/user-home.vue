<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { useAppStore } from '@/store/modules/app';
import { fetchLoginAnnouncement } from '@/service/api';
import JingdouBalance from './jingdou-balance.vue';
import UserTaskStats from './user-task-stats.vue';
import QuickTaskTemplates from './quick-task-templates.vue';

const appStore = useAppStore();
const gap = computed(() => (appStore.isMobile ? 0 : 16));

// 组件引用
const jingdouBalanceRef = ref<InstanceType<typeof JingdouBalance> | null>(null);
const taskStatsRef = ref<InstanceType<typeof UserTaskStats> | null>(null);

// 刷新所有统计
function refreshStats() {
  taskStatsRef.value?.refresh();
  jingdouBalanceRef.value?.refresh();
}

// 公告内容
const announcement = ref('');

// 获取公告
async function loadAnnouncement() {
  const { data, error } = await fetchLoginAnnouncement();
  if (!error && data?.announcement) {
    announcement.value = data.announcement;
  }
}

onMounted(() => {
  loadAnnouncement();
});
</script>

<template>
  <NSpace vertical :size="16">
    <!-- 京豆余额统计 - 最重要，放在最上部 -->
    <JingdouBalance ref="jingdouBalanceRef" />

    <!-- 系统公告 -->
    <NAlert v-if="announcement" title="系统公告" type="info" closable>
      {{ announcement }}
    </NAlert>

    <!-- 任务统计与快速下发 -->
    <NGrid :x-gap="gap" :y-gap="16" responsive="screen" item-responsive>
      <NGi span="24 s:24 m:10">
        <UserTaskStats ref="taskStatsRef" />
      </NGi>
      <NGi span="24 s:24 m:14">
        <QuickTaskTemplates @task-created="refreshStats" />
      </NGi>
    </NGrid>
  </NSpace>
</template>

<style scoped></style>
