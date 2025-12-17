import { request } from '../request';

/** 任务类型（管理员完整信息） */
export interface TaskTypeAdmin {
  id: number;
  type_code: string;
  type_name: string;
  jingdou_price: number;
  is_active: boolean;
  execute_multiplier: number; // 执行倍数（仅管理员可见）
  is_system_preset: boolean;
  has_time_limit: boolean;
  time_slot1_start?: string;
  time_slot1_end?: string;
  time_slot2_start?: string;
  time_slot2_end?: string;
  time_slots: string[];
  created_at: string;
  updated_at: string;
}

/** 任务类型列表响应 */
export interface TaskTypesResponse {
  task_types: TaskTypeAdmin[];
  total: number;
}

/** 更新任务类型请求 */
export interface UpdateTaskTypeRequest {
  type_name?: string;
  jingdou_price?: number;
  is_active?: boolean;
  execute_multiplier?: number;
  time_slot1_start?: string;
  time_slot1_end?: string;
  time_slot2_start?: string;
  time_slot2_end?: string;
}

/** 获取任务类型列表（管理员） */
export function fetchTaskTypesAdmin() {
  return request<TaskTypesResponse>({
    url: '/api/tasks/types',
    method: 'get'
  });
}

/** 更新任务类型（管理员） */
export function updateTaskType(typeId: number, data: UpdateTaskTypeRequest) {
  return request({
    url: `/api/tasks/types/${typeId}`,
    method: 'put',
    data
  });
}

/** 获取系统公告 */
export function fetchAnnouncement() {
  return request<{ announcement: string }>({
    url: '/api/settings/announcement',
    method: 'get'
  });
}

/** 更新系统公告（管理员） */
export function updateAnnouncement(announcement: string) {
  return request({
    url: '/api/settings/announcement',
    method: 'put',
    data: { announcement }
  });
}

/** 获取设备认证密钥（管理员） */
export function fetchDeviceAuthKey() {
  return request<{ device_key: string }>({
    url: '/api/settings/device-key',
    method: 'get'
  });
}

/** 更新设备认证密钥（管理员） */
export function updateDeviceAuthKey(deviceKey: string) {
  return request({
    url: '/api/settings/device-key',
    method: 'put',
    data: { device_key: deviceKey }
  });
}
