import React from 'react';
import { CategoryBadge } from './CategoryBadge';

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

interface TransactionPreviewRowProps {
  transaction: Transaction;
}

// Component for displaying a single transaction in preview with category
export const TransactionPreviewRow: React.FC<TransactionPreviewRowProps> = ({
  transaction,
}) => {
  const categoryColors: Record<string, string> = {
    'Food & Dining': '#FF6B6B',
    'Shopping': '#4ECDC4',
    'Transport': '#45B7D1',
    'Housing': '#F7B731',
    'Utilities': '#5F27CD',
    'Entertainment': '#EE5A6F',
    'Income': '#2ECC71',
    'Healthcare': '#FF4757',
    'Education': '#1E90FF',
    'Miscellaneous': '#95A5A6',
    'Uncategorized': '#95A5A6',
  };

  const categoryIcons: Record<string, string> = {
    'Food & Dining': '🍔',
    'Shopping': '🛍️',
    'Transport': '🚗',
    'Housing': '🏠',
    'Utilities': '💡',
    'Entertainment': '🎬',
    'Income': '💰',
    'Healthcare': '🏥',
    'Education': '📚',
    'Miscellaneous': '📌',
    'Uncategorized': '❓',
  };

  const category = transaction.category || 'Uncategorized';
  const color = categoryColors[category] || '#95A5A6';
  const icon = categoryIcons[category] || '📌';

  return (
    <tr className="border-b hover:bg-gray-50 dark:hover:bg-gray-800">
      <td className="px-4 py-3 text-sm">{transaction.transaction_date}</td>
      <td className="px-4 py-3 text-sm font-medium">{transaction.merchant}</td>
      <td className="px-4 py-3 text-sm text-right">
        ₹{transaction.amount.toFixed(2)}
      </td>
      <td className="px-4 py-3 text-sm">
        <span className={`inline-block px-2 py-1 rounded text-xs font-medium ${
          transaction.type === 'CREDIT'
            ? 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200'
            : 'bg-red-100 text-red-800 dark:bg-red-900 dark:text-red-200'
        }`}>
          {transaction.type}
        </span>
      </td>
      <td className="px-4 py-3 text-sm">
        <CategoryBadge
          name={category}
          color={color}
          icon={icon}
          confidence={transaction.category_confidence}
          method={transaction.category_method}
          size="sm"
        />
      </td>
      {transaction.description && (
        <td className="px-4 py-3 text-sm text-gray-600 dark:text-gray-400">
          {transaction.description}
        </td>
      )}
    </tr>
  );
};

export default TransactionPreviewRow;
