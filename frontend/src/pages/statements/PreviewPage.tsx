import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '@/contexts/AuthContext';
import Navbar from '@/components/Navbar';
import { TransactionPreview } from '@/components/TransactionPreview';
import { ValidationSummary } from '@/components/ValidationSummary';
import {
  statementApi,
  PreviewResponse,
  Transaction,
} from '@/services/statementApi';

export default function PreviewPage() {
  const router = useRouter();
  const { isLoading: authLoading, isAuthenticated } = useAuth();
  const statementId = router.query.id as string | undefined;

  const [preview, setPreview] = useState<PreviewResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isConfirming, setIsConfirming] = useState(false);
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);
  const [processingProgress, setProcessingProgress] = useState<{ current: number; total: number } | null>(null);
  const [isEnhancingAI, setIsEnhancingAI] = useState(false);
  const [selectedForEnhance, setSelectedForEnhance] = useState<Set<string>>(new Set());
  const [isImported, setIsImported] = useState(false);
  const [successMessage, setSuccessMessage] = useState<string | null>(null);

  useEffect(() => {
    // Wait for auth to fully load before checking
    if (authLoading) {
      return;
    }

    // Check token in localStorage as well (in case AuthContext hasn't synced yet)
    const storedToken = localStorage.getItem('authToken');

    // Only redirect if auth is done loading and user is NOT authenticated
    if (!isAuthenticated && !storedToken) {
      router.push('/auth/login');
      return;
    }

    // Auth is ready and user is authenticated, fetch preview
    if (!statementId) {
      setError('No statement ID provided');
      setIsLoading(false);
      return;
    }

    const fetchPreview = async () => {
      try {
        setIsLoading(true);
        const token = localStorage.getItem('authToken');
        if (!token) {
          setError('Authentication token not found');
          router.push('/auth/login');
          return;
        }

        // Poll for preview with progress updates
        const maxAttempts = 30;
        const delayMs = 1000;
        let previewData: PreviewResponse | null = null;

        for (let attempt = 0; attempt < maxAttempts; attempt++) {
          try {
            previewData = await statementApi.getPreview(statementId);
            break;
          } catch (error) {
            // Show progress during polling
            setProcessingProgress({ current: attempt + 1, total: maxAttempts });
            if (attempt === maxAttempts - 1) {
              throw error;
            }
            await new Promise((resolve) => setTimeout(resolve, delayMs));
          }
        }

        if (!previewData) {
          throw new Error('Failed to fetch preview');
        }

        setPreview(previewData);
        setError(null);
      } catch (err: any) {
        const errorMessage = err.message || 'Failed to load statement preview';
        setError(errorMessage);
        setPreview(null);
      } finally {
        setIsLoading(false);
        setProcessingProgress(null);
      }
    };

    fetchPreview();
  }, [statementId, authLoading, isAuthenticated, router]);

  const handleConfirmImport = async () => {
    if (!statementId || !preview?.transactions) return;

    try {
      setIsConfirming(true);
      setError(null);

      // Send confirm with only explicitly categorized transactions
      // Uncategorized transactions will be LLM-categorized by the backend on confirm
      const txnsToConfirm = preview.transactions
        .filter((t) => t.category?.name && t.category.name !== 'Uncategorized')
        .map((t) => ({
          transaction_id: t.transaction_id,
          category_name: t.category.name,
          confidence: t.category.confidence || 0,
          method: t.category.method || 'none',
        }));

      console.log(`Confirming with ${txnsToConfirm.length} categorized transactions (${preview.transactions.length - txnsToConfirm.length} will be LLM-categorized on backend)`);

      const response = await statementApi.confirmImport({
        statement_id: statementId,
        confirmed: true,
        transactions: txnsToConfirm,
      });

      // Fetch fresh preview from DB to show saved categories
      const freshPreview = await statementApi.getPreview(statementId);
      setPreview(freshPreview);
      setIsImported(true);
      setSuccessMessage(
        `✓ Successfully imported ${response.transactions_imported} transactions. Categories have been saved to the database.`
      );

      // Auto-clear success message after 5 seconds
      setTimeout(() => setSuccessMessage(null), 5000);
    } catch (err: any) {
      setError(err.message || 'Failed to confirm import');
    } finally {
      setIsConfirming(false);
    }
  };

  const handleCancel = () => {
    router.back();
  };

  const handleEnhanceWithAI = async () => {
    if (!preview?.transactions || selectedForEnhance.size === 0) return;

    try {
      setIsEnhancingAI(true);

      // Get only selected transactions
      const txnsToEnhance = preview.transactions.filter((t) =>
        selectedForEnhance.has(t.transaction_id)
      );

      // Call categorization endpoint with LLM for selected transactions
      const enhanced = await statementApi.categorizeTransactions(txnsToEnhance);

      // Update preview by merging enhanced results
      setPreview((current) => {
        if (!current) return current;
        const enhancedMap = new Map(
          enhanced.map((t) => [t.transaction_id, t])
        );
        return {
          ...current,
          transactions: current.transactions.map((t) =>
            enhancedMap.has(t.transaction_id) ? enhancedMap.get(t.transaction_id)! : t
          ),
        };
      });

      // Clear selection after successful enhancement
      setSelectedForEnhance(new Set());
    } catch (err: any) {
      setError(err.message || 'Failed to enhance with AI');
    } finally {
      setIsEnhancingAI(false);
    }
  };

  const toggleTransactionSelection = (transactionId: string) => {
    const newSelected = new Set(selectedForEnhance);
    if (newSelected.has(transactionId)) {
      newSelected.delete(transactionId);
    } else {
      newSelected.add(transactionId);
    }
    setSelectedForEnhance(newSelected);
  };

  const toggleSelectAll = () => {
    if (!preview?.transactions) return;
    if (selectedForEnhance.size === preview.transactions.length) {
      setSelectedForEnhance(new Set());
    } else {
      const allIds = new Set(preview.transactions.map((t) => t.transaction_id));
      setSelectedForEnhance(allIds);
    }
  };

  // Auth initialization loading state
  if (authLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-xl p-8 text-center">
          <div className="inline-block p-3 bg-indigo-100 rounded-full mb-4">
            <svg className="w-8 h-8 text-indigo-600 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900 mb-2">Authenticating</h2>
          <p className="text-gray-600 text-sm">Verifying your credentials...</p>
        </div>
      </div>
    );
  }

  // Loading state
  if (isLoading) {
    const progressPercent = processingProgress
      ? Math.round((processingProgress.current / processingProgress.total) * 100)
      : 0;

    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-xl p-8 text-center">
          <div className="inline-block p-3 bg-blue-100 rounded-full mb-4">
            <svg className="w-8 h-8 text-blue-600 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900 mb-2">Processing Statement</h2>
          {processingProgress ? (
            <>
              <p className="text-gray-600 text-sm mb-4">
                Extracting & categorizing transactions...
                <br />
                <span className="font-semibold text-indigo-600">
                  {processingProgress.current} of {processingProgress.total}
                </span>
              </p>
              <div className="w-full bg-gray-200 rounded-full h-2 mb-2 overflow-hidden">
                <div
                  className="bg-gradient-to-r from-indigo-500 to-blue-600 h-full transition-all duration-300"
                  style={{ width: `${progressPercent}%` }}
                ></div>
              </div>
              <p className="text-gray-500 text-xs">{progressPercent}% complete</p>
            </>
          ) : (
            <p className="text-gray-600 text-sm">Extracting transactions from your statement...</p>
          )}
        </div>
      </div>
    );
  }

  // Error state
  if (error && !isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-lg shadow-xl p-8">
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg mb-6">
              <p className="text-red-800">
                <strong>Error:</strong> {error}
              </p>
            </div>
            <button
              onClick={handleCancel}
              className="px-6 py-2 bg-gray-600 text-white rounded-lg hover:bg-gray-700 transition-colors"
            >
              Go Back
            </button>
          </div>
        </div>
      </div>
    );
  }

  // Loading state
  if (isLoading || !preview) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
        <div className="max-w-4xl mx-auto">
          <div className="bg-white rounded-lg shadow-xl p-8">
            <div className="flex items-center justify-center gap-3">
              <div className="animate-spin h-5 w-5 border-2 border-blue-500 border-t-transparent rounded-full"></div>
              <span className="text-gray-700">Processing statement...</span>
            </div>
          </div>
        </div>
      </div>
    );
  }

  return (
    <>
      <Navbar />
      <div className="min-h-screen bg-gradient-to-br from-violet-50 via-purple-50 to-indigo-100 py-8 px-4">
        <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-4xl font-bold bg-gradient-to-r from-violet-600 via-purple-600 to-indigo-600 bg-clip-text text-transparent mb-1">
            Review & Confirm
          </h1>
          <p className="text-sm text-gray-600">
            Review extracted transactions below. Click Confirm to import.
          </p>
        </div>

        {/* Main Content */}
        <div className="bg-white/80 backdrop-blur-xl rounded-2xl shadow-2xl p-8 space-y-6 border border-white/40">
          {/* Success Message */}
          {successMessage && (
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-green-800 font-semibold">{successMessage}</p>
            </div>
          )}

          {/* Validation Summary */}
          <section>
            <h2 className="text-lg font-bold text-gray-900 mb-3">Validation Summary</h2>
            {preview?.validation_summary ? (
              <ValidationSummary
                totalRows={preview.validation_summary.total_rows || 0}
                validTransactions={preview.validation_summary.valid_transactions || 0}
                invalidTransactions={preview.validation_summary.invalid_transactions || 0}
                errors={preview.validation_summary.errors || []}
                periodStart={preview.validation_summary.period_start}
                periodEnd={preview.validation_summary.period_end}
              />
            ) : (
              <p className="text-gray-500">No validation data available</p>
            )}
          </section>

          {/* Transaction Details */}
          {selectedTransaction && (
            <section className="bg-blue-50 border border-blue-200 rounded-lg p-4">
              <div className="flex items-start justify-between">
                <div>
                  <h3 className="text-lg font-semibold text-blue-900">
                    Transaction Details
                  </h3>
                  <div className="mt-3 space-y-2 text-sm text-blue-800">
                    <p><strong>Date:</strong> {new Date(selectedTransaction.transaction_date).toLocaleDateString('en-IN')}</p>
                    <p><strong>Description:</strong> {selectedTransaction.description || selectedTransaction.merchant}</p>
                    <p><strong>Amount:</strong> ₹{selectedTransaction.amount.toFixed(2)}</p>
                    <p><strong>Type:</strong> {selectedTransaction.type}</p>
                  </div>
                </div>
                <button
                  onClick={() => setSelectedTransaction(null)}
                  className="text-blue-600 hover:text-blue-800 font-semibold"
                >
                  Close
                </button>
              </div>
            </section>
          )}

          {/* Transactions Table */}
          <section>
            <div className="flex items-center justify-between mb-4">
              <h2 className="text-lg font-bold text-gray-900">
                Extracted Transactions ({preview?.transactions?.length || 0})
              </h2>
              <button
                onClick={handleEnhanceWithAI}
                disabled={isEnhancingAI || selectedForEnhance.size === 0}
                className={`flex items-center gap-2 px-4 py-2 rounded-lg font-semibold transition-all ${
                  isEnhancingAI
                    ? 'bg-indigo-100 text-indigo-600 cursor-wait'
                    : selectedForEnhance.size > 0
                      ? 'bg-indigo-600 text-white hover:bg-indigo-700 hover:shadow-lg'
                      : 'bg-gray-300 text-gray-500'
                } disabled:opacity-50 disabled:cursor-not-allowed`}
              >
                {isEnhancingAI ? (
                  <>
                    <div className="w-4 h-4 border-2 border-indigo-300 border-t-indigo-600 rounded-full animate-spin"></div>
                    Enhancing {selectedForEnhance.size}...
                  </>
                ) : (
                  <>
                    <span>✨ Enhance with AI ({selectedForEnhance.size})</span>
                  </>
                )}
              </button>
            </div>
            <TransactionPreview
              transactions={preview?.transactions || []}
              pageSize={15}
              onSelectTransaction={setSelectedTransaction}
              selectedForEnhance={selectedForEnhance}
              onToggleSelect={toggleTransactionSelection}
              onToggleSelectAll={toggleSelectAll}
            />
          </section>

          {/* Action Buttons */}
          <div className="flex gap-4 pt-6 border-t-2 border-purple-100">
            {isImported ? (
              <>
                <button
                  onClick={() => router.push('/statements')}
                  className="flex-1 px-6 py-3 bg-gradient-to-r from-blue-500 to-blue-600 text-white rounded-xl hover:shadow-lg hover:shadow-blue-400/50 transition-all font-semibold transform hover:scale-105"
                >
                  → Go to Statements
                </button>
                <button
                  onClick={handleCancel}
                  className="flex-1 px-6 py-3 bg-gray-200 text-gray-800 rounded-xl hover:bg-gray-300 transition-all font-semibold transform hover:scale-105"
                >
                  Go Back
                </button>
              </>
            ) : (
              <>
                <button
                  onClick={handleConfirmImport}
                  disabled={isConfirming || (preview?.validation_summary?.invalid_transactions || 0) > 0}
                  className={`flex-1 px-6 py-3 rounded-xl font-semibold text-white transition-all transform ${
                    (preview?.validation_summary?.invalid_transactions || 0) > 0
                      ? 'bg-gray-400 cursor-not-allowed'
                      : 'bg-gradient-to-r from-emerald-500 to-teal-500 hover:shadow-lg hover:shadow-emerald-400/50 hover:scale-105'
                  }`}
                >
                  {isConfirming ? 'Confirming...' : '✓ Confirm & Import'}
                </button>
                <button
                  onClick={handleCancel}
                  disabled={isConfirming}
                  className="flex-1 px-6 py-3 bg-gray-200 text-gray-800 rounded-xl hover:bg-gray-300 transition-all font-semibold disabled:opacity-50 transform hover:scale-105"
                >
                  Cancel
                </button>
              </>
            )}
          </div>

          {(preview?.validation_summary?.invalid_transactions || 0) > 0 && (
            <div className="p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
              <p className="text-yellow-800 text-sm">
                <strong>Note:</strong> Please fix validation errors before importing.
                All transactions must be valid for import.
              </p>
            </div>
          )}
        </div>
      </div>
    </div>
    </>
  );
}
