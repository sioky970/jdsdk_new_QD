<script setup lang="ts">
import { computed, onMounted, ref } from 'vue';
import { NInput, NIcon, NSpace, NTimePicker, useMessage, useDialog } from 'naive-ui';
import { fetchTaskTemplates, fetchTemplatePrice, quickCreateTask, fetchActiveTaskTypes, updateTemplateRemark } from '@/service/api';
import type { TaskTemplate, TemplatePriceResponse, TaskTypeItem } from '@/service/api';
import { useAuthStore } from '@/store/modules/auth';

const emit = defineEmits<{
  (e: 'task-created'): void;
}>();

const message = useMessage();
const dialog = useDialog();
const authStore = useAuthStore();
const templates = ref<TaskTemplate[]>([]);
const loading = ref(true);
const showModal = ref(false);
const selectedTemplate = ref<TaskTemplate | null>(null);
const isCreatingNew = ref(false); // 标记是否是创建新任务模式
const editingRemarkId = ref<number | null>(null); // 正在编辑备注的模板ID
const editRemarkText = ref(''); // 编辑中的备注文本
const page = ref(1); // 当前页码
const limit = ref(9); // 每次加载9个模板
const hasMoreTemplates = ref(true); // 是否还有更多数据
const loadingMore = ref(false); // 是否正在加载更多
const formData = ref({
  executeCount: 1,
  selectedDates: [] as number[], // 多选日期
  taskType: '', // 选中的任务类型
  keyword: '', // 搜索关键词
  shopName: '', // 店铺名称
  sku: '' // SKU编号
});
const priceInfo = ref<TemplatePriceResponse | null>(null);
const priceLoading = ref(false);
const submitLoading = ref(false);
const taskTypes = ref<TaskTypeItem[]>([]);
const showCustomTimePicker = ref(false); // 显示自定义时间选择器
const customTime = ref<number | null>(null); // 自定义时间

// 计算总消耗（每个日期创建一个任务）
const totalConsume = computed(() => {
  if (!priceInfo.value) return 0;
  return priceInfo.value.consume_jingdou * formData.value.selectedDates.length;
});

const isSufficient = computed(() => {
  if (!priceInfo.value) return true;
  return priceInfo.value.jingdou_balance >= totalConsume.value;
});

async function loadTemplates(append = false) {
  if (append) {
    loadingMore.value = true;
  } else {
    loading.value = true;
    page.value = 1;
  }
  
  const { data, error } = await fetchTaskTemplates(page.value * limit.value);
  if (!error && data) {
    const newTemplates = data.templates || [];
    if (append) {
      templates.value = [...templates.value, ...newTemplates.slice(templates.value.length)];
    } else {
      templates.value = newTemplates;
    }
    // 判断是否还有更多数据
    hasMoreTemplates.value = newTemplates.length >= page.value * limit.value;
  }
  
  if (append) {
    loadingMore.value = false;
  } else {
    loading.value = false;
  }
}

// 加载更多模板
function loadMoreTemplates() {
  if (!loadingMore.value && hasMoreTemplates.value) {
    page.value++;
    loadTemplates(true);
  }
}

// 处理滚动事件
function handleTemplateScroll(e: Event) {
  const target = e.target as HTMLElement;
  const scrollTop = target.scrollTop;
  const scrollHeight = target.scrollHeight;
  const clientHeight = target.clientHeight;
  
  // 当滚动到距离底部50px时触发加载更多
  if (scrollHeight - scrollTop - clientHeight < 50) {
    loadMoreTemplates();
  }
}

// 加载已启用的任务类型
async function loadTaskTypes() {
  const { data, error } = await fetchActiveTaskTypes();
  if (!error && data) {
    // 只显示已启用的任务类型
    taskTypes.value = (data.task_types || []).filter(t => t.is_active);
  }
}

// 是否需要显示关键词和店铺名输入框
const isSearchBrowse = computed(() => formData.value.taskType === 'search_browse');

// 关键词是否必填（模板没有则必填）
const isKeywordRequired = computed(() => isSearchBrowse.value && !selectedTemplate.value?.keyword);
const isShopNameRequired = computed(() => isSearchBrowse.value && !selectedTemplate.value?.shop_name);

// 表单验证
const isFormValid = computed(() => {
  if (!isSearchBrowse.value) return true;
  // 关键词：模板有或者用户填写了
  const hasKeyword = selectedTemplate.value?.keyword || formData.value.keyword.trim();
  // 店铺名：模板有或者用户填写了
  const hasShopName = selectedTemplate.value?.shop_name || formData.value.shopName.trim();
  return hasKeyword && hasShopName;
});

// 任务类型选项计算属性
const taskTypeOptions = computed(() => {
  return taskTypes.value.map(t => ({
    label: `${t.type_name} (${t.jingdou_price}京豆/次)`,
    value: t.type_code
  }));
});

function openQuickCreate(template: TaskTemplate) {
  isCreatingNew.value = false;
  selectedTemplate.value = template;
  formData.value = {
    executeCount: 1,
    selectedDates: [],
    taskType: template.task_type, // 默认使用模板的任务类型
    keyword: template.keyword || '', // 自动填充模板的关键词
    shopName: template.shop_name || '', // 自动填充模板的店铺名
    sku: template.sku
  };
  priceInfo.value = null;
  showModal.value = true;
  calculatePrice();
}

// 创建新任务
function createNewTask() {
  isCreatingNew.value = true;
  selectedTemplate.value = null;
  formData.value = {
    executeCount: 1,
    selectedDates: [],
    taskType: taskTypes.value.length > 0 ? taskTypes.value[0].type_code : '',
    keyword: '',
    shopName: '',
    sku: ''
  };
  priceInfo.value = null;
  showModal.value = true;
  if (formData.value.taskType) {
    calculatePriceForNew();
  }
}

// 日期范围选择后批量添加（不设置时间，等待用户为每个日期设置时间）
function onDateRangeConfirm(value: [number, number] | null) {
  if (!value || value.length !== 2) return;
  
  const [startTs, endTs] = value;
  const startDate = new Date(startTs);
  const endDate = new Date(endTs);
  
  // 不设置默认时间，只设置日期（时间为 00:00:00）
  const dates: number[] = [];
  const current = new Date(startDate);
  current.setHours(0, 0, 0, 0);
  
  while (current <= endDate) {
    const dateTs = current.getTime();
    // 检查是否已存在
    const exists = formData.value.selectedDates.some(d => {
      const d1 = new Date(d);
      return d1.getFullYear() === current.getFullYear() &&
             d1.getMonth() === current.getMonth() &&
             d1.getDate() === current.getDate();
    });
    
    if (!exists) {
      dates.push(dateTs);
    }
    
    current.setDate(current.getDate() + 1);
  }
  
  if (dates.length > 0) {
    formData.value.selectedDates.push(...dates);
    formData.value.selectedDates.sort((a, b) => a - b);
    message.success(`已添加 ${dates.length} 个日期，请为每个日期设置具体时间`);
  } else {
    message.warning('选择的日期已存在');
  }
}

// 修改某个日期的时间
function updateDateTime(index: number, newTime: number | null) {
  if (newTime === null) return;
  
  const oldDate = new Date(formData.value.selectedDates[index]);
  const newDate = new Date(newTime);
  
  // 保持原日期，只更新时间
  oldDate.setHours(newDate.getHours(), newDate.getMinutes(), 0, 0);
  
  formData.value.selectedDates[index] = oldDate.getTime();
  formData.value.selectedDates.sort((a, b) => a - b);
}

// 快捷设置时间（批量设置所有日期）
function setQuickTime(hour: number, minute: number = 0) {
  formData.value.selectedDates = formData.value.selectedDates.map(dateTs => {
    const date = new Date(dateTs);
    date.setHours(hour, minute, 0, 0);
    return date.getTime();
  });
  formData.value.selectedDates.sort((a, b) => a - b);
  message.success(`已将所有日期时间设置为 ${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`);
}

// 打开自定义时间选择器
function openCustomTimePicker() {
  // 初始化为当前时间
  const now = new Date();
  now.setSeconds(0, 0);
  customTime.value = now.getTime();
  showCustomTimePicker.value = true;
}

// 确认自定义时间
function confirmCustomTime() {
  if (!customTime.value) return;
  
  const time = new Date(customTime.value);
  const hour = time.getHours();
  const minute = time.getMinutes();
  
  formData.value.selectedDates = formData.value.selectedDates.map(dateTs => {
    const date = new Date(dateTs);
    date.setHours(hour, minute, 0, 0);
    return date.getTime();
  });
  formData.value.selectedDates.sort((a, b) => a - b);
  
  message.success(`已将所有日期时间设置为 ${hour.toString().padStart(2, '0')}:${minute.toString().padStart(2, '0')}`);
  showCustomTimePicker.value = false;
}

// 删除日期
function removeDate(index: number) {
  formData.value.selectedDates.splice(index, 1);
}

// 获取日期相对描述和样式
function getDateLabel(ts: number) {
  const now = new Date();
  const target = new Date(ts);
  
  // 计算天数差
  const todayStart = new Date(now.getFullYear(), now.getMonth(), now.getDate()).getTime();
  const targetStart = new Date(target.getFullYear(), target.getMonth(), target.getDate()).getTime();
  const dayDiff = Math.round((targetStart - todayStart) / (1000 * 60 * 60 * 24));
  
  const timeStr = target.toLocaleTimeString('zh-CN', { hour: '2-digit', minute: '2-digit' });
  const dateStr = target.toLocaleDateString('zh-CN', { month: 'numeric', day: 'numeric' });
  
  if (dayDiff === 0) {
    return { text: `今日 ${timeStr}`, type: 'success' as const, color: '#18a058' };
  } else if (dayDiff === 1) {
    return { text: `明日 ${timeStr}`, type: 'info' as const, color: '#2080f0' };
  } else if (dayDiff === 2) {
    return { text: `后天 ${timeStr}`, type: 'warning' as const, color: '#f0a020' };
  } else if (dayDiff > 2 && dayDiff <= 7) {
    return { text: `${dayDiff}天后 ${dateStr} ${timeStr}`, type: 'warning' as const, color: '#f0a020' };
  } else if (dayDiff > 7) {
    return { text: `${dayDiff}天后 ${dateStr} ${timeStr}`, type: 'default' as const, color: '#909399' };
  } else {
    // 过去的日期
    return { text: `${dateStr} ${timeStr}`, type: 'error' as const, color: '#d03050' };
  }
}

async function calculatePrice() {
  if (!selectedTemplate.value) return;
  priceLoading.value = true;
  const { data, error } = await fetchTemplatePrice(
    selectedTemplate.value.id,
    formData.value.executeCount,
    formData.value.taskType // 传入选中的任务类型
  );
  if (!error && data) {
    priceInfo.value = data;
  }
  priceLoading.value = false;
}

// 新任务模式的价格计算
async function calculatePriceForNew() {
  if (!formData.value.taskType) return;
  priceLoading.value = true;
  // 直接根据任务类型计算价格
  const taskType = taskTypes.value.find(t => t.type_code === formData.value.taskType);
  if (taskType) {
    // 使用用户当前余额
    priceInfo.value = {
      jingdou_price: taskType.jingdou_price,
      consume_jingdou: taskType.jingdou_price * formData.value.executeCount,
      jingdou_balance: authStore.userInfo.jingdou_balance,
      is_sufficient: authStore.userInfo.jingdou_balance >= taskType.jingdou_price * formData.value.executeCount
    };
  }
  priceLoading.value = false;
}

// 任务类型变化时重新计算价格
function onTaskTypeChange() {
  // 如果切换到关键词搜索任务，自动填充模板的关键词和店铺名
  if (formData.value.taskType === 'search_browse') {
    if (!formData.value.keyword && selectedTemplate.value?.keyword) {
      formData.value.keyword = selectedTemplate.value.keyword;
    }
    if (!formData.value.shopName && selectedTemplate.value?.shop_name) {
      formData.value.shopName = selectedTemplate.value.shop_name;
    }
  }
  // 根据是否是新任务模式调用不同的计算方法
  if (isCreatingNew.value) {
    calculatePriceForNew();
  } else {
    calculatePrice();
  }
}

async function handleSubmit() {
  if (formData.value.selectedDates.length === 0) {
    message.error('请至少选择一个日期');
    return;
  }

  // 验证所有日期是否都设置了时间（不是 00:00）
  const hasUnsetTime = formData.value.selectedDates.some(dateTs => {
    const date = new Date(dateTs);
    return date.getHours() === 0 && date.getMinutes() === 0;
  });
  
  if (hasUnsetTime) {
    message.error('请为所有日期设置具体时间');
    return;
  }

  // 新任务模式需要验证SKU
  if (isCreatingNew.value && !formData.value.sku.trim()) {
    message.error('请输入SKU编号');
    return;
  }

  // 关键词搜索浏览任务验证
  if (formData.value.taskType === 'search_browse') {
    const finalKeyword = formData.value.keyword.trim() || selectedTemplate.value?.keyword;
    const finalShopName = formData.value.shopName.trim() || selectedTemplate.value?.shop_name;
    if (!finalKeyword) {
      message.error('请填写搜索关键词');
      return;
    }
    if (!finalShopName) {
      message.error('请填写店铺名称');
      return;
    }
  }

  if (!isSufficient.value) {
    message.error('京豆余额不足');
    return;
  }

  submitLoading.value = true;
  let successCount = 0;
  let totalConsumed = 0;

  // 为每个选中的日期创建一个任务
  for (const dateTs of formData.value.selectedDates) {
    const requestData: any = {
      execute_count: formData.value.executeCount,
      start_time: new Date(dateTs).toISOString(),
      task_type: formData.value.taskType
    };
    
    // 新任务模式直接传递SKU
    if (isCreatingNew.value) {
      requestData.sku = formData.value.sku.trim();
    } else {
      // 模板模式传递模板ID
      requestData.template_id = selectedTemplate.value.id;
    }
    
    // 关键词搜索浏览任务需要传递关键词和店铺名
    if (formData.value.taskType === 'search_browse') {
      if (formData.value.keyword.trim()) {
        requestData.keyword = formData.value.keyword.trim();
      }
      if (formData.value.shopName.trim()) {
        requestData.shop_name = formData.value.shopName.trim();
      }
    }
    
    const { data, error } = await quickCreateTask(requestData);

    if (!error && data) {
      successCount++;
      totalConsumed += data.consume_jingdou;
    }
  }

  if (successCount > 0) {
    message.success(`成功创建 ${successCount} 个任务，共消耗 ${totalConsumed} 京豆`);
    showModal.value = false;
    loadTemplates();
    // 通知父组件刷新统计数据
    emit('task-created');
  } else {
    message.error('创建失败');
  }
  submitLoading.value = false;
}

function formatDate(dateStr: string) {
  if (!dateStr) return '-';
  const date = new Date(dateStr);
  return date.toLocaleDateString('zh-CN', { month: 'short', day: 'numeric', hour: '2-digit', minute: '2-digit' });
}

// 开始编辑备注
function startEditRemark(template: TaskTemplate, event: Event) {
  event.stopPropagation(); // 阻止事件冒泡，避免触发模板卡片的点击事件
  editingRemarkId.value = template.id;
  editRemarkText.value = template.remark || '';
}

// 取消编辑备注
function cancelEditRemark() {
  editingRemarkId.value = null;
  editRemarkText.value = '';
}

// 保存备注
async function saveRemark(templateId: number) {
  const { error } = await updateTemplateRemark(templateId, editRemarkText.value);
  if (!error) {
    message.success('备注更新成功');
    // 更新本地数据
    const template = templates.value.find(t => t.id === templateId);
    if (template) {
      template.remark = editRemarkText.value;
    }
    cancelEditRemark();
  }
}

onMounted(() => {
  loadTemplates();
  loadTaskTypes();
});
</script>

<template>
  <NCard title="快速下发任务" :bordered="false" class="card-wrapper">
    <template #header-extra>
      <NSpace>
        <NButton type="primary" @click="createNewTask">
          <template #icon>
            <NIcon><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M19 13h-6v6h-2v-6H5v-2h6V5h2v6h6v2z"/></svg></NIcon>
          </template>
          创建新任务
        </NButton>
        <NButton text type="primary" @click="loadTemplates">
          刷新
        </NButton>
      </NSpace>
    </template>

    <NSpin :show="loading">
      <div v-if="templates.length === 0" class="py-8">
        <NEmpty description="暂无任务模板，创建任务后将自动保存模板" />
      </div>

      <div v-else class="template-container" @scroll="handleTemplateScroll">
        <div class="template-list">
          <div
            v-for="template in templates"
            :key="template.id"
            class="template-card"
            @click="openQuickCreate(template)"
          >
            <div class="template-header">
              <NTag :type="template.task_type_name ? 'info' : 'default'" size="medium">
                {{ template.task_type_name || template.task_type }}
              </NTag>
              <span class="template-price">{{ template.jingdou_price }}<span class="price-unit">京豆/次</span></span>
            </div>
            <div class="template-sku">
              <span class="sku-label">SKU</span>
              <span class="sku-value">{{ template.sku }}</span>
            </div>
            <div class="template-info shop-info">
              <span class="info-label">店铺</span>
              <span class="info-value shop-value">{{ template.shop_name || '-' }}</span>
            </div>
            <div class="template-info keyword-info">
              <span class="info-label">关键词</span>
              <span class="info-value keyword-value">{{ template.keyword || '-' }}</span>
            </div>
            <!-- 备注区域 -->
            <div class="template-remark">
              <div v-if="editingRemarkId === template.id" class="remark-edit" @click.stop>
                <NInput
                  v-model:value="editRemarkText"
                  type="textarea"
                  :rows="2"
                  :maxlength="100"
                  placeholder="请输入备注（最多100字）"
                  show-count
                  size="small"
                />
                <div class="remark-actions">
                  <NButton size="tiny" @click="cancelEditRemark">取消</NButton>
                  <NButton type="primary" size="tiny" @click="saveRemark(template.id)">保存</NButton>
                </div>
              </div>
              <div v-else class="remark-display">
                <span class="remark-label">备注:</span>
                <span class="remark-text" :title="template.remark || '暂无备注'">{{ template.remark || '暂无备注' }}</span>
                <NButton
                  text
                  type="primary"
                  size="tiny"
                  @click="(e) => startEditRemark(template, e)"
                >
                  <template #icon>
                    <NIcon><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M3 17.25V21h3.75L17.81 9.94l-3.75-3.75L3 17.25zM20.71 7.04c.39-.39.39-1.02 0-1.41l-2.34-2.34c-.39-.39-1.02-.39-1.41 0l-1.83 1.83 3.75 3.75 1.83-1.83z"/></svg></NIcon>
                  </template>
                  修改
                </NButton>
              </div>
            </div>
            <div class="template-footer">
              <span class="usage-count">
                <NIcon size="14" class="mr-1"><svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2C6.48 2 2 6.48 2 12s4.48 10 10 10 10-4.48 10-10S17.52 2 12 2zm-2 15l-5-5 1.41-1.41L10 14.17l7.59-7.59L19 8l-9 9z"/></svg></NIcon>
                {{ template.total_created_count }} 次
              </span>
              <span class="last-used">
                {{ formatDate(template.last_used_at) }}
              </span>
            </div>
          </div>
        </div>
        <!-- 加载更多提示 -->
        <div v-if="loadingMore" class="loading-more">
          <NSpin size="small" />
          <span class="ml-2">加载中...</span>
        </div>
        <div v-else-if="!hasMoreTemplates && templates.length > 0" class="no-more">
          已加载全部 {{ templates.length }} 个模板
        </div>
      </div>
    </NSpin>

    <!-- 快速创建弹窗 -->
    <NModal
      v-model:show="showModal"
      preset="card"
      :title="isCreatingNew ? '创建新任务' : '快速创建任务'"
      style="width: 450px"
      :mask-closable="false"
    >
      <template v-if="isCreatingNew || selectedTemplate">
        <!-- 新任务模式：显示SKU输入框 -->
        <template v-if="isCreatingNew">
          <NForm label-placement="left" label-width="80px">
            <NFormItem label="SKU编号" required>
              <NInput
                v-model:value="formData.sku"
                placeholder="请输入SKU编号"
                clearable
              />
            </NFormItem>
          </NForm>
        </template>
        <!-- 模板模式：显示模板信息 -->
        <NDescriptions v-else :column="1" label-placement="left" bordered>
          <NDescriptionsItem label="SKU">
            {{ selectedTemplate.sku }}
          </NDescriptionsItem>
          <NDescriptionsItem v-if="selectedTemplate.shop_name && !isSearchBrowse" label="店铺">
            {{ selectedTemplate.shop_name }}
          </NDescriptionsItem>
          <NDescriptionsItem v-if="selectedTemplate.task_type === 'search_browse' && selectedTemplate.keyword && !isSearchBrowse" label="关键词">
            {{ selectedTemplate.keyword }}
          </NDescriptionsItem>
        </NDescriptions>

        <NDivider />

        <NForm label-placement="left" label-width="80px">
          <NFormItem label="任务类型">
            <NSelect
              v-model:value="formData.taskType"
              :options="taskTypeOptions"
              placeholder="选择任务类型"
              style="width: 100%"
              @update:value="onTaskTypeChange"
            />
          </NFormItem>
          <!-- 关键词搜索浏览任务的额外字段 -->
          <template v-if="isSearchBrowse">
            <NFormItem label="搜索关键词" :required="isKeywordRequired">
              <NInput
                v-model:value="formData.keyword"
                :placeholder="selectedTemplate?.keyword ? `默认: ${selectedTemplate.keyword}` : '请输入搜索关键词'"
                clearable
              />
            </NFormItem>
            <NFormItem label="店铺名称" :required="isShopNameRequired">
              <NInput
                v-model:value="formData.shopName"
                :placeholder="selectedTemplate?.shop_name ? `默认: ${selectedTemplate.shop_name}` : '请输入店铺名称'"
                clearable
              />
            </NFormItem>
          </template>
          <NFormItem label="执行次数">
            <NInputNumber
              v-model:value="formData.executeCount"
              :min="1"
              :max="1000"
              style="width: 100%"
              @update:value="isCreatingNew ? calculatePriceForNew() : calculatePrice()"
            />
          </NFormItem>
          <NFormItem label="选择日期">
            <NDatePicker
              type="daterange"
              style="width: 100%"
              :is-date-disabled="(ts: number) => ts < Date.now() - 86400000"
              placeholder="选择开始日期到结束日期"
              clearable
              @confirm="onDateRangeConfirm"
            />
          </NFormItem>
          <NFormItem v-if="formData.selectedDates.length > 0" label="已选日期">
            <div class="dates-and-buttons-wrapper">
              <div class="selected-dates-with-time">
                <div 
                  v-for="(dateTs, index) in formData.selectedDates" 
                  :key="index" 
                  class="date-time-item"
                >
                  <NTag 
                    closable
                    size="medium"
                    :type="getDateLabel(dateTs).type"
                    :bordered="false"
                    class="date-tag"
                    @close="removeDate(index)"
                  >
                    <span class="date-tag-content">
                      <span class="date-label" :style="{ color: getDateLabel(dateTs).color }">
                        {{ getDateLabel(dateTs).text.split(' ')[0] }}
                      </span>
                      <span class="date-time">{{ getDateLabel(dateTs).text.split(' ').slice(1).join(' ') }}</span>
                    </span>
                  </NTag>
                  <NTimePicker
                    :value="dateTs"
                    format="HH:mm"
                    value-format="timestamp"
                    :hours="Array.from({ length: 24 }, (_, i) => i)"
                    :minutes="Array.from({ length: 60 }, (_, i) => i)"
                    style="width: 100px"
                    size="small"
                    @update:value="(val) => updateDateTime(index, val)"
                  />
                </div>
              </div>
              <div class="quick-time-buttons">
                <NButton size="tiny" secondary @click="setQuickTime(1, 0)">1:00</NButton>
                <NButton size="tiny" secondary @click="setQuickTime(8, 0)">8:00</NButton>
                <NButton size="tiny" secondary @click="setQuickTime(12, 0)">12:00</NButton>
                <NButton size="tiny" secondary @click="setQuickTime(16, 0)">16:00</NButton>
                <NButton size="tiny" secondary @click="setQuickTime(21, 0)">21:00</NButton>
                <NButton size="tiny" type="primary" @click="openCustomTimePicker">
                  自定义
                </NButton>
              </div>
            </div>
          </NFormItem>
        </NForm>

        <NDivider />

        <div class="price-info">
          <NSpin :show="priceLoading" size="small">
            <div v-if="priceInfo" class="flex justify-between items-center">
              <div>
                <span class="text-gray-400">单价: </span>
                <span class="font-bold">{{ priceInfo.jingdou_price }}</span>
                <span class="text-gray-400"> 京豆/次</span>
              </div>
              <div>
                <span class="text-gray-400">单任务消耗: </span>
                <span class="font-bold">{{ priceInfo.consume_jingdou }}</span>
                <span class="text-gray-400"> 京豆</span>
              </div>
            </div>
            <div v-if="priceInfo && formData.selectedDates.length > 0" class="mt-2">
              <div class="flex justify-between items-center">
                <div>
                  <span class="text-gray-400">创建任务数: </span>
                  <span class="font-bold text-primary">{{ formData.selectedDates.length }}</span>
                  <span class="text-gray-400"> 个</span>
                </div>
                <div>
                  <span class="text-gray-400">总消耗: </span>
                  <span class="font-bold text-lg" :class="isSufficient ? 'text-primary' : 'text-error'">
                    {{ totalConsume }}
                  </span>
                  <span class="text-gray-400"> 京豆</span>
                </div>
              </div>
            </div>
            <div v-if="priceInfo" class="mt-2 text-right">
              <span class="text-gray-400">当前余额: </span>
              <span class="font-bold">{{ priceInfo.jingdou_balance }}</span>
              <span class="text-gray-400"> 京豆</span>
              <NTag v-if="!isSufficient" type="error" size="small" class="ml-2">
                余额不足
              </NTag>
            </div>
          </NSpin>
        </div>
      </template>

      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showModal = false">取消</NButton>
          <NButton
            type="primary"
            :loading="submitLoading"
            :disabled="priceLoading || !isSufficient || formData.selectedDates.length === 0 || !isFormValid"
            @click="handleSubmit"
          >
            创建 {{ formData.selectedDates.length }} 个任务
          </NButton>
        </div>
      </template>
    </NModal>

    <!-- 自定义时间选择弹窗 -->
    <NModal
      v-model:show="showCustomTimePicker"
      preset="card"
      title="自定义时间"
      style="width: 350px"
      :mask-closable="false"
    >
      <div class="custom-time-picker-content">
        <div class="mb-4">
          <span class="text-gray-400">请选择要设置的时间：</span>
        </div>
        <NTimePicker
          v-model:value="customTime"
          format="HH:mm"
          value-format="timestamp"
          :hours="Array.from({ length: 24 }, (_, i) => i)"
          :minutes="Array.from({ length: 60 }, (_, i) => i)"
          style="width: 100%"
          size="large"
        />
        <div class="mt-4 text-sm text-gray-400">
          将为所有已选日期设置相同的时间
        </div>
      </div>
      <template #footer>
        <div class="flex justify-end gap-3">
          <NButton @click="showCustomTimePicker = false">取消</NButton>
          <NButton type="primary" @click="confirmCustomTime">确定</NButton>
        </div>
      </template>
    </NModal>
  </NCard>
</template>

<style scoped>
.template-container {
  max-height: calc(3 * 200px + 2 * 16px); /* 3行卡片高度 + 间距 */
  overflow-y: auto;
  padding-right: 8px;
}

.template-list {
  display: grid;
  grid-template-columns: repeat(auto-fill, minmax(260px, 1fr));
  gap: 16px;
}

.template-card {
  padding: 16px;
  border: 1px solid rgba(255, 255, 255, 0.12);
  border-radius: 12px;
  cursor: pointer;
  transition: all 0.25s ease;
  background: linear-gradient(135deg, rgba(255, 255, 255, 0.03) 0%, rgba(255, 255, 255, 0.01) 100%);
  min-height: 180px;
  display: flex;
  flex-direction: column;
}

.template-card:hover {
  border-color: #2080f0;
  background: linear-gradient(135deg, rgba(32, 128, 240, 0.15) 0%, rgba(32, 128, 240, 0.05) 100%);
  transform: translateY(-2px);
  box-shadow: 0 8px 24px rgba(32, 128, 240, 0.15);
}

.template-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-bottom: 14px;
}

.template-price {
  font-size: 20px;
  font-weight: 700;
  color: #f0a020;
  text-shadow: 0 0 10px rgba(240, 160, 32, 0.3);
}

.price-unit {
  font-size: 12px;
  font-weight: 400;
  color: #f0a020;
  opacity: 0.8;
  margin-left: 2px;
}

.template-sku {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 16px;
  margin-bottom: 10px;
  padding: 8px 10px;
  background: rgba(32, 128, 240, 0.1);
  border-radius: 6px;
}

.sku-label {
  color: #2080f0;
  font-weight: 600;
  font-size: 13px;
}

.sku-value {
  font-family: 'Consolas', 'Monaco', monospace;
  font-size: 15px;
  font-weight: 600;
  color: #63b3ed;
  letter-spacing: 0.5px;
}

.template-info {
  display: flex;
  align-items: center;
  gap: 8px;
  font-size: 15px;
  margin-bottom: 6px;
  padding: 4px 0;
}

.info-label {
  font-size: 13px;
  font-weight: 500;
  min-width: 45px;
}

.info-value {
  font-size: 15px;
  font-weight: 500;
  max-width: 160px;
  overflow: hidden;
  text-overflow: ellipsis;
  white-space: nowrap;
}

/* 店铺名称 - 绿色系 */
.shop-info .info-label {
  color: #36d399;
}

.shop-value {
  color: #6ee7b7;
}

/* 关键词 - 紫色系 */
.keyword-info .info-label {
  color: #a78bfa;
}

.keyword-value {
  color: #c4b5fd;
}

.template-footer {
  display: flex;
  justify-content: space-between;
  align-items: center;
  margin-top: 12px;
  padding-top: 12px;
  border-top: 1px solid rgba(255, 255, 255, 0.08);
}

.usage-count {
  display: flex;
  align-items: center;
  font-size: 13px;
  color: #22c55e;
  font-weight: 500;
}

.last-used {
  font-size: 12px;
  color: #9ca3af;
}

/* 备注区域 */
.template-remark {
  margin-top: 8px;
  margin-bottom: 8px;
  flex: 1;
  display: flex;
  flex-direction: column;
}

.remark-display {
  display: flex;
  align-items: flex-start;
  gap: 4px;
  font-size: 13px;
}

.remark-label {
  color: #f59e0b;
  font-weight: 600;
  flex-shrink: 0;
}

.remark-text {
  flex: 1;
  color: #fbbf24;
  overflow: hidden;
  text-overflow: ellipsis;
  display: -webkit-box;
  -webkit-line-clamp: 3;
  -webkit-box-orient: vertical;
  line-height: 1.4;
  max-height: calc(1.4em * 3);
}

.remark-edit {
  display: flex;
  flex-direction: column;
  gap: 8px;
}

.remark-actions {
  display: flex;
  justify-content: flex-end;
  gap: 8px;
}

/* 加载更多 */
.loading-more,
.no-more {
  display: flex;
  align-items: center;
  justify-content: center;
  padding: 20px;
  color: #9ca3af;
  font-size: 14px;
}

.no-more {
  padding-top: 16px;
  padding-bottom: 24px;
}

.dates-and-buttons-wrapper {
  display: flex;
  gap: 16px;
  align-items: flex-start;
}

.selected-dates-with-time {
  display: flex;
  flex-direction: column;
  gap: 12px;
  max-height: 300px;
  overflow-y: auto;
  padding-right: 4px;
  flex: 1;
}

.date-time-item {
  display: flex;
  align-items: center;
  gap: 12px;
}

.quick-time-buttons {
  display: flex;
  flex-direction: column;
  gap: 4px;
}

.date-tag {
  padding: 4px 8px;
  border-radius: 6px;
  flex: 1;
  min-width: 0;
}

.date-tag-content {
  display: flex;
  align-items: center;
  gap: 6px;
}

.date-label {
  font-weight: 600;
  font-size: 13px;
}

.date-time {
  color: rgba(255, 255, 255, 0.7);
  font-size: 12px;
}

.price-info {
  padding: 12px;
  background: rgba(255, 255, 255, 0.03);
  border-radius: 8px;
}

.text-primary {
  color: #2080f0;
}

.text-error {
  color: #d03050;
}

.custom-time-picker-content {
  padding: 8px 0;
}
</style>
