import { request } from '../request';

/** 今日任务统计响应 */
export interface TodayTaskStats {
  stat_mode: string;
  task_type: string;
  unit: string;
  total_tasks: number;
  total_value: number;
  pending_value: number;
  completed_value: number;
  running_value: number;
  pending_percent: number;
  completed_percent: number;
  running_percent: number;
}

/** 今日任务统计查询参数 */
export interface TodayTaskStatsQuery {
  stat_mode?: 'count' | 'execute';
  task_type?: string;
}

/** 每日任务数量 */
export interface DayTaskCount {
  day: string;
  count: number;
}

/** 任务执行压力响应 */
export interface TaskPressure {
  stat_mode: string;
  task_type: string;
  unit: string;
  future_tasks: DayTaskCount[];
  total_future_pending: number;
  yesterday_completed: number;
  avg_3days_completed: number;
  pressure_level: string;
  pressure_value: number;
}

/** 低余额用户 */
export interface LowBalanceUser {
  id: number;
  username: string;
  nickname: string;
  jingdou_balance: number;
}

/** 财务统计响应 */
export interface FinanceStats {
  avg_daily_recharge: number;
  avg_daily_consume: number;
  today_recharge: number;
  today_consume: number;
  low_balance_users: LowBalanceUser[];
}

/** 获取今日任务统计 */
export function fetchTodayTaskStats(params?: TodayTaskStatsQuery) {
  return request<TodayTaskStats>({
    url: '/api/admin/dashboard/today-tasks',
    method: 'get',
    params
  });
}

/** 获取任务执行压力 */
export function fetchTaskPressure(params?: TodayTaskStatsQuery) {
  return request<TaskPressure>({
    url: '/api/admin/dashboard/task-pressure',
    method: 'get',
    params
  });
}

/** 获取财务统计 */
export function fetchFinanceStats() {
  return request<FinanceStats>({
    url: '/api/admin/dashboard/finance',
    method: 'get'
  });
}
