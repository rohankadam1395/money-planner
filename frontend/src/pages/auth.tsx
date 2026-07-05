'use client';

import { useState } from 'react';
import { useRouter } from 'next/router';
import { useAuth } from '@/contexts/AuthContext';

export default function AuthPage() {
  const router = useRouter();
  const { login } = useAuth();
  const [username, setUsername] = useState('');
  const [error, setError] = useState('');
  const [isLoading, setIsLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setIsLoading(true);

    try {
      if (!username.trim()) {
        setError('Please enter a username');
        setIsLoading(false);
        return;
      }

      // Call backend to get a proper JWT token
      const response = await fetch('http://localhost:8080/api/auth/login', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          username: username.trim(),
        }),
      });

      if (!response.ok) {
        const errorText = await response.text();
        throw new Error(errorText || 'Login failed');
      }

      const data = await response.json();

      console.log('Auth response:', data);
      console.log('Logging in with userID:', data.user_id);

      // Login with token from backend
      login(data.token, data.user_id);

      // Redirect to statements page
      setTimeout(() => {
        router.push('/statements');
      }, 500);
    } catch (err: any) {
      setError(err.message || 'Login failed. Make sure the backend is running on http://localhost:8080');
      setIsLoading(false);
    }
  };

  return (
    <div className="min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100 flex items-center justify-center py-12 px-4">
      <div className="w-full max-w-md">
        <div className="bg-white rounded-lg shadow-xl p-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">
            Money Planner
          </h1>
          <p className="text-gray-600 mb-8">Test Authentication</p>

          <form onSubmit={handleLogin} className="space-y-6">
            <div>
              <label htmlFor="username" className="block text-sm font-medium text-gray-700 mb-2">
                Username
              </label>
              <input
                id="username"
                type="text"
                value={username}
                onChange={(e) => setUsername(e.target.value)}
                placeholder="Enter any username"
                disabled={isLoading}
                className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:ring-2 focus:ring-blue-500 focus:border-transparent outline-none disabled:bg-gray-100"
              />
              <p className="text-xs text-gray-500 mt-1">
                For testing: Enter any username (e.g., testuser@example.com)
              </p>
            </div>

            {error && (
              <div className="p-4 bg-red-50 border border-red-200 rounded-lg">
                <p className="text-sm text-red-800">{error}</p>
              </div>
            )}

            <button
              type="submit"
              disabled={isLoading}
              className="w-full px-4 py-2 bg-blue-600 text-white font-medium rounded-lg hover:bg-blue-700 disabled:bg-gray-400 disabled:cursor-not-allowed transition-colors"
            >
              {isLoading ? 'Logging in...' : 'Login'}
            </button>
          </form>

          <div className="mt-8 p-4 bg-blue-50 border border-blue-200 rounded-lg">
            <h3 className="font-semibold text-blue-900 mb-2">📝 Test Credentials</h3>
            <p className="text-sm text-blue-800 mb-3">
              This is a test authentication page. Enter any username to generate a test token.
            </p>
            <ul className="text-sm text-blue-800 space-y-1">
              <li>✓ Token valid for 24 hours</li>
              <li>✓ Stored in localStorage</li>
              <li>✓ Can access upload functionality</li>
            </ul>
          </div>

          <div className="mt-6 p-4 bg-yellow-50 border border-yellow-200 rounded-lg">
            <p className="text-xs text-yellow-800">
              <strong>Note:</strong> This auth page is for local testing only. In production, implement proper OAuth/JWT authentication.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
}
