'use client';

import React, { useState } from 'react';
import { FileDropZone } from '@/components/FileDropZone';
import { BankSelector } from '@/components/BankSelector';
import { useAuth } from '@/contexts/AuthContext';
import { apiClient } from '@/services/api';

export default function UploadPage() {
  const { isAuthenticated } = useAuth();
  const [selectedFile, setSelectedFile] = useState<File | null>(null);
  const [selectedBank, setSelectedBank] = useState<string>('HDFC');
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [success, setSuccess] = useState<string | null>(null);

  const handleFileSelect = (file: File) => {
    setSelectedFile(file);
    setError(null);
    setSuccess(null);
  };

  const handleBankSelect = (bankCode: string) => {
    setSelectedBank(bankCode);
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();

    if (!selectedFile) {
      setError('Please select a file');
      return;
    }

    if (!selectedBank) {
      setError('Please select a bank');
      return;
    }

    setIsLoading(true);
    setError(null);

    try {
      // Create FormData for multipart upload
      const formData = new FormData();
      formData.append('file', selectedFile);
      formData.append('bank_code', selectedBank);

      // Upload the file
      const response = await apiClient.post('/statements/upload', formData, {
        headers: {
          'Content-Type': 'multipart/form-data',
        },
      });

      if (response.status === 202 || response.status === 200) {
        setSuccess(
          `Statement uploaded successfully! Statement ID: ${response.data.statement_id}`
        );
        setSelectedFile(null);
        setSelectedBank('HDFC');

        // Optionally redirect to preview page
        // setTimeout(() => {
        //   router.push(`/statements/${response.data.statement_id}/preview`);
        // }, 2000);
      }
    } catch (err: any) {
      const errorMessage =
        err.response?.data?.error ||
        err.response?.data?.message ||
        err.message ||
        'Failed to upload statement';
      setError(errorMessage);
    } finally {
      setIsLoading(false);
    }
  };

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="text-center">
          <h1 className="text-2xl font-bold text-gray-900 mb-4">
            Authentication Required
          </h1>
          <p className="text-gray-600">
            Please log in to upload bank statements.
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 py-12 px-4">
      <div className="max-w-2xl mx-auto">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-2">
            Upload Bank Statement
          </h1>
          <p className="text-gray-600">
            Upload your bank statement to extract transactions and begin your
            financial analysis
          </p>
        </div>

        {/* Form */}
        <form
          onSubmit={handleSubmit}
          className="bg-white rounded-lg shadow-xl p-8 space-y-8"
        >
          {/* Bank Selector */}
          <div>
            <BankSelector
              value={selectedBank}
              onChange={handleBankSelect}
              required
            />
          </div>

          {/* File Drop Zone */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-4">
              Statement File {<span className="text-red-500">*</span>}
            </label>
            <FileDropZone
              onFileSelect={handleFileSelect}
              acceptedFormats={['pdf', 'csv', 'xlsx']}
              maxSizeMB={50}
            />
            {selectedFile && (
              <div className="mt-4 p-3 bg-green-50 border border-green-200 rounded-lg">
                <p className="text-sm text-green-800">
                  ✓ File selected: <strong>{selectedFile.name}</strong> (
                  {(selectedFile.size / 1024).toFixed(2)} KB)
                </p>
              </div>
            )}
          </div>

          {/* Error Message */}
          {error && (
            <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
              <p className="text-sm text-red-800">
                <strong>Error:</strong> {error}
              </p>
            </div>
          )}

          {/* Success Message */}
          {success && (
            <div className="p-4 bg-green-50 border border-green-200 rounded-lg">
              <p className="text-sm text-green-800">
                <strong>Success:</strong> {success}
              </p>
            </div>
          )}

          {/* Submit Button */}
          <div className="flex gap-4">
            <button
              type="submit"
              disabled={!selectedFile || isLoading}
              className="flex-1 px-6 py-3 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
            >
              {isLoading ? 'Uploading...' : 'Upload Statement'}
            </button>
            <button
              type="button"
              onClick={() => {
                setSelectedFile(null);
                setError(null);
              }}
              className="px-6 py-3 bg-gray-200 text-gray-800 font-medium rounded-lg hover:bg-gray-300 transition-colors"
            >
              Clear
            </button>
          </div>

          {/* Helper Text */}
          <div className="p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="font-semibold text-blue-900 mb-2">📋 Requirements</h3>
            <ul className="text-sm text-blue-800 space-y-1">
              <li>✓ Supported formats: PDF, CSV, XLSX</li>
              <li>✓ Maximum file size: 50 MB</li>
              <li>✓ Ensure statement contains transaction data</li>
              <li>✓ Bank format must be recognized</li>
            </ul>
          </div>
        </form>
      </div>
    </div>
  );
}
