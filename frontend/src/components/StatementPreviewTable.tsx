import React from 'react';
import { TransactionPreviewRow } from './TransactionPreviewRow';

interface Transaction {
  transaction_id: string;
  transaction_date: string;
  merchant: string;
  amount: number;
  type: string;
  balance?: number;
  description?: string;
  currency?: string;
  category?: string;
  category_confidence?: number;
  category_method?: string;
}

interface CategorizationStats {
  total: number;
  categorized: number;
  uncategorized: number;
  by_method: Record<string, number>;
  avg_confidence: number;
}

interface StatementPreviewTableProps {
  transactions: Transaction[];
  categorizationStats?: CategorizationStats;
}

// Component for displaying the full statement preview with transactions and categories
export const StatementPreviewTable: React.FC<StatementPreviewTableProps> = ({
  transactions,
  categorizationStats,
}) => {
  if (!transactions || transactions.length === 0) {
    return (
      <div className="text-center py-8 text-gray-500">
        No transactions to display
      </div>
    );
  }

  return (
    <div className="w-full">
      {/* Categorization Summary */}
      {categorizationStats && (
        <div className="mb-6 p-4 bg-blue-50 dark:bg-blue-900 rounded-lg">
          <h3 className="text-sm font-semibold text-blue-900 dark:text-blue-100 mb-3">
            Categorization Summary
          </h3>
          <div className="grid grid-cols-4 gap-4 text-sm">
            <div>
              <p className="text-blue-600 dark:text-blue-300 font-medium">
                {categorizationStats.total}
              </p>
              <p className="text-blue-700 dark:text-blue-200">Total</p>
            </div>
            <div>
              <p className="text-green-600 dark:text-green-300 font-medium">
                {categorizationStats.categorized}
              </p>
              <p className="text-green-700 dark:text-green-200">Categorized</p>
            </div>
            <div>
              <p className="text-yellow-600 dark:text-yellow-300 font-medium">
                {categorizationStats.uncategorized}
              </p>
              <p className="text-yellow-700 dark:text-yellow-200">Uncategorized</p>
            </div>
            <div>
              <p className="text-purple-600 dark:text-purple-300 font-medium">
                {(categorizationStats.avg_confidence * 100).toFixed(0)}%
              </p>
              <p className="text-purple-700 dark:text-purple-200">Avg Confidence</p>
            </div>
          </div>
          {Object.keys(categorizationStats.by_method).length > 0 && (
            <div className="mt-3 pt-3 border-t border-blue-200 dark:border-blue-700">
              <p className="text-xs font-semibold text-blue-900 dark:text-blue-100 mb-2">
                By Method:
              </p>
              <div className="flex gap-4 flex-wrap text-xs">
                {Object.entries(categorizationStats.by_method).map(([method, count]) => (
                  <span key={method} className="text-blue-700 dark:text-blue-300">
                    {method}: <span className="font-semibold">{count}</span>
                  </span>
                ))}
              </div>
            </div>
          )}
        </div>
      )}

      {/* Transactions Table */}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="border-b bg-gray-50 dark:bg-gray-800">
              <th className="px-4 py-3 text-left font-semibold text-gray-900 dark:text-white">
                Date
              </th>
              <th className="px-4 py-3 text-left font-semibold text-gray-900 dark:text-white">
                Merchant
              </th>
              <th className="px-4 py-3 text-right font-semibold text-gray-900 dark:text-white">
                Amount
              </th>
              <th className="px-4 py-3 text-left font-semibold text-gray-900 dark:text-white">
                Type
              </th>
              <th className="px-4 py-3 text-left font-semibold text-gray-900 dark:text-white">
                Category
              </th>
              {transactions.some(t => t.description) && (
                <th className="px-4 py-3 text-left font-semibold text-gray-900 dark:text-white">
                  Description
                </th>
              )}
            </tr>
          </thead>
          <tbody>
            {transactions.map((transaction) => (
              <TransactionPreviewRow
                key={transaction.transaction_id}
                transaction={transaction}
              />
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
};

export default StatementPreviewTable;
