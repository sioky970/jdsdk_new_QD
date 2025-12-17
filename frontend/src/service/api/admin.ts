import { request } from '../request';

/** 用户列表项 */
export interface AdminUser {
  id: number;
  username: string;
  nickname: string;
  role: string;
  jingdou_balance: number;
  is_active: boolean;
  created_at: string;
  last_login: string | null;
  api_key?: string;
  api_key_created_at?: string | null;
  api_key_last_used_at?: string | null;
  // 用户统计信息
  pending_task_count?: number;      // 未完成任务数
  pending_execute_count?: number;   // 未完成执行次数
  pending_jingdou?: number;         // 未完成任务京豆合计
  pending_task_percent?: number;    // 未完成任务百分比
  pending_execute_percent?: number; // 未完成次数百分比
  history_consumed_jingdou?: number; // 历史消耗京豆
}

/** 用户列表响应 */
export interface AdminUsersResponse {
  items: AdminUser[];
  total: number;
  page: number;
  per_page: number;
  pages: number;
}

/** 创建用户请求 */
export interface CreateUserRequest {
  username: string;
  password: string;
  nickname?: string;
  role?: string;
  jingdou_balance?: number;
}

/** 更新用户请求 */
export interface UpdateUserRequest {
  nickname?: string;
  role?: string;
  is_active?: boolean;
  jingdou_balance?: number;
  password?: string;
}

/** 获取用户列表（管理员） */
export function fetchAdminUsers(params?: {
  page?: number;
  per_page?: number;
  search?: string;
}) {
  return request<AdminUsersResponse>({
    url: '/api/users',
    method: 'get',
    params
  });
}

/** 获取用户详情（管理员） */
export function fetchAdminUserById(userId: number) {
  return request<AdminUser>({
    url: `/api/users/${userId}`,
    method: 'get'
  });
}

/** 创建用户（管理员） */
export function createAdminUser(data: CreateUserRequest) {
  return request<{ id: number; username: string }>({
    url: '/api/users',
    method: 'post',
    data
  });
}

/** 更新用户（管理员） */
export function updateAdminUser(userId: number, data: UpdateUserRequest) {
  return request({
    url: `/api/users/${userId}`,
    method: 'put',
    data
  });
}

/** 删除用户（管理员） */
export function deleteAdminUser(userId: number) {
  return request({
    url: `/api/users/${userId}`,
    method: 'delete'
  });
}

/** 搜索用户（管理员） */
export function searchAdminUsers(params: { keyword: string }) {
  return request<{ users: AdminUser[] }>({
    url: '/api/users/search',
    method: 'get',
    params
  });
}

/** 调整任务优先级（管理员） */
export function updateTaskPriority(taskId: number, priority: number) {
  return request({
    url: `/api/tasks/${taskId}/priority`,
    method: 'put',
    data: { priority }
  });
}

/** 调整用户京豆（管理员） */
export function adjustUserJingdou(
  userId: number,
  data: {
    amount: number;
    operation_type?: string;
    remark?: string;
  }
) {
  return request<{ new_balance: number }>({
    url: `/api/users/${userId}/jingdou`,
    method: 'post',
    data
  });
}

/** 获取所有用户的京豆记录（管理员） */
export function fetchAllJingdouRecords(params?: {
  user_id?: number;
  type?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  per_page?: number;
}) {
  return request<Api.Jingdou.RecordsResponse>({
    url: '/api/jingdou/records',
    method: 'get',
    params
  });
}

/** 获取所有任务（管理员） */
export function fetchAllTasks(params?: {
  user_id?: number;
  status?: string;
  start_date?: string;
  end_date?: string;
  page?: number;
  per_page?: number;
  sort_by?: string;
  sort_order?: string;
  keyword?: string;
}) {
  return request({
    url: '/api/tasks',
    method: 'get',
    params
  });
}

/** 获取任务详情（管理员） */
export function fetchTaskDetail(taskId: number) {
  return request({
    url: `/api/tasks/${taskId}`,
    method: 'get'
  });
}

/** 获取用户API Key（管理员） */
export function getUserApiKey(userId: number) {
  return request<{
    api_key: string | null;
    created_at: string | null;
    last_used_at: string | null;
  }>({
    url: `/api/users/${userId}/apikey`,
    method: 'get'
  });
}

/** 重置用户API Key（管理员） */
export function resetUserApiKey(userId: number) {
  return request<{
    api_key: string;
    created_at: string;
  }>({
    url: `/api/users/${userId}/apikey`,
    method: 'post'
  });
}

/** 删除用户API Key（管理员） */
export function deleteUserApiKey(userId: number) {
  return request({
    url: `/api/users/${userId}/apikey`,
    method: 'delete'
  });
}
