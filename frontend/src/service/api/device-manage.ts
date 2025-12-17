import { request } from '../request';

/** 设备信息 */
export interface Device {
  id: number;
  device_id: string;
  device_name: string;
  device_type: 'android' | 'ios' | '';  // 设备类型
  device_model: string;                  // 设备型号
  os_version: string;                    // 系统版本
  app_version: string;                   // 应用版本
  ip: string;
  location: string;
  os_info: string;                       // 兼容旧字段
  version: string;                       // 兼容旧字段
  status: 'online' | 'offline' | 'working' | 'idle';
  is_blocked: boolean;
  last_heartbeat: string | null;
  last_active: string | null;
  last_task_time: string | null;
  task_count: number;
  hourly_rate: number;    // 每小时任务执行数
  daily_estimate: number; // 预估每天可完成数
  created_at: string;
}

/** 设备列表响应 */
export interface DeviceListResponse {
  items: Device[];
  total: number;
  page: number;
  page_size: number;
}

/** 设备统计 */
export interface DeviceStatistics {
  total_devices: number;
  online_devices: number;
  offline_devices: number;
  working_devices: number;
  idle_devices: number;
}

/** 获取设备列表 */
export function fetchDevices(params?: { page?: number; page_size?: number }) {
  return request<DeviceListResponse>({
    url: '/api/devices',
    method: 'get',
    params
  });
}

/** 获取设备统计 */
export function fetchDeviceStatistics() {
  return request<DeviceStatistics>({
    url: '/api/devices/statistics',
    method: 'get'
  });
}

/** 获取设备详情 */
export function fetchDeviceById(id: number) {
  return request<Device>({
    url: `/api/devices/${id}`,
    method: 'get'
  });
}

/** 更新设备封禁状态 */
export function updateDeviceBlockStatus(id: number, isBlocked: boolean) {
  return request({
    url: `/api/devices/${id}/status`,
    method: 'put',
    data: { is_blocked: isBlocked }
  });
}

/** 清空所有设备 */
export function clearAllDevices() {
  return request<{ deleted_count: number }>({
    url: '/api/devices/clear-all',
    method: 'post'
  });
}
