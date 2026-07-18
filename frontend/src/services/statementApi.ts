import { apiClient } from './api';
import { getCategoryStyle } from '@/constants/categories';

// Types for statement operations
export interface StatementUploadRequest {
  file: File;
  bank_code: string;
}

export interface StatementUploadResponse {
  statement_id: string;
  status: string;
  bank_code: string;
  file_name: string;
  file_format: string;
  uploaded_at: string;
  error_message?: string;
}

export interface Transaction {
  transaction_id: string;
  statement_id: string;
  user_id: string;
  transaction_date: string;
  merchant: string;
  amount: number;
  type: 'DEBIT' | 'CREDIT';
  description?: string;
  balance?: number;
  currency?: string;
  bank_code?: string;
  account_number_hash?: string;
  imported_at: string;
  created_at: string;
  updated_at: string;
  category?: {
    id: string;
    name: string;
    color: string;
    icon: string;
    confidence: number;
    method: 'rule_based' | 'fuzzy' | 'llm' | 'none';
    llm_provider?: string;
  };
}

export interface ValidationError {
  transaction_index: number;
  field: string;
  error: string;
}

export interface PreviewResponse {
  statement_id: string;
  transactions: Transaction[];
  validation_summary: {
    total_rows: number;
    valid_transactions: number;
    invalid_transactions: number;
    errors: ValidationError[];
    period_start?: string;
    period_end?: string;
  };
}

export interface ConfirmImportRequest {
  statement_id: string;
  confirmed: boolean;
}

export interface ConfirmImportResponse {
  statement_id: string;
  status: string;
  transactions_imported: number;
  message: string;
}

export interface Statement {
  id: string;
  user_id: string;
  bank_code: string;
  file_name: string;
  file_format: string;
  file_size_bytes: number;
  file_hash: string;
  statement_period_start?: string;
  statement_period_end?: string;
  status: string;
  created_at: string;
  updated_at: string;
}

// Statement API service
export const statementApi = {
  // Upload a statement file
  uploadStatement: async (request: StatementUploadRequest): Promise<StatementUploadResponse> => {
    const formData = new FormData();
    formData.append('file', request.file);
    formData.append('bank_code', request.bank_code);

    try {
      const response = await apiClient.post<StatementUploadResponse>(
        '/api/v1/statements/upload',
        formData,
        {
          headers: {
            'Content-Type': 'multipart/form-data',
          },
        }
      );
      return response.data;
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        'Failed to upload statement';
      throw new Error(errorMessage);
    }
  },

  // Get preview of extracted transactions
  getPreview: async (statementId: string): Promise<PreviewResponse> => {
    try {
      const token = localStorage.getItem('authToken');
      const response = await apiClient.get<PreviewResponse>(
        `/api/v1/statements/${statementId}/preview`,
        {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        }
      );
      return response.data;
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        'Failed to fetch preview';
      throw new Error(errorMessage);
    }
  },

  // Confirm and persist imported transactions
  confirmImport: async (request: ConfirmImportRequest): Promise<ConfirmImportResponse> => {
    try {
      const token = localStorage.getItem('authToken');
      const response = await apiClient.post<ConfirmImportResponse>(
        `/api/v1/statements/${request.statement_id}/confirm`,
        { confirmed: request.confirmed },
        {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        }
      );
      return response.data;
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        'Failed to confirm import';
      throw new Error(errorMessage);
    }
  },

  // Get list of user's uploaded statements
  getStatements: async (limit = 10, offset = 0): Promise<Statement[]> => {
    try {
      const response = await apiClient.get<Statement[]>('/api/v1/statements', {
        params: { limit, offset },
      });
      return response.data;
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        'Failed to fetch statements';
      throw new Error(errorMessage);
    }
  },

  // Get all transactions across statements
  getTransactions: async (
    bankCode?: string,
    dateStart?: string,
    dateEnd?: string,
    limit = 50,
    offset = 0
  ): Promise<Transaction[]> => {
    try {
      const params: any = { limit, offset };
      if (bankCode) params.bank_code = bankCode;
      if (dateStart) params.date_start = dateStart;
      if (dateEnd) params.date_end = dateEnd;

      const response = await apiClient.get<Transaction[]>('/api/v1/transactions', {
        params,
      });
      return response.data;
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        'Failed to fetch transactions';
      throw new Error(errorMessage);
    }
  },

  // Delete a statement
  deleteStatement: async (statementId: string): Promise<void> => {
    try {
      const token = localStorage.getItem('authToken');
      await apiClient.delete(
        `/api/v1/statements/${statementId}`,
        {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        }
      );
    } catch (error: any) {
      const errorMessage =
        error.response?.data?.error ||
        error.response?.data?.message ||
        error.message ||
        'Failed to delete statement';
      throw new Error(errorMessage);
    }
  },

  // Categorize transactions
  categorizeTransactions: async (transactions: Transaction[]): Promise<Transaction[]> => {
    try {
      const token = localStorage.getItem('authToken');

      // Build request payload
      const req = {
        transactions: transactions.map((t) => ({
          id: t.transaction_id,
          merchant: t.merchant,
          amount: t.amount,
          timestamp: new Date(t.transaction_date).getTime() / 1000,
        })),
      };

      const response = await apiClient.post<{
        transactions: Array<{
          id: string;
          category: string;
          confidence: number;
          method: string;
          explanation: string;
          llm_provider?: string;
        }>;
        stats: any;
      }>(
        '/api/transactions/categorize',
        req,
        {
          headers: token ? { Authorization: `Bearer ${token}` } : {},
        }
      );

      // Map API response back to transactions with category details
      const categorizeMap = new Map(
        response.data.transactions.map((c) => [
          c.id,
          {
            name: c.category,
            confidence: c.confidence,
            method: c.method as 'rule_based' | 'fuzzy' | 'llm' | 'none',
            llm_provider: c.llm_provider,
          },
        ])
      );

      return transactions.map((t) => {
        const categorization = categorizeMap.get(t.transaction_id);
        if (categorization) {
          const { color, icon } = getCategoryStyle(categorization.name);
          // Generate category ID from name (e.g., "Food & Dining" -> "cat_food")
          const categoryId = `cat_${categorization.name.toLowerCase().replace(/[^a-z0-9]+/g, '_')}`;
          return {
            ...t,
            category: {
              id: categoryId,
              name: categorization.name,
              color,
              icon,
              confidence: categorization.confidence,
              method: categorization.method,
              llm_provider: categorizeMap.get(t.transaction_id)?.llm_provider,
            },
          };
        }
        return t;
      });
    } catch (error: any) {
      console.warn('Categorization failed, returning transactions without categories:', error);
      return transactions;
    }
  },
};

// Polling utility for statement processing
export const pollStatementPreview = async (
  statementId: string,
  maxAttempts = 30,
  delayMs = 1000
): Promise<PreviewResponse> => {
  for (let attempt = 0; attempt < maxAttempts; attempt++) {
    try {
      return await statementApi.getPreview(statementId);
    } catch (error) {
      if (attempt === maxAttempts - 1) {
        throw error;
      }
      await new Promise((resolve) => setTimeout(resolve, delayMs));
    }
  }
  throw new Error('Statement processing timeout');
};
