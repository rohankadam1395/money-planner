import axios from 'axios';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export interface CategorizeTransactionInput {
  id: string;
  merchant: string;
  amount: number;
  timestamp: number;
}

export interface CategorizeTransactionResult {
  id: string;
  category: string;
  confidence: number;
  method: string;
  explanation: string;
}

export interface CategorizationStats {
  total: number;
  categorized: number;
  uncategorized: number;
  by_method: Record<string, number>;
  avg_confidence: number;
}

export interface CategorizeResponse {
  transactions: CategorizeTransactionResult[];
  stats: CategorizationStats;
}

export const categorizationApi = {
  // Categorize a batch of transactions
  async categorize(transactions: CategorizeTransactionInput[]): Promise<CategorizeResponse> {
    try {
      const response = await axios.post<CategorizeResponse>(
        `${API_BASE_URL}/api/v1/transactions/categorize`,
        { transactions }
      );
      return response.data;
    } catch (error) {
      console.error('Categorization error:', error);
      throw error;
    }
  },

  // Get all categories
  async getCategories() {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/categories`);
      return response.data;
    } catch (error) {
      console.error('Get categories error:', error);
      throw error;
    }
  },

  // Get category by ID
  async getCategoryById(categoryId: string) {
    try {
      const response = await axios.get(`${API_BASE_URL}/api/v1/categories/${categoryId}`);
      return response.data;
    } catch (error) {
      console.error('Get category error:', error);
      throw error;
    }
  },
};

export default categorizationApi;
