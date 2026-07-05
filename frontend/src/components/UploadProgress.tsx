'use client';

import React from 'react';

interface UploadProgressProps {
  isLoading: boolean;
  fileName?: string;
  progress?: number; // 0-100
  status?: 'uploading' | 'processing' | 'complete' | 'error';
  message?: string;
}

export function UploadProgress({
  isLoading,
  fileName,
  progress = 0,
  status = 'uploading',
  message,
}: UploadProgressProps) {
  if (!isLoading) {
    return null;
  }

  const statusLabels = {
    uploading: 'Uploading file...',
    processing: 'Processing statement...',
    complete: 'Complete!',
    error: 'Error',
  };

  const statusColors = {
    uploading: 'bg-blue-500',
    processing: 'bg-indigo-500',
    complete: 'bg-green-500',
    error: 'bg-red-500',
  };

  return (
    <div className="fixed inset-0 bg-black bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg shadow-2xl p-8 max-w-md w-full mx-4">
        {/* Header */}
        <div className="text-center mb-6">
          <h2 className="text-xl font-bold text-gray-900 mb-2">
            {statusLabels[status]}
          </h2>
          {fileName && (
            <p className="text-sm text-gray-600 truncate">
              {fileName}
            </p>
          )}
        </div>

        {/* Progress Bar */}
        <div className="mb-6">
          <div className="w-full bg-gray-200 rounded-full h-3 overflow-hidden">
            <div
              className={`h-full ${statusColors[status]} transition-all duration-500 ease-out`}
              style={{ width: `${progress}%` }}
            />
          </div>
          <p className="text-sm text-gray-600 text-center mt-2">
            {progress}%
          </p>
        </div>

        {/* Status Messages */}
        {status === 'uploading' && (
          <div className="text-center text-sm text-gray-600">
            <div className="flex items-center justify-center gap-2">
              <div className="animate-spin h-4 w-4 border-2 border-blue-500 border-t-transparent rounded-full"></div>
              <span>Please wait while we upload your file...</span>
            </div>
          </div>
        )}

        {status === 'processing' && (
          <div className="text-center text-sm text-gray-600">
            <div className="flex items-center justify-center gap-2">
              <div className="animate-spin h-4 w-4 border-2 border-indigo-500 border-t-transparent rounded-full"></div>
              <span>Extracting transactions from your statement...</span>
            </div>
          </div>
        )}

        {status === 'complete' && (
          <div className="text-center">
            <div className="inline-block p-3 bg-green-100 rounded-full mb-3">
              <svg className="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
              </svg>
            </div>
            <p className="text-green-700 font-semibold">Upload successful!</p>
            <p className="text-sm text-gray-600 mt-1">
              Redirecting to preview...
            </p>
          </div>
        )}

        {status === 'error' && (
          <div className="text-center">
            <div className="inline-block p-3 bg-red-100 rounded-full mb-3">
              <svg className="w-6 h-6 text-red-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
              </svg>
            </div>
            <p className="text-red-700 font-semibold">Upload failed</p>
            {message && (
              <p className="text-sm text-red-600 mt-2">{message}</p>
            )}
          </div>
        )}

        {/* Additional Message */}
        {message && status !== 'error' && (
          <div className="mt-4 p-3 bg-blue-50 border border-blue-200 rounded">
            <p className="text-sm text-blue-800">{message}</p>
          </div>
        )}
      </div>
    </div>
  );
}
