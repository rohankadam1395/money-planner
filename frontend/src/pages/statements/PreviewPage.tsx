'use client';

import React, { useEffect, useState } from 'react';
import { useSearchParams, useRouter } from 'next/navigation';
import { TransactionPreview } from '@/components/TransactionPreview';
import { ValidationSummary } from '@/components/ValidationSummary';
import {
  statementApi,
  pollStatementPreview,
  PreviewResponse,
  Transaction,
} from '@/services/statementApi';

export default function PreviewPage() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const statementId = searchParams.get('id');

  const [preview, setPreview] = useState<PreviewResponse | null>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isConfirming, setIsConfirming] = useState(false);
  const [selectedTransaction, setSelectedTransaction] = useState<Transaction | null>(null);

  useEffect(() => {
    const fetchPreview = async () => {
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
  }, [statementId]);

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
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Review Statement Preview
          </h1>
          <p className="text-gray-600">
            Review extracted transactions below. Click Confirm to import.
          </p>
        </div>

        {/* Main Content */}
        <div className="bg-white rounded-lg shadow-xl p-8 space-y-8">
          {/* Validation Summary */}
          <section>
            <h2 className="text-2xl font-bold text-gray-900 mb-4">Validation Summary</h2>
            <ValidationSummary
              totalRows={preview.validation_summary.total_rows}
              validTransactions={preview.validation_summary.valid_transactions}
              invalidTransactions={preview.validation_summary.invalid_transactions}
              errors={preview.validation_summary.errors}
              periodStart={preview.validation_summary.period_start}
              periodEnd={preview.validation_summary.period_end}
            />
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
            <h2 className="text-2xl font-bold text-gray-900 mb-4">
              Extracted Transactions ({preview.transactions.length})
            </h2>
            <TransactionPreview
              transactions={preview.transactions}
              pageSize={15}
              onSelectTransaction={setSelectedTransaction}
            />
          </section>

          {/* Action Buttons */}
          <div className="flex gap-4 pt-6 border-t">
            <button
              onClick={handleConfirmImport}
              disabled={isConfirming || preview.validation_summary.invalid_transactions > 0}
              className={`flex-1 px-6 py-3 rounded-lg font-medium text-white transition-colors ${
                preview.validation_summary.invalid_transactions > 0
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

          {preview.validation_summary.invalid_transactions > 0 && (
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
