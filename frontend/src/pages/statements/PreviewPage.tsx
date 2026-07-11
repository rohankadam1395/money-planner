import React, { useEffect, useState } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '@/contexts/AuthContext';
import { TransactionPreview } from '@/components/TransactionPreview';
import { ValidationSummary } from '@/components/ValidationSummary';
import {
  statementApi,
  pollStatementPreview,
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

  useEffect(() => {
    // Check auth first
    if (!authLoading && !isAuthenticated) {
      router.push('/auth/login');
      return;
    }

    const fetchPreview = async () => {
      // Wait for auth to load
      if (authLoading) {
        return;
      }

      if (!statementId) {
        setError('No statement ID provided');
        setIsLoading(false);
        return;
      }

      try {
        setIsLoading(true);
        // Poll for preview with up to 30 seconds timeout
        const previewData = await pollStatementPreview(statementId, 30, 1000);
        setPreview(previewData);
        setError(null);
      } catch (err: any) {
        const errorMessage = err.message || 'Failed to load statement preview';
        setError(errorMessage);
        setPreview(null);
      } finally {
        setIsLoading(false);
      }
    };

    fetchPreview();
  }, [statementId, authLoading, isAuthenticated, router]);

  const handleConfirmImport = async () => {
    if (!statementId) return;

    try {
      setIsConfirming(true);
      const response = await statementApi.confirmImport({
        statement_id: statementId,
        confirmed: true,
      });

      // Show success message and redirect to success page or statements list
      alert(`✓ Successfully imported ${response.transactions_imported} transactions`);
      router.push('/statements');
    } catch (err: any) {
      setError(err.message || 'Failed to confirm import');
    } finally {
      setIsConfirming(false);
    }
  };

  const handleCancel = () => {
    router.back();
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4 flex items-center justify-center">
        <div className="max-w-md w-full bg-white rounded-lg shadow-xl p-8 text-center">
          <div className="inline-block p-3 bg-blue-100 rounded-full mb-4">
            <svg className="w-8 h-8 text-blue-600 animate-spin" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 4v5h.582m15.356 2A8.001 8.001 0 004.582 9m0 0H9m11 11v-5h-.581m0 0a8.003 8.003 0 01-15.357-2m15.357 2H15" />
            </svg>
          </div>
          <h2 className="text-xl font-bold text-gray-900 mb-2">Processing Statement</h2>
          <p className="text-gray-600 text-sm">Extracting transactions from your statement...</p>
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
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-8 px-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-6">
          <h1 className="text-2xl font-bold text-gray-900 mb-1">
            Review Statement Preview
          </h1>
          <p className="text-sm text-gray-600">
            Review extracted transactions below. Click Confirm to import.
          </p>
        </div>

        {/* Main Content */}
        <div className="bg-white rounded-lg shadow-xl p-6 space-y-6">
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
                    <p><strong>Date:</strong> {new Date(selectedTransaction.date).toLocaleDateString('en-IN')}</p>
                    <p><strong>Description:</strong> {selectedTransaction.description}</p>
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
            <h2 className="text-lg font-bold text-gray-900 mb-3">
              Extracted Transactions ({preview?.transactions?.length || 0})
            </h2>
            <TransactionPreview
              transactions={preview?.transactions || []}
              pageSize={15}
              onSelectTransaction={setSelectedTransaction}
            />
          </section>

          {/* Action Buttons */}
          <div className="flex gap-4 pt-6 border-t">
            <button
              onClick={handleConfirmImport}
              disabled={isConfirming || (preview?.validation_summary?.invalid_transactions || 0) > 0}
              className={`flex-1 px-6 py-3 rounded-lg font-medium text-white transition-colors ${
                (preview?.validation_summary?.invalid_transactions || 0) > 0
                  ? 'bg-gray-400 cursor-not-allowed'
                  : 'bg-green-600 hover:bg-green-700'
              }`}
            >
              {isConfirming ? 'Confirming...' : 'Confirm & Import'}
            </button>
            <button
              onClick={handleCancel}
              disabled={isConfirming}
              className="flex-1 px-6 py-3 bg-gray-200 text-gray-800 rounded-lg hover:bg-gray-300 transition-colors font-medium disabled:opacity-50"
            >
              Cancel
            </button>
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
  );
}
