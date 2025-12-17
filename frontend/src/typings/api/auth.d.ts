declare namespace Api {
  /**
   * namespace Auth
   *
   * backend api module: "auth"
   */
  namespace Auth {
    interface LoginToken {
      id: number;
      username: string;
      nickname: string;
      role: string;
      access_token: string;
      refresh_token: string;
      expires: number;
    }

    interface UserInfo {
      id: number;
      username: string;
      nickname: string;
      avatar?: string;
      role: string;
      jingdou_balance: number;
      created_at: string;
      last_login?: string | null;
    }
  }
}
