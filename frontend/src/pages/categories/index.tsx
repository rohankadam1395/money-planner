import React, { useState, useEffect } from 'react';
import { useAuth } from '@/contexts/AuthContext';
import Link from 'next/link';

interface CategoryStat {
  id: string;
  name: string;
  description: string;
  color: string;
  icon: string;
  totalSpent: number;
  transactionCount: number;
  averageTransaction: number;
}

interface CategoryDashboardProps {
  period?: string;
}

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

export default function CategoryDashboard({ period }: CategoryDashboardProps) {
  const { isAuthenticated } = useAuth();
  const [categories, setCategories] = useState<CategoryStat[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [selectedPeriod, setSelectedPeriod] = useState(period || '2026-07');

  const fetchCategories = React.useCallback(async () => {
    try {
      setLoading(true);
      const response = await fetch(
        `/api/v1/categories?include_stats=true&period=${selectedPeriod}`,
        {
          headers: {
            'Authorization': `Bearer ${localStorage.getItem('authToken')}`,
          },
        }
      );

      if (!response.ok) throw new Error('Failed to fetch categories');

      const data = await response.json();
      setCategories(data.categories || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load categories');
    } finally {
      setLoading(false);
    }
  }, [selectedPeriod]);

  useEffect(() => {
    if (!isAuthenticated) return;
    fetchCategories();
  }, [isAuthenticated, fetchCategories]);

  if (!isAuthenticated) {
    return (
      <div className="p-8 text-center">
        <p>Please log in to view your categories</p>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-slate-50 to-slate-100 p-8">
      <div className="max-w-7xl mx-auto">
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-slate-900 mb-2">Category Analytics</h1>
          <p className="text-slate-600">View your spending by category</p>
        </div>

        {/* Period Selector */}
        <div className="mb-8 flex items-center gap-4">
          <label className="text-sm font-medium text-slate-600">Period:</label>
          <input
            type="month"
            value={selectedPeriod}
            onChange={(e) => setSelectedPeriod(e.target.value)}
            className="px-4 py-2 border border-slate-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
        </div>

        {error && (
          <div className="mb-8 p-4 bg-red-100 border border-red-400 text-red-700 rounded-lg">
            {error}
          </div>
        )}

        {loading ? (
          <div className="text-center py-12">
            <p className="text-slate-600">Loading categories...</p>
          </div>
        ) : (
          <>
            {/* Category Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6 mb-12">
              {categories.map((cat) => (
                <Link href={`/categories/${cat.id}`} key={cat.id}>
                  <div
                    className="p-6 rounded-lg shadow-md hover:shadow-lg transition-shadow cursor-pointer bg-white border-l-4"
                    style={{ borderLeftColor: cat.color }}
                  >
                    <div className="flex items-center justify-between mb-4">
                      <h3 className="text-lg font-semibold text-slate-900">
                        {cat.name}
                      </h3>
                      <span className="text-3xl">{categoryIcons[cat.name] || '📊'}</span>
                    </div>

                    <div className="space-y-3">
                      <div>
                        <p className="text-sm text-slate-600">Total Spent</p>
                        <p className="text-2xl font-bold text-slate-900">
                          ₹{cat.totalSpent.toLocaleString('en-IN', {
                            minimumFractionDigits: 2,
                            maximumFractionDigits: 2,
                          })}
                        </p>
                      </div>

                      <div className="grid grid-cols-2 gap-4">
                        <div>
                          <p className="text-xs text-slate-600">Transactions</p>
                          <p className="text-lg font-semibold text-slate-900">
                            {cat.transactionCount}
                          </p>
                        </div>
                        <div>
                          <p className="text-xs text-slate-600">Average</p>
                          <p className="text-lg font-semibold text-slate-900">
                            ₹{cat.averageTransaction.toLocaleString('en-IN', {
                              maximumFractionDigits: 0,
                            })}
                          </p>
                        </div>
                      </div>
                    </div>

                    <div className="mt-4 text-sm text-blue-600 hover:text-blue-700 font-medium">
                      View Transactions →
                    </div>
                  </div>
                </Link>
              ))}
            </div>

            {/* Summary Stats */}
            <div className="bg-white rounded-lg shadow-md p-8">
              <h2 className="text-xl font-bold text-slate-900 mb-6">Summary</h2>
              <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                <div>
                  <p className="text-sm text-slate-600 mb-1">Total Spent</p>
                  <p className="text-3xl font-bold text-slate-900">
                    ₹{categories
                      .reduce((sum, cat) => sum + cat.totalSpent, 0)
                      .toLocaleString('en-IN', {
                        minimumFractionDigits: 2,
                        maximumFractionDigits: 2,
                      })}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-slate-600 mb-1">Total Transactions</p>
                  <p className="text-3xl font-bold text-slate-900">
                    {categories.reduce((sum, cat) => sum + cat.transactionCount, 0)}
                  </p>
                </div>
                <div>
                  <p className="text-sm text-slate-600 mb-1">Average Transaction</p>
                  <p className="text-3xl font-bold text-slate-900">
                    ₹{(
                      categories.reduce((sum, cat) => sum + cat.totalSpent, 0) /
                      Math.max(
                        categories.reduce((sum, cat) => sum + cat.transactionCount, 0),
                        1
                      )
                    ).toLocaleString('en-IN', { maximumFractionDigits: 0 })}
                  </p>
                </div>
              </div>
            </div>
          </>
        )}
      </div>
    </div>
  );
}
