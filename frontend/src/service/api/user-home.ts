import { request } from '../request';

/** 用户今日任务统计响应 */
export interface UserTodayStats {
  today_total: number;
  today_completed: number;
  today_running: number;
  today_waiting: number;
  today_pending: number;
  today_failed: number;
  today_partial_completed: number;
  completion_rate: string;
  today_consumed_jingdou: number;
  jingdou_balance: number;
  total_execute_count: number; // 今日总执行次数
  total_executed_count: number; // 今日已执行次数
  execute_completion_rate: string; // 执行完成率
}

/** 任务模板 */
export interface TaskTemplate {
  id: number;
  task_type: string;
  task_type_name: string;
  sku: string;
  shop_name: string;
  keyword: string;
  remark: string;
  total_created_count: number;
  last_used_at: string;
  jingdou_price: number;
}

/** 任务模板列表响应 */
export interface TaskTemplatesResponse {
  templates: TaskTemplate[];
  total: number;
}

/** 快速创建任务请求 */
export interface QuickCreateTaskRequest {
  template_id: number;
  execute_count: number;
  start_time: string;
  task_type?: string; // 可选，覆盖模板的任务类型
  keyword?: string; // 可选，搜索关键词
  shop_name?: string; // 可选，店铺名称
}

/** 任务类型 */
export interface TaskTypeItem {
  id: number;
  type_code: string;
  type_name: string;
  jingdou_price: number;
  is_active: boolean;
}

/** 模板价格响应 */
export interface TemplatePriceResponse {
  template_id: number;
  execute_count: number;
  task_type?: string;
  task_type_name?: string;
  jingdou_price: number;
  consume_jingdou: number;
  jingdou_balance: number;
  is_sufficient: boolean;
}

/** 快速创建任务响应 */
export interface QuickCreateTaskResponse {
  message: string;
  task_id: number;
  consume_jingdou: number;
  jingdou_balance: number;
}

/** 获取用户今日任务统计 */
export function fetchUserTodayStats() {
  return request<UserTodayStats>({
    url: '/api/user/home/today-stats',
    method: 'get'
  });
}

/** 获取任务模板列表 */
export function fetchTaskTemplates(limit: number = 10) {
  return request<TaskTemplatesResponse>({
    url: '/api/user/home/templates',
    method: 'get',
    params: { limit }
  });
}

/** 计算模板任务价格 */
export function fetchTemplatePrice(templateId: number, executeCount: number, taskType?: string) {
  return request<TemplatePriceResponse>({
    url: '/api/user/home/template-price',
    method: 'get',
    params: { template_id: templateId, execute_count: executeCount, task_type: taskType }
  });
}

/** 快速创建任务 */
export function quickCreateTask(data: QuickCreateTaskRequest) {
  return request<QuickCreateTaskResponse>({
    url: '/api/user/home/quick-create',
    method: 'post',
    data
  });
}

/** 获取已启用的任务类型列表 */
export function fetchActiveTaskTypes() {
  return request<{ task_types: TaskTypeItem[]; total: number }>({
    url: '/api/tasks/types',
    method: 'get'
  });
}

/** 京豆统计响应 */
export interface JingdouStatsResponse {
  jingdou_balance: number; // 当前余额
  today_consumed: number; // 今日已消耗
  past_7_days_consumed: number; // 过去7天消耗
  past_30_days_consumed: number; // 过去30天消耗
  future_consumed: number; // 未来待执行任务预计消耗
  available_balance: number; // 扣除未来任务后可用余额
  daily_avg_consumed: string; // 日均消耗
  estimated_days: string; // 预计可消耗天数
  estimated_days_after_future: string; // 扣除未来任务后可消耗天数
}

/** 获取用户京豆统计 */
export function fetchJingdouStats() {
  return request<JingdouStatsResponse>({
    url: '/api/user/home/jingdou-stats',
    method: 'get'
  });
}

/** 更新模板备注 */
export function updateTemplateRemark(templateId: number, remark: string) {
  return request<{ message: string; remark: string }>({
    url: `/api/user/home/templates/${templateId}/remark`,
    method: 'put',
    data: { remark }
  });
}
