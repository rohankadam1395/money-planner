import { useEffect } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '@/contexts/AuthContext';

export default function Home() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useAuth();

  useEffect(() => {
    if (!isLoading) {
      if (isAuthenticated) {
        router.push('/statements/list');
      } else {
        router.push('/auth/login');
      }
    }
  }, [isAuthenticated, isLoading, router]);

  return (
    <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-indigo-100 via-purple-50 to-pink-100">
      <div className="text-center">
        <h1 className="text-6xl font-bold bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 bg-clip-text text-transparent mb-2">💰 Money Planner</h1>
        <p className="text-gray-600 mt-4 text-lg">Loading your dashboard...</p>
        <div className="mt-6 inline-block">
          <div className="animate-pulse">
            <div className="h-2 w-48 bg-gradient-to-r from-indigo-400 via-purple-400 to-pink-400 rounded-full"></div>
          </div>
        </div>
      </div>
    </div>
  );
}
