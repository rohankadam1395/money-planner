import React, { useState } from 'react';
import { Transaction } from '@/services/statementApi';

interface TransactionPreviewProps {
  transactions: Transaction[];
  pageSize?: number;
  onSelectTransaction?: (transaction: Transaction) => void;
}

export function TransactionPreview({
  transactions,
  pageSize = 10,
  onSelectTransaction,
}: TransactionPreviewProps) {
  const [currentPage, setCurrentPage] = useState(0);

  const totalPages = Math.ceil(transactions.length / pageSize);
  const startIndex = currentPage * pageSize;
  const endIndex = startIndex + pageSize;
  const paginatedTransactions = transactions.slice(startIndex, endIndex);

  const formatDate = (dateStr: string) => {
    try {
      return new Date(dateStr).toLocaleDateString('en-IN', {
        year: 'numeric',
        month: 'short',
        day: 'numeric',
      });
    } catch {
      return dateStr;
    }
  };

  const formatAmount = (amount: number) => {
    return new Intl.NumberFormat('en-IN', {
      style: 'currency',
      currency: 'INR',
      minimumFractionDigits: 2,
    }).format(amount);
  };

  return (
    <div className="w-full">
      <div className="overflow-x-auto">
        <table className="w-full border-collapse">
          <thead className="bg-gray-100 sticky top-0">
            <tr>
              <th className="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">
                Date
              </th>
              <th className="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">
                Description
              </th>
              <th className="px-4 py-2 text-right text-sm font-semibold text-gray-700 border-b">
                Amount
              </th>
              <th className="px-4 py-2 text-center text-sm font-semibold text-gray-700 border-b">
                Type
              </th>
            </tr>
          </thead>
          <tbody>
            {paginatedTransactions.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-4 py-8 text-center text-gray-500">
                  No transactions to display
                </td>
              </tr>
            ) : (
              paginatedTransactions.map((txn, idx) => (
                <tr
                  key={txn.transaction_id || idx}
                  onClick={() => onSelectTransaction?.(txn)}
                  className="border-b hover:bg-blue-50 transition-colors cursor-pointer"
                >
                  <td className="px-4 py-3 text-sm text-gray-900">
                    {formatDate(txn.transaction_date)}
                  </td>
                  <td className="px-4 py-3 text-sm text-gray-700">
                    <div className="truncate">{txn.description}</div>
                    {txn.merchant && (
                      <div className="text-xs text-gray-500">{txn.merchant}</div>
                    )}
                  </td>
                  <td className="px-4 py-3 text-sm text-right font-medium text-gray-900">
                    {formatAmount(txn.amount)}
                  </td>
                  <td className="px-4 py-3 text-sm text-center">
                    <span
                      className={`inline-block px-2 py-1 rounded text-xs font-semibold ${
                        txn.type === 'CREDIT'
                          ? 'bg-green-100 text-green-800'
                          : 'bg-red-100 text-red-800'
                      }`}
                    >
                      {txn.type}
                    </span>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination controls */}
      {totalPages > 1 && (
        <div className="mt-4 flex items-center justify-between">
          <div className="text-sm text-gray-600">
            Showing {startIndex + 1} to {Math.min(endIndex, transactions.length)} of{' '}
            {transactions.length} transactions
          </div>
          <div className="flex gap-2">
            <button
              onClick={() => setCurrentPage((p) => Math.max(0, p - 1))}
              disabled={currentPage === 0}
              className="px-3 py-1 rounded border border-gray-300 text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Previous
            </button>
            <div className="flex items-center gap-1">
              {Array.from({ length: totalPages }, (_, i) => (
                <button
                  key={i}
                  onClick={() => setCurrentPage(i)}
                  className={`px-2 py-1 rounded text-sm ${
                    currentPage === i
                      ? 'bg-blue-600 text-white'
                      : 'border border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  {i + 1}
                </button>
              ))}
            </div>
            <button
              onClick={() => setCurrentPage((p) => Math.min(totalPages - 1, p + 1))}
              disabled={currentPage === totalPages - 1}
              className="px-3 py-1 rounded border border-gray-300 text-sm disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
            >
              Next
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
