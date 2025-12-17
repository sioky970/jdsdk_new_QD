import { request } from '../request';

/**
 * 获取用户的API Key信息
 */
export function fetchApiKey() {
  return request<Api.ApiKey.Info>({
    url: '/api/apikey',
    method: 'get'
  });
}

/**
 * 生成新的API Key
 */
export function generateApiKey() {
  return request<Api.ApiKey.Info>({
    url: '/api/apikey/generate',
    method: 'post'
  });
}

/**
 * 重置API Key
 */
export function resetApiKey() {
  return request<Api.ApiKey.Info>({
    url: '/api/apikey/reset',
    method: 'post'
  });
}

/**
 * 删除API Key
 */
export function deleteApiKey() {
  return request({
    url: '/api/apikey',
    method: 'delete'
  });
}
