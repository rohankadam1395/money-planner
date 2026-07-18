import React, { createContext, useContext, useEffect, useState } from 'react';
import { apiClient } from '@/services/api';

export interface AuthContextType {
  token: string | null;
  userId: string | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (token: string, userId: string) => void;
  logout: () => void;
  setToken: (token: string) => void;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

const TOKEN_STORAGE_KEY = 'authToken';
const USER_ID_STORAGE_KEY = 'userId';

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [token, setTokenState] = useState<string | null>(null);
  const [userId, setUserIdState] = useState<string | null>(null);
  const [isLoading, setIsLoading] = useState(true);

  // Initialize from localStorage on mount
  useEffect(() => {
    const storedToken = localStorage.getItem(TOKEN_STORAGE_KEY);
    const storedUserId = localStorage.getItem(USER_ID_STORAGE_KEY);

    if (storedToken && storedUserId) {
      setTokenState(storedToken);
      setUserIdState(storedUserId);
      apiClient.setTokenGetter(async () => storedToken);
    }

    setIsLoading(false);
  }, []);

  // Setup unauthorized handler
  useEffect(() => {
    apiClient.setUnauthorizedHandler(() => {
      logout();
    });
  }, []);

  const login = (newToken: string, newUserId: string) => {
    setTokenState(newToken);
    setUserIdState(newUserId);
    localStorage.setItem(TOKEN_STORAGE_KEY, newToken);
    localStorage.setItem(USER_ID_STORAGE_KEY, newUserId);
    apiClient.setTokenGetter(async () => newToken);
  };

  const logout = () => {
    setTokenState(null);
    setUserIdState(null);
    localStorage.removeItem(TOKEN_STORAGE_KEY);
    localStorage.removeItem(USER_ID_STORAGE_KEY);
    apiClient.setTokenGetter(async () => '');
  };

  const setToken = (newToken: string) => {
    if (!userId) {
      console.error('Cannot set token without userId');
      return;
    }
    login(newToken, userId);
  };

  const value: AuthContextType = {
    token,
    userId,
    isAuthenticated: !!token,
    isLoading,
    login,
    logout,
    setToken,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}

export function useAuth(): AuthContextType {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}
