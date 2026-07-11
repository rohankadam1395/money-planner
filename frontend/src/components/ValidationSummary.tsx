'use client';

import React, { useState } from 'react';
import { ValidationError } from '@/services/statementApi';

interface ValidationSummaryProps {
  totalRows: number;
  validTransactions: number;
  invalidTransactions: number;
  errors: ValidationError[];
  periodStart?: string;
  periodEnd?: string;
}

export function ValidationSummary({
  totalRows,
  validTransactions,
  invalidTransactions,
  errors,
  periodStart,
  periodEnd,
}: ValidationSummaryProps) {
  const [expandErrors, setExpandErrors] = useState(false);
  const validPercentage = totalRows > 0 ? Math.round((validTransactions / totalRows) * 100) : 0;
  const isHealthy = validPercentage >= 95;

  return (
    <div className="w-full space-y-4">
      {/* Summary Stats */}
      <div className="grid grid-cols-3 gap-4">
        <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
          <div className="text-xs text-blue-600 font-semibold uppercase">Total Rows</div>
          <div className="text-lg font-bold text-blue-900">{totalRows}</div>
        </div>
        <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
          <div className="text-xs text-green-600 font-semibold uppercase">Valid</div>
          <div className="text-lg font-bold text-green-900">{validTransactions}</div>
          <div className="text-xs text-green-600 mt-1">{validPercentage}%</div>
        </div>
        <div className={`p-4 ${invalidTransactions > 0 ? 'bg-red-50 border-red-200' : 'bg-gray-50 border-gray-200'} border rounded-lg`}>
          <div className={`text-xs font-semibold uppercase ${invalidTransactions > 0 ? 'text-red-600' : 'text-gray-600'}`}>
            Invalid
          </div>
          <div className={`text-lg font-bold ${invalidTransactions > 0 ? 'text-red-900' : 'text-gray-900'}`}>
            {invalidTransactions}
          </div>
        </div>
      </div>

      {/* Period Information */}
      {(periodStart || periodEnd) && (
        <div className="p-4 bg-indigo-50 border border-indigo-200 rounded-lg">
          <div className="text-sm text-indigo-600 font-semibold mb-2">Statement Period</div>
          <div className="text-indigo-900">
            {periodStart && periodEnd ? (
              `${new Date(periodStart).toLocaleDateString('en-IN')} to ${new Date(periodEnd).toLocaleDateString('en-IN')}`
            ) : (
              'Period dates could not be extracted'
            )}
          </div>
        </div>
      )}

      {/* Validation Status Banner */}
      <div
        className={`p-4 rounded-lg border ${
          isHealthy
            ? 'bg-green-50 border-green-200'
            : invalidTransactions === 0
              ? 'bg-blue-50 border-blue-200'
              : 'bg-yellow-50 border-yellow-200'
        }`}
      >
        <div
          className={`text-sm font-semibold ${
            isHealthy
              ? 'text-green-700'
              : invalidTransactions === 0
                ? 'text-blue-700'
                : 'text-yellow-700'
          }`}
        >
          {isHealthy
            ? '✓ Extraction Quality: Excellent (≥95%)'
            : invalidTransactions === 0
              ? '✓ All transactions are valid'
              : '⚠ Some transactions have validation issues'}
        </div>
        {invalidTransactions > 0 && (
          <p className="text-sm text-yellow-700 mt-2">
            Review the errors below to understand which transactions failed validation.
          </p>
        )}
      </div>

      {/* Detailed Error List */}
      {errors.length > 0 && (
        <div className="border border-red-200 rounded-lg overflow-hidden">
          <button
            onClick={() => setExpandErrors(!expandErrors)}
            className="w-full px-4 py-3 bg-red-50 hover:bg-red-100 transition-colors flex items-center justify-between"
          >
            <span className="font-semibold text-red-700">
              Validation Errors ({errors.length})
            </span>
            <span className="text-red-600">
              {expandErrors ? '▼' : '▶'}
            </span>
          </button>

          {expandErrors && (
            <div className="max-h-64 overflow-y-auto">
              {errors.slice(0, 20).map((error, idx) => (
                <div
                  key={idx}
                  className="px-4 py-3 border-t border-red-100 hover:bg-red-50"
                >
                  <div className="flex items-start gap-3">
                    <div className="text-xs font-semibold text-red-600 bg-red-100 px-2 py-1 rounded whitespace-nowrap mt-0.5">
                      Row {error.transaction_index}
                    </div>
                    <div className="flex-1">
                      <div className="text-sm font-semibold text-red-900">
                        {error.field}
                      </div>
                      <div className="text-sm text-red-700">{error.error}</div>
                    </div>
                  </div>
                </div>
              ))}
              {errors.length > 20 && (
                <div className="px-4 py-3 border-t border-red-100 text-center text-sm text-gray-600">
                  ... and {errors.length - 20} more errors
                </div>
              )}
            </div>
          )}
        </div>
      )}
    </div>
  );
}
