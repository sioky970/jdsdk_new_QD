/**
 * Namespace ApiKey
 *
 * Backend API module: "apikey"
 */
declare namespace Api {
  namespace ApiKey {
    /** API Key信息 */
    interface Info {
      /** API Key */
      api_key: string;
      /** 创建时间 */
      created_at: string;
      /** 最后使用时间 */
      last_used_at: string;
    }
  }
}
