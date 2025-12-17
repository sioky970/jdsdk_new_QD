declare namespace Api {
  namespace Jingdou {
    /** Jingdou record */
    interface Record {
      id: number;
      user_id: number;
      type: string;
      amount: number;
      balance_after: number;
      task_id?: number;
      remark?: string;
      created_at: string;
    }

    /** Jingdou records response */
    interface RecordsResponse {
      records: Record[];
      total: number;
      page: number;
      per_page: number;
    }
  }
}
