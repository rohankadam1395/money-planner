import React, { useState, useEffect } from 'react';
import { useRouter } from 'next/router';
import Link from 'next/link';
import UploadPage from './UploadPage';
import { UploadProgress } from '@/components/UploadProgress';

interface UploadState {
  fileName?: string;
  statementId?: string;
  progress: number;
  status: 'idle' | 'uploading' | 'processing' | 'complete' | 'error';
  errorMessage?: string;
}

export default function StatementsPage() {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [loading, setLoading] = useState(true);
  const [uploadState, setUploadState] = useState<UploadState>({
    progress: 0,
    status: 'idle',
  });

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      router.push('/auth/login');
    } else {
      setIsAuthenticated(true);
      setLoading(false);
    }
  }, [router]);

  const handleUploadStart = (fileName: string) => {
    setUploadState({
      fileName,
      progress: 10,
      status: 'uploading',
    });
  };

  const handleUploadSuccess = async (statementId: string) => {
    setUploadState((prev) => ({
      ...prev,
      statementId,
      progress: 95,
      status: 'processing',
    }));

    // Give a brief moment for visual feedback before redirect
    setTimeout(() => {
      setUploadState((prev) => ({
        ...prev,
        progress: 100,
        status: 'complete',
      }));

      // Redirect to preview page
      setTimeout(() => {
        router.push(`/statements/preview?id=${statementId}`);
        resetUploadState();
      }, 1500);
    }, 500);
  };

  const handleUploadError = (errorMessage: string) => {
    setUploadState((prev) => ({
      ...prev,
      progress: 0,
      status: 'error',
      errorMessage,
    }));
  };

  const resetUploadState = () => {
    setUploadState({
      progress: 0,
      status: 'idle',
    });
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading...</p>
        </div>
      </div>
    );
  }

  if (!isAuthenticated) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <h1 className="text-3xl font-bold text-gray-900 mb-4">Please Sign In</h1>
          <p className="text-gray-600 mb-6">You need to be authenticated to upload statements</p>
          <Link
            href="/auth/login"
            className="bg-indigo-600 text-white px-8 py-3 rounded-lg hover:bg-indigo-700 transition inline-block"
          >
            Sign In
          </Link>
        </div>
      </div>
    );
  }

  return (
    <>
      <UploadPage onUploadStart={handleUploadStart} onUploadSuccess={handleUploadSuccess} onUploadError={handleUploadError} />
      <UploadProgress
        isLoading={uploadState.status !== 'idle'}
        fileName={uploadState.fileName}
        progress={uploadState.progress}
        status={uploadState.status as 'uploading' | 'processing' | 'complete' | 'error'}
        message={uploadState.errorMessage || (uploadState.status === 'processing' ? 'Extracting transactions from your statement...' : undefined)}
      />
    </>
  );
}
