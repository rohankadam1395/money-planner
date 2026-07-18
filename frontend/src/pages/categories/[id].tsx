import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '@/contexts/AuthContext';
import RecategorizeModal, { Category as ModalCategory } from '@/components/RecategorizeModal';
import Navbar from '@/components/Navbar';

interface CategoryTransaction {
  transaction_id: string;
  date: string;
  merchant: string;
  amount: number;
  method: string;
  llm_provider?: string;
  confidence: number;
}

interface Category {
  id: string;
  name: string;
  description: string;
  color: string;
  icon: string;
  totalSpent: number;
  transactionCount: number;
}

export default function CategoryDetail() {
  const router = useRouter();
  const { id } = router.query;
  const { isAuthenticated } = useAuth();
  const [category, setCategory] = useState<Category | null>(null);
  const [transactions, setTransactions] = useState<CategoryTransaction[]>([]);
  const [allCategories, setAllCategories] = useState<ModalCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [recategorizeModal, setRecategorizeModal] = useState<{
    isOpen: boolean;
    transaction?: CategoryTransaction;
  }>({ isOpen: false });

  const fetchCategoryDetail = React.useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(`/api/v1/categories/${id}/transactions`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch category details');

      const data = await response.json();
      setCategory({
        id: data.category_id,
        name: data.category_name,
        description: '',
        color: '#4ECDC4',
        icon: '📊',
        totalSpent: data.total_spent,
        transactionCount: data.total,
      });
      setTransactions(data.transactions || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load category');
    } finally {
      setLoading(false);
    }
  }, [id]);

  const fetchAllCategories = React.useCallback(async () => {
    try {
      const response = await fetch('/api/v1/categories', {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
        },
      });

      if (!response.ok) throw new Error('Failed to fetch categories');

      const data = await response.json();
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
      };

      const formattedCategories: ModalCategory[] = (data.categories || []).map(
        (cat: any) => ({
          id: cat.id,
          name: cat.name,
          color: cat.color || '#999',
          icon: categoryIcons[cat.name] || '📊',
          description: cat.description,
        })
      );

      setAllCategories(formattedCategories);
    } catch (err) {
      console.error('Failed to fetch categories:', err);
    }
  }, []);

  useEffect(() => {
    if (!isAuthenticated || !id) return;
    fetchCategoryDetail();
    fetchAllCategories();
  }, [isAuthenticated, id, fetchCategoryDetail, fetchAllCategories]);

  const handleRecategorize = async (
    transactionId: string,
    newCategoryId: string,
    learnCorrection: boolean
  ) => {
    try {
      const response = await fetch(`/api/v1/transactions/${transactionId}/recategorize`, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          transaction_id: transactionId,
          new_category_id: newCategoryId,
          learn_correction: learnCorrection,
        }),
      });

      if (!response.ok) throw new Error('Failed to recategorize transaction');

      setRecategorizeModal({ isOpen: false });
      // Refresh the category detail
      await fetchCategoryDetail();
    } catch (err) {
      alert(err instanceof Error ? err.message : 'Failed to recategorize');
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="p-8 text-center">
        <p>Please log in to view category details</p>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-8 flex items-center justify-center">
        <p className="text-slate-600">Loading category details...</p>
      </div>
    );
  }

  if (error || !category) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-8">
        <div className="max-w-7xl mx-auto">
          <div className="p-4 bg-red-100 border border-red-400 text-red-700 rounded-lg">
            {error || 'Category not found'}
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      <Navbar />
      <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-8">
        <div className="max-w-7xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <a
            onClick={() => router.back()}
            className="text-blue-600 hover:text-blue-700 mb-4 inline-block cursor-pointer"
          >
            ← Back
          </a>
          <h1 className="text-4xl font-bold text-slate-900 mb-2">
            {category.icon} {category.name}
          </h1>
          <p className="text-slate-600">{category.description}</p>
        </div>

        {/* Stats */}
        <div className="bg-white rounded-lg shadow-md p-8 mb-8">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
            <div>
              <p className="text-sm text-slate-600 mb-1">Total Spent</p>
              <p className="text-3xl font-bold" style={{ color: category.color }}>
                ₹{category.totalSpent.toLocaleString('en-IN', {
                  minimumFractionDigits: 2,
                  maximumFractionDigits: 2,
                })}
              </p>
            </div>
            <div>
              <p className="text-sm text-slate-600 mb-1">Transactions</p>
              <p className="text-3xl font-bold text-slate-900">
                {category.transactionCount}
              </p>
            </div>
            <div>
              <p className="text-sm text-slate-600 mb-1">Average</p>
              <p className="text-3xl font-bold text-slate-900">
                ₹{(category.totalSpent / Math.max(category.transactionCount, 1)).toLocaleString(
                  'en-IN',
                  { maximumFractionDigits: 0 }
                )}
              </p>
            </div>
          </div>
        </div>

        {/* Transactions Table */}
        <div className="bg-white rounded-lg shadow-md overflow-hidden">
          <div className="px-6 py-4 border-b border-slate-200">
            <h2 className="text-xl font-bold text-slate-900">Transactions</h2>
          </div>

          {transactions.length === 0 ? (
            <div className="p-8 text-center text-slate-600">
              No transactions in this category
            </div>
          ) : (
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-slate-50 border-b border-slate-200">
                  <tr>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900">
                      Date
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900">
                      Merchant
                    </th>
                    <th className="px-6 py-3 text-right text-sm font-semibold text-slate-900">
                      Amount
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900">
                      Method
                    </th>
                    <th className="px-6 py-3 text-right text-sm font-semibold text-slate-900">
                      Confidence
                    </th>
                    <th className="px-6 py-3 text-left text-sm font-semibold text-slate-900">
                      Action
                    </th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-slate-200">
                  {transactions.map((txn) => (
                    <tr
                      key={txn.transaction_id}
                      className="hover:bg-slate-50 transition-colors"
                    >
                      <td className="px-6 py-4 text-sm text-slate-900">
                        {new Date(txn.date).toLocaleDateString()}
                      </td>
                      <td className="px-6 py-4 text-sm text-slate-900 font-medium">
                        {txn.merchant}
                      </td>
                      <td className="px-6 py-4 text-right text-sm text-slate-900">
                        ₹{txn.amount.toLocaleString('en-IN', {
                          minimumFractionDigits: 2,
                          maximumFractionDigits: 2,
                        })}
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <span
                          className="px-3 py-1 rounded-full text-xs font-medium"
                          style={{
                            backgroundColor: `${category.color}20`,
                            color: category.color,
                          }}
                        >
                          {txn.method}
                          {txn.llm_provider && ` (${txn.llm_provider})`}
                        </span>
                      </td>
                      <td className="px-6 py-4 text-right text-sm text-slate-900">
                        {(txn.confidence * 100).toFixed(0)}%
                      </td>
                      <td className="px-6 py-4 text-sm">
                        <button
                          onClick={() =>
                            setRecategorizeModal({ isOpen: true, transaction: txn })
                          }
                          className="text-blue-600 hover:text-blue-700 font-medium"
                        >
                          Recategorize
                        </button>
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          )}
        </div>

        {/* Recategorize Modal */}
        {recategorizeModal.transaction && (
          <RecategorizeModal
            isOpen={recategorizeModal.isOpen}
            transaction={{
              id: recategorizeModal.transaction.transaction_id,
              merchant: recategorizeModal.transaction.merchant,
              amount: recategorizeModal.transaction.amount,
              currentCategory: category.name,
              currentCategoryId: category.id,
            }}
            categories={allCategories}
            onConfirm={(categoryId, learnCorrection) =>
              handleRecategorize(
                recategorizeModal.transaction!.transaction_id,
                categoryId,
                learnCorrection
              )
            }
            onCancel={() => setRecategorizeModal({ isOpen: false })}
          />
        )}
        </div>
      </div>
    </>
  );
}
