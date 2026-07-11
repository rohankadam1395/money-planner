import Link from 'next/link';
import { useRouter } from 'next/router';
import { useEffect, useState } from 'react';

export default function Navbar() {
  const router = useRouter();
  const [isAuthenticated, setIsAuthenticated] = useState(false);
  const [email, setEmail] = useState('');

  useEffect(() => {
    const token = localStorage.getItem('token');
    const userEmail = localStorage.getItem('email');
    if (token) {
      setIsAuthenticated(true);
      setEmail(userEmail || '');
    }
  }, []);

  const handleLogout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('user_id');
    localStorage.removeItem('email');
    router.push('/auth/login');
  };

  return (
    <nav className="bg-white shadow-md">
      <div className="max-w-7xl mx-auto px-4 py-4 flex justify-between items-center">
        <Link href="/" className="text-2xl font-bold text-indigo-600">
          Money Planner
        </Link>

        {isAuthenticated ? (
          <div className="flex items-center gap-6">
            <Link href="/statements/list" className="text-gray-600 hover:text-gray-900">
              My Statements
            </Link>
            <Link href="/statements" className="text-gray-600 hover:text-gray-900">
              Upload
            </Link>
            <span className="text-sm text-gray-500">{email}</span>
            <button
              onClick={handleLogout}
              className="bg-red-600 text-white px-4 py-2 rounded-lg hover:bg-red-700 transition text-sm"
            >
              Logout
            </button>
          </div>
        ) : (
          <div className="flex gap-4">
            <Link
              href="/auth/login"
              className="text-gray-600 hover:text-gray-900"
            >
              Sign In
            </Link>
            <Link
              href="/auth/register"
              className="bg-indigo-600 text-white px-4 py-2 rounded-lg hover:bg-indigo-700"
            >
              Register
            </Link>
          </div>
        )}
      </div>
    </nav>
  );
}
