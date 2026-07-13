import { useState, useEffect, useCallback } from 'react';
import { AuthProvider, useAuth } from './AuthContext.jsx';
import HotelsPage from './HotelsPage.jsx';
import SettingsPage from './SettingsPage.jsx';
import LoginPage from './LoginPage.jsx';
import SignupPage from './SignupPage.jsx';

function getHash() {
  const hash = window.location.hash.slice(1);
  const idx = hash.indexOf('?');
  return idx >= 0 ? hash.slice(0, idx) : hash || 'login';
}

function getTokenFromHash() {
  const hash = window.location.hash.slice(1);
  const params = new URLSearchParams(hash.includes('?') ? hash.slice(hash.indexOf('?')) : '');
  return params.get('token');
}

function AppInner() {
  const { user, loading, login } = useAuth();
  const [page, setPage] = useState(getHash);

  useEffect(() => {
    const token = getTokenFromHash();
    if (token) {
      window.location.hash = 'login';
      localStorage.setItem('token', token);
      window.location.reload();
      return;
    }
  }, []);

  const onHashChange = useCallback(() => setPage(getHash()), []);

  useEffect(() => {
    window.addEventListener('hashchange', onHashChange);
    return () => window.removeEventListener('hashchange', onHashChange);
  }, [onHashChange]);

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-dvh text-slate-400 text-sm">
        Loading…
      </div>
    );
  }

  if (!user && page !== 'signup') {
    return <LoginPage />;
  }

  if (!user && page === 'signup') {
    return <SignupPage />;
  }

  return (
    <div className="min-h-dvh grid grid-rows-[auto_1fr]">
      <header className="sticky top-0 z-10 flex items-center justify-between px-6 py-3 border-b border-slate-200 bg-white/90 backdrop-blur-md">
        <a href="#hotels" className="text-lg font-bold tracking-tight text-slate-900 no-underline hover:no-underline">
          Prices Recommender
        </a>
        <div className="flex items-center gap-2">
          <nav className="flex gap-1">
            <a
              href="#hotels"
              className={`px-3.5 py-1.5 rounded-md text-sm font-medium transition no-underline hover:no-underline ${
                page === 'hotels'
                  ? 'bg-blue-600 text-white'
                  : 'text-slate-500 hover:text-blue-600 hover:bg-blue-50'
              }`}
            >
              Hotels
            </a>
            <a
              href="#settings"
              className={`px-3.5 py-1.5 rounded-md text-sm font-medium transition no-underline hover:no-underline ${
                page === 'settings'
                  ? 'bg-blue-600 text-white'
                  : 'text-slate-500 hover:text-blue-600 hover:bg-blue-50'
              }`}
            >
              Settings
            </a>
          </nav>
          <div className="ml-3 pl-3 border-l border-slate-200 flex items-center gap-2">
            <span className="text-sm text-slate-500 hidden sm:inline">{user?.display_name || user?.email}</span>
            <button
              onClick={() => {
                localStorage.removeItem('token');
                window.location.reload();
              }}
              className="text-xs text-slate-400 hover:text-red-500 transition"
            >
              Logout
            </button>
          </div>
        </div>
      </header>
      <main className="p-6">
        {page === 'hotels' && <HotelsPage />}
        {page === 'settings' && <SettingsPage />}
      </main>
    </div>
  );
}

export default function App() {
  return (
    <AuthProvider>
      <AppInner />
    </AuthProvider>
  );
}
