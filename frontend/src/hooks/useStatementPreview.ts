import { useState, useEffect, useCallback } from 'react';
import { statementApi, PreviewResponse } from '@/services/statementApi';

interface UseStatementPreviewOptions {
  autoFetch?: boolean;
  pollInterval?: number;
  maxAttempts?: number;
}

interface UseStatementPreviewResult {
  preview: PreviewResponse | null;
  isLoading: boolean;
  error: string | null;
  refetch: () => Promise<void>;
  reset: () => void;
}

export function useStatementPreview(
  statementId: string | null,
  options: UseStatementPreviewOptions = {}
): UseStatementPreviewResult {
  const { autoFetch = true, pollInterval = 1000, maxAttempts = 30 } = options;

  const [preview, setPreview] = useState<PreviewResponse | null>(null);
  const [isLoading, setIsLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const fetchPreview = useCallback(async () => {
    if (!statementId) {
      setError('No statement ID provided');
      return;
    }

    let attempts = 0;
    const maxRetries = maxAttempts;

    const pollPreview = async (): Promise<void> => {
      try {
        const data = await statementApi.getPreview(statementId);
        setPreview(data);
        setError(null);
        setIsLoading(false);
      } catch (err: any) {
        attempts++;

        if (attempts < maxRetries) {
          // Continue polling
          await new Promise((resolve) => setTimeout(resolve, pollInterval));
          return pollPreview();
        } else {
          // Max retries exceeded
          const errorMessage = err.message || 'Failed to fetch statement preview';
          setError(errorMessage);
          setIsLoading(false);
        }
      }
    };

    setIsLoading(true);
    await pollPreview();
  }, [statementId, pollInterval, maxAttempts]);

  useEffect(() => {
    if (autoFetch && statementId) {
      fetchPreview();
    }
  }, [statementId, autoFetch, fetchPreview]);

  const reset = useCallback(() => {
    setPreview(null);
    setError(null);
    setIsLoading(false);
  }, []);

  return {
    preview,
    isLoading,
    error,
    refetch: fetchPreview,
    reset,
  };
}
