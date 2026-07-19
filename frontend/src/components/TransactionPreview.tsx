import React, { useState } from 'react';
import { Transaction } from '@/services/statementApi';
import RecategorizeModal from './RecategorizeModal';

interface TransactionPreviewProps {
  transactions: Transaction[];
  pageSize?: number;
  onSelectTransaction?: (transaction: Transaction) => void;
  onRecategorize?: (transactionId: string, newCategoryId: string, learnCorrection: boolean) => Promise<void>;
  categories?: any[];
  selectedForEnhance?: Set<string>;
  onToggleSelect?: (transactionId: string) => void;
  onToggleSelectAll?: () => void;
}

export function TransactionPreview({
  transactions,
  pageSize = 10,
  onSelectTransaction,
  onRecategorize,
  categories = [],
  selectedForEnhance = new Set(),
  onToggleSelect,
  onToggleSelectAll,
}: TransactionPreviewProps) {
  const [currentPage, setCurrentPage] = useState(0);
  const [recategorizeModal, setRecategorizeModal] = useState<{
    isOpen: boolean;
    transaction?: Transaction;
  }>({ isOpen: false });
  const [isRecategorizing, setIsRecategorizing] = useState(false);

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
              {onToggleSelect && (
                <th className="px-4 py-2 text-center text-sm font-semibold text-gray-700 border-b w-10">
                  <input
                    type="checkbox"
                    checked={selectedForEnhance.size === transactions.length && transactions.length > 0}
                    onChange={onToggleSelectAll}
                    className="w-4 h-4 cursor-pointer"
                    title="Select all for enhancement"
                  />
                </th>
              )}
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
              <th className="px-4 py-2 text-left text-sm font-semibold text-gray-700 border-b">
                Category
              </th>
            </tr>
          </thead>
          <tbody>
            {paginatedTransactions.length === 0 ? (
              <tr>
                <td colSpan={5} className="px-4 py-8 text-center text-gray-500">
                  No transactions to display
                </td>
              </tr>
            ) : (
              paginatedTransactions.map((txn, idx) => (
                <tr
                  key={txn.transaction_id || idx}
                  className={`border-b transition-colors ${
                    selectedForEnhance.has(txn.transaction_id)
                      ? 'bg-indigo-50'
                      : 'hover:bg-blue-50'
                  } cursor-pointer`}
                >
                  {onToggleSelect && (
                    <td className="px-4 py-3 text-center w-10">
                      <input
                        type="checkbox"
                        checked={selectedForEnhance.has(txn.transaction_id)}
                        onChange={(e) => {
                          e.stopPropagation();
                          onToggleSelect(txn.transaction_id);
                        }}
                        className="w-4 h-4 cursor-pointer"
                      />
                    </td>
                  )}
                  <td
                    className="px-4 py-3 text-sm text-gray-900 cursor-pointer"
                    onClick={() => onSelectTransaction?.(txn)}
                  >
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
                  <td className="px-4 py-3 text-sm">
                    {txn.category ? (
                      <div className={`flex flex-col gap-2 ${
                        txn.category.confidence < 0.75 ? 'p-2 bg-amber-50 rounded-lg' : ''
                      }`}>
                        <div
                          className={`px-2 py-1 rounded text-xs font-semibold text-white w-fit ${
                            txn.category.confidence < 0.75 ? 'ring-2 ring-amber-400' : ''
                          }`}
                          style={{
                            backgroundColor: txn.category.confidence < 0.75 ? '#F59E0B' : txn.category.color
                          }}
                        >
                          {txn.category.icon} {txn.category.name}
                          {txn.category.confidence < 0.75 && ' ⚠️'}
                        </div>
                        <div className="flex items-center justify-between">
                          <div className="text-xs text-gray-600">
                            {(txn.category.confidence * 100).toFixed(0)}% • {txn.category.method}
                            {txn.category.llm_provider && ` (${txn.category.llm_provider})`}
                          </div>
                          {onRecategorize && (
                            <button
                              onClick={(e) => {
                                e.stopPropagation();
                                setRecategorizeModal({ isOpen: true, transaction: txn });
                              }}
                              className="text-xs px-2 py-1 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                            >
                              Edit
                            </button>
                          )}
                        </div>
                      </div>
                    ) : (
                      <div className="flex items-center justify-between">
                        <span className="text-xs text-gray-400">Uncategorized</span>
                        {onRecategorize && (
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              setRecategorizeModal({ isOpen: true, transaction: txn });
                            }}
                            className="text-xs px-2 py-1 text-blue-600 hover:bg-blue-50 rounded transition-colors"
                          >
                            Categorize
                          </button>
                        )}
                      </div>
                    )}
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Recategorize Modal */}
      {recategorizeModal.transaction && (
        <RecategorizeModal
          isOpen={recategorizeModal.isOpen}
          transaction={{
            id: recategorizeModal.transaction.transaction_id || '',
            merchant: recategorizeModal.transaction.merchant || 'Unknown',
            amount: recategorizeModal.transaction.amount,
            currentCategory: recategorizeModal.transaction.category?.name || 'Uncategorized',
            currentCategoryId: recategorizeModal.transaction.category?.id || '',
          }}
          categories={categories}
          onConfirm={async (newCategoryId, learnCorrection) => {
            if (onRecategorize && recategorizeModal.transaction) {
              setIsRecategorizing(true);
              try {
                await onRecategorize(
                  recategorizeModal.transaction.transaction_id || '',
                  newCategoryId,
                  learnCorrection
                );
                setRecategorizeModal({ isOpen: false });
              } finally {
                setIsRecategorizing(false);
              }
            }
          }}
          onCancel={() => setRecategorizeModal({ isOpen: false })}
          isLoading={isRecategorizing}
        />
      )}

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
