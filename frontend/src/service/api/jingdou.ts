import { request } from '../request';

/**
 * Fetch jingdou records
 *
 * @param params Query parameters
 */
export function fetchJingdouRecords(params?: {
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
