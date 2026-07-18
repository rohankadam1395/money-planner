import React, { useState } from 'react';

export interface Category {
  id: string;
  name: string;
  color: string;
  icon: string;
  description?: string;
}

interface RecategorizeModalProps {
  isOpen: boolean;
  transaction: {
    id: string;
    merchant: string;
    amount: number;
    currentCategory: string;
    currentCategoryId: string;
  };
  categories: Category[];
  onConfirm: (newCategoryId: string, learnCorrection: boolean) => void;
  onCancel: () => void;
  isLoading?: boolean;
}

export const RecategorizeModal: React.FC<RecategorizeModalProps> = ({
  isOpen,
  transaction,
  categories,
  onConfirm,
  onCancel,
  isLoading = false,
}) => {
  const [selectedCategoryId, setSelectedCategoryId] = useState(transaction.currentCategoryId);
  const [learnCorrection, setLearnCorrection] = useState(false);

  if (!isOpen) return null;

  const handleConfirm = () => {
    if (selectedCategoryId && selectedCategoryId !== transaction.currentCategoryId) {
      onConfirm(selectedCategoryId, learnCorrection);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black bg-opacity-50">
      <div className="w-full max-w-md rounded-lg bg-white p-6 shadow-xl">
        <div className="mb-6">
          <h2 className="text-2xl font-bold text-gray-900">Recategorize Transaction</h2>
          <p className="mt-2 text-gray-600">
            {transaction.merchant} - ₹{transaction.amount.toFixed(2)}
          </p>
        </div>

        <div className="mb-6">
          <label className="block text-sm font-medium text-gray-700 mb-3">
            Current Category: <span className="font-semibold text-gray-900">{transaction.currentCategory}</span>
          </label>

          <label className="block text-sm font-medium text-gray-700 mb-3">
            New Category
          </label>
          <div className="grid grid-cols-2 gap-2 max-h-64 overflow-y-auto">
            {categories.map(category => (
              <button
                key={category.id}
                onClick={() => setSelectedCategoryId(category.id)}
                className={`p-3 rounded-lg text-left transition-all ${
                  selectedCategoryId === category.id
                    ? 'ring-2 ring-offset-2 ring-blue-500'
                    : 'hover:bg-gray-50'
                }`}
                style={{
                  backgroundColor: selectedCategoryId === category.id ? `${category.color}20` : 'transparent',
                  borderLeft: `3px solid ${category.color}`,
                }}
              >
                <div className="flex items-center gap-2">
                  <span className="text-lg">{category.icon}</span>
                  <div>
                    <div className="font-medium text-gray-900 text-sm">{category.name}</div>
                    {category.description && (
                      <div className="text-xs text-gray-500">{category.description}</div>
                    )}
                  </div>
                </div>
              </button>
            ))}
          </div>
        </div>

        <div className="mb-6">
          <label className="flex items-center gap-3">
            <input
              type="checkbox"
              checked={learnCorrection}
              onChange={(e) => setLearnCorrection(e.target.checked)}
              disabled={isLoading}
              className="w-4 h-4 rounded border-gray-300"
            />
            <span className="text-sm text-gray-700">
              Learn this correction for future transactions
            </span>
          </label>
        </div>

        <div className="flex gap-3 justify-end">
          <button
            onClick={onCancel}
            disabled={isLoading}
            className="px-4 py-2 text-gray-700 bg-gray-100 hover:bg-gray-200 rounded-lg font-medium transition-colors disabled:opacity-50"
          >
            Cancel
          </button>
          <button
            onClick={handleConfirm}
            disabled={isLoading || selectedCategoryId === transaction.currentCategoryId}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg font-medium transition-colors disabled:opacity-50"
          >
            {isLoading ? 'Saving...' : 'Confirm'}
          </button>
        </div>
      </div>
    </div>
  );
};

export default RecategorizeModal;
