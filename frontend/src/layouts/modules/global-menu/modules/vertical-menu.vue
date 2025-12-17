<script setup lang="ts">
import { computed, ref, watch } from 'vue';
import { useRoute } from 'vue-router';
import { SimpleScrollbar } from '@sa/materials';
import { GLOBAL_SIDER_MENU_ID } from '@/constants/app';
import { useAppStore } from '@/store/modules/app';
import { useThemeStore } from '@/store/modules/theme';
import { useRouteStore } from '@/store/modules/route';
import { useRouterPush } from '@/hooks/common/router';
import { useMenu } from '../context';

defineOptions({
  name: 'VerticalMenu'
});

const route = useRoute();
const appStore = useAppStore();
const themeStore = useThemeStore();
const routeStore = useRouteStore();
const { routerPushByKeyWithMetaQuery } = useRouterPush();
const { selectedKey } = useMenu();

// 修复：NaiveUI NMenu 在 value 变化时会触发 @update:value 事件并传入旧值（反弹事件）
// 策略：记录最后一次导航的目标，如果新的导航目标不同且在短时间内，则忽略
const lastNavigation = ref({ key: '', time: 0 });

function handleMenuSelect(key: string) {
  const now = Date.now();
  const timeSinceLastNav = now - lastNavigation.value.time;
  
  console.log('[Menu] handleMenuSelect:', key, 'selectedKey:', selectedKey.value, 'timeSince:', timeSinceLastNav);
  
  // 如果点击的是当前已选中的菜单项，忽略
  if (key === selectedKey.value) {
    console.log('[Menu] Ignored: same as selectedKey');
    return;
  }
  
  // 如果距离上次导航不到 500ms，且新的 key 不是上次导航的目标，忽略（这是反弹事件）
  if (timeSinceLastNav < 500 && key !== lastNavigation.value.key) {
    console.log('[Menu] Ignored: bounce event detected');
    return;
  }
  
  console.log('[Menu] Navigating to:', key);
  lastNavigation.value = { key, time: now };
  routerPushByKeyWithMetaQuery(key as any);
}

const inverted = computed(() => !themeStore.darkMode && themeStore.sider.inverted);

const expandedKeys = ref<string[]>([]);

function updateExpandedKeys() {
  if (appStore.siderCollapse || !selectedKey.value) {
    expandedKeys.value = [];
    return;
  }
  expandedKeys.value = routeStore.getSelectedMenuKeyPath(selectedKey.value);
}

watch(
  () => route.name,
  () => {
    updateExpandedKeys();
  },
  { immediate: true }
);
</script>

<template>
  <Teleport :to="`#${GLOBAL_SIDER_MENU_ID}`">
    <SimpleScrollbar>
      <NMenu
        v-model:expanded-keys="expandedKeys"
        mode="vertical"
        :value="selectedKey"
        :collapsed="appStore.siderCollapse"
        :collapsed-width="themeStore.sider.collapsedWidth"
        :collapsed-icon-size="22"
        :options="routeStore.menus"
        :inverted="inverted"
        :indent="18"
        @update:value="handleMenuSelect"
      />
    </SimpleScrollbar>
  </Teleport>
</template>

<style scoped></style>
