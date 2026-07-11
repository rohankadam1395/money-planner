import { useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/router';
import Navbar from '@/components/Navbar';

interface Statement {
  statement_id: string;
  file_name: string;
  file_format: string;
  bank_code: string;
  transaction_count: number;
  status: string;
  uploaded_at: string;
}

interface ListResponse {
  data: Statement[];
  pagination: {
    total: number;
    limit: number;
    offset: number;
  };
}

export default function StatementsListPage() {
  const router = useRouter();
  const [statements, setStatements] = useState<Statement[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchStatements = async () => {
      try {
        const token = localStorage.getItem('authToken');
        if (!token) {
          router.push('/auth/login');
          return;
        }

        const response = await fetch(
          `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/statements`,
          {
            headers: {
              Authorization: `Bearer ${token}`,
            },
          }
        );

        if (!response.ok) {
          throw new Error('Failed to fetch statements');
        }

        const data: ListResponse = await response.json();
        setStatements(data.data || []);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Error fetching statements');
      } finally {
        setLoading(false);
      }
    };

    fetchStatements();
  }, [router]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen bg-gradient-to-br from-blue-50 to-indigo-100">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-indigo-600 mx-auto mb-4"></div>
          <p className="text-gray-600">Loading statements...</p>
        </div>
      </div>
    );
  }

  return (
    <>
      <Navbar />
      <div className="min-h-screen bg-gradient-to-br from-teal-50 via-cyan-50 to-blue-100 p-8">
      <div className="max-w-6xl mx-auto">
        {/* Header */}
        <div className="flex justify-between items-center mb-8">
          <h1 className="text-5xl font-bold bg-gradient-to-r from-teal-600 via-cyan-600 to-blue-600 bg-clip-text text-transparent">Your Statements</h1>
          <Link
            href="/statements"
            className="bg-gradient-to-r from-teal-500 to-cyan-500 text-white px-6 py-3 rounded-xl hover:shadow-lg hover:shadow-teal-400/50 transition-all transform hover:scale-105 font-semibold"
          >
            + Upload Statement
          </Link>
        </div>

        {error && (
          <div className="bg-red-100 border border-red-400 text-red-700 px-4 py-3 rounded mb-4">
            {error}
          </div>
        )}

        {statements.length === 0 ? (
          <div className="bg-white/80 backdrop-blur-xl rounded-2xl shadow-lg p-12 text-center border border-white/40">
            <p className="text-gray-500 mb-4 text-lg">No statements uploaded yet</p>
            <Link
              href="/statements"
              className="text-cyan-600 hover:text-cyan-700 font-semibold hover:underline transition-all text-lg"
            >
              Upload your first statement →
            </Link>
          </div>
        ) : (
          <div className="overflow-x-auto bg-white/80 backdrop-blur-xl rounded-2xl shadow-lg border border-white/40">
            <table className="w-full">
              <thead className="bg-gradient-to-r from-cyan-50 to-teal-50 border-b-2 border-cyan-200">
                <tr>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">File Name</th>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">Bank</th>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">Format</th>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">Transactions</th>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">Status</th>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">Uploaded</th>
                  <th className="px-6 py-4 text-left text-sm font-bold text-cyan-900">Action</th>
                </tr>
              </thead>
              <tbody>
                {statements.map((stmt) => (
                  <tr key={stmt.statement_id} className="border-b hover:bg-cyan-50/50 transition">
                    <td className="px-6 py-4 text-sm font-medium text-gray-900">{stmt.file_name}</td>
                    <td className="px-6 py-4 text-sm text-gray-700">{stmt.bank_code}</td>
                    <td className="px-6 py-4 text-sm text-gray-700">{stmt.file_format}</td>
                    <td className="px-6 py-4 text-sm font-semibold text-cyan-600">{stmt.transaction_count}</td>
                    <td className="px-6 py-4 text-sm">
                      <span
                        className={`px-3 py-1 rounded-full text-xs font-bold ${
                          stmt.status === 'SUCCESS'
                            ? 'bg-emerald-200 text-emerald-800'
                            : stmt.status === 'PENDING'
                            ? 'bg-amber-200 text-amber-800'
                            : 'bg-rose-200 text-rose-800'
                        }`}
                      >
                        {stmt.status}
                      </span>
                    </td>
                    <td className="px-6 py-4 text-sm text-gray-700">{stmt.uploaded_at}</td>
                    <td className="px-6 py-4 text-sm">
                      <Link
                        href={`/statements/preview?id=${stmt.statement_id}`}
                        className="text-cyan-600 hover:text-cyan-700 font-bold hover:underline transition-all"
                      >
                        View
                      </Link>
                    </td>
                  </tr>
                ))}
              </tbody>
            </table>
          </div>
        )}
      </div>
      </div>
    </>
  );
}
