import { createContext, useContext, useState, useEffect, useCallback } from 'react';
import { setAuthToken, getMe, login as apiLogin, signup as apiSignup } from './api.js';

const AuthContext = createContext(null);

export function AuthProvider({ children }) {
  const [user, setUser] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (!token) {
      setLoading(false);
      return;
    }
    setAuthToken(token);
    getMe()
      .then((u) => setUser(u))
      .catch(() => {
        localStorage.removeItem('token');
        setAuthToken(null);
      })
      .finally(() => setLoading(false));
  }, []);

  const login = useCallback(async (email, password) => {
    const data = await apiLogin(email, password);
    localStorage.setItem('token', data.token);
    setAuthToken(data.token);
    setUser(data.user);
    return data;
  }, []);

  const signup = useCallback(async (email, password, displayName) => {
    const data = await apiSignup(email, password, displayName);
    localStorage.setItem('token', data.token);
    setAuthToken(data.token);
    setUser(data.user);
    return data;
  }, []);

  const logout = useCallback(() => {
    localStorage.removeItem('token');
    setAuthToken(null);
    setUser(null);
    window.location.hash = 'login';
  }, []);

  return (
    <AuthContext.Provider value={{ user, loading, login, signup, logout }}>
      {children}
    </AuthContext.Provider>
  );
}

export function useAuth() {
  const ctx = useContext(AuthContext);
  if (!ctx) throw new Error('useAuth must be used within AuthProvider');
  return ctx;
}
