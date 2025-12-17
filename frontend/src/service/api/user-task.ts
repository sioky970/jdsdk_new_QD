import { request } from '../request';

/** 用户任务项 */
export interface UserTask {
  id: number;
  task_type: string;
  task_type_name: string;
  sku: string;
  shop_name: string;
  keyword: string;
  start_time: string;
  execute_count: number;
  executed_count: number;
  priority: number;
  status: string;
  status_text: string;
  consume_jingdou: number;
  can_cancel: boolean;
  can_edit: boolean;
  created_at: string;
}

/** 用户任务列表响应 */
export interface UserTasksResponse {
  tasks: UserTask[];
  total: number;
  page: number;
  per_page: number;
  pages: number;
}

/** 任务查询参数 */
export interface UserTaskQuery {
  status?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  per_page?: number;
}

/** 状态选项 */
export interface StatusOption {
  value: string;
  label: string;
}

/** 更新任务请求 */
export interface UpdateTaskRequest {
  shop_name?: string;
  keyword?: string;
  start_time?: string;
  execute_count?: number;
}

/** 获取用户任务列表 */
export function fetchUserTasks(params: UserTaskQuery) {
  return request<UserTasksResponse>({
    url: '/api/user/tasks',
    method: 'get',
    params
  });
}

/** 获取任务状态选项 */
export function fetchTaskStatusOptions() {
  return request<{ options: StatusOption[] }>({
    url: '/api/user/tasks/status-options',
    method: 'get'
  });
}

/** 取消任务 */
export function cancelUserTask(taskId: number) {
  return request<{ task_id: number; refund_jingdou: number; new_balance: number }>({
    url: `/api/user/tasks/${taskId}/cancel`,
    method: 'post'
  });
}

/** 更新任务 */
export function updateUserTask(taskId: number, data: UpdateTaskRequest) {
  return request<{ task_id: number; additional_jingdou: number }>({
    url: `/api/user/tasks/${taskId}`,
    method: 'put',
    data
  });
}

/** 创建任务请求 */
export interface CreateTaskRequest {
  task_type: string;
  sku: string;
  shop_name?: string;
  keyword?: string;
  start_time: string;
  execute_count: number;
  priority?: number;
  remark?: string;
}

/** 批量创建任务请求 */
export interface BatchCreateTaskRequest {
  tasks: CreateTaskRequest[];
}

/** 创建任务响应 */
export interface CreateTaskResponse {
  task_id: number;
  consume_jingdou: number;
  balance: number;
}

/** 批量创建任务响应 */
export interface BatchCreateTaskResponse {
  success_count: number;
  fail_count: number;
  total_consume: number;
  balance: number;
  created_tasks: Array<{
    task_id: number;
    sku: string;
    consume_jingdou: number;
  }>;
  failed_tasks?: Array<{
    index: number;
    sku: string;
    error: string;
  }>;
}

/** 创建单个任务 */
export function createTask(data: CreateTaskRequest) {
  return request<CreateTaskResponse>({
    url: '/api/tasks',
    method: 'post',
    data
  });
}

/** 批量创建任务 */
export function batchCreateTasks(data: BatchCreateTaskRequest) {
  return request<BatchCreateTaskResponse>({
    url: '/api/tasks/apikey/batch',
    method: 'post',
    data
  });
}
