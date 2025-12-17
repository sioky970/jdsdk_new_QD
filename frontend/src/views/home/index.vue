<script setup lang="ts">
import { computed } from 'vue';
import { useAuthStore } from '@/store/modules/auth';
// 管理员首页组件
import AdminHome from '@/views/admin-home/index.vue';
// 普通用户首页组件
import UserHome from './modules/user-home.vue';

defineOptions({
  name: 'Home'
});

const authStore = useAuthStore();

// 用户信息是否已加载
const isUserInfoLoaded = computed(() => Boolean(authStore.userInfo.id && authStore.userInfo.role));

// 是否为管理员
const isAdmin = computed(() => authStore.userInfo.role === 'admin');
</script>

<template>
  <div class="h-full">
    <!-- 等待用户信息加载完成 -->
    <template v-if="isUserInfoLoaded">
      <!-- 管理员首页 -->
      <AdminHome v-if="isAdmin" />
      
      <!-- 普通用户首页 -->
      <UserHome v-else />
    </template>
    
    <!-- 加载中占位 -->
    <div v-else class="h-full flex items-center justify-center">
      <NSpin size="large" />
    </div>
  </div>
</template>

<style scoped></style>
