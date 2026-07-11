import Link from 'next/link';
import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';

export default function Navbar() {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [email, setEmail] = useState('');

  useEffect(() => {
    const token = localStorage.getItem('authToken');
    const userEmail = localStorage.getItem('email');
    if (token) {
      setIsAuthenticated(true);
      setEmail(userEmail || '');
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('authToken');
    localStorage.removeItem('userId');
    localStorage.removeItem('email');
    router.push('/auth/login');
  };

  return (
    <nav className="bg-gradient-to-r from-indigo-600 via-purple-600 to-pink-600 shadow-lg">
      <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
        <Link href="/" className="text-3xl font-bold text-white drop-shadow-lg hover:scale-105 transition-transform">
          💰 Money Planner
        </Link>

        {isAuthenticated ? (
          <div className="flex items-center gap-6">
            <Link href="/statements/list" className="text-white/90 hover:text-white font-medium transition-colors">
              📊 My Statements
            </Link>
            <Link href="/statements" className="text-white/90 hover:text-white font-medium transition-colors">
              📤 Upload
            </Link>
            <span className="text-sm text-white/80 px-3 py-1 bg-white/20 rounded-full">{email}</span>
            <button
              onClick={handleLogout}
              className="bg-rose-500 hover:bg-rose-600 text-white px-4 py-2 rounded-lg transition-all font-semibold transform hover:scale-105 text-sm"
            >
              Logout
            </button>
          </div>
        ) : (
          <div className="flex gap-4">
            <Link
              href="/auth/login"
              className="text-white/90 hover:text-white font-medium transition-colors"
            >
              Sign In
            </Link>
            <Link
              href="/auth/register"
              className="bg-white/20 hover:bg-white/30 text-white px-4 py-2 rounded-lg transition-all font-semibold"
            >
              Register
            </Link>
          </div>
        )}
      </div>
    </nav>
  );
}
