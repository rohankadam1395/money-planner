'use client';

import React, { useState } from 'react';
import { useRouter } from 'next/navigation';
import UploadPage from './UploadPage';
import PreviewPage from './PreviewPage';
import { UploadProgress } from '@/components/UploadProgress';
import { statementApi } from '@/services/statementApi';

type StatementsPageView = 'upload' | 'preview' | 'history';

interface UploadState {
  fileName?: string;
  statementId?: string;
  progress: number;
  status: 'idle' | 'uploading' | 'processing' | 'complete' | 'error';
  errorMessage?: string;
}

export default function StatementsPage() {
  const router = useRouter();
  const [currentView, setCurrentView] = useState<StatementsPageView>('upload');
  const [uploadState, setUploadState] = useState<UploadState>({
    progress: 0,
    status: 'idle',
  });

  const handleUploadStart = (fileName: string) => {
    setUploadState({
      fileName,
      progress: 10,
      status: 'uploading',
    });
  };

  const handleUploadProgress = (progress: number) => {
    setUploadState((prev) => ({
      ...prev,
      progress: Math.min(progress, 90), // Cap at 90% until processing complete
    }));
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

  // Render based on current view
  if (currentView === 'upload') {
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

  if (currentView === 'preview') {
    return <PreviewPage />;
  }

  // Default: upload
  return <UploadPage onUploadStart={handleUploadStart} onUploadSuccess={handleUploadSuccess} onUploadError={handleUploadError} />;
}
