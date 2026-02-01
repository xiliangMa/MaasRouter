'use client';

import { createContext, useContext, useState, useEffect, ReactNode } from 'react';
import { useRouter } from 'next/navigation';
import api from '@/lib/api';

interface User {
  id: string;
  username: string;
  email: string;
  role: string;
  status: string;
  email_verified: boolean;
  created_at: string;
  updated_at: string;
  deleted_at: string | null;
  balance?: number;
}

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (email: string, password: string) => Promise<void>;
  register: (username: string, email: string, password: string) => Promise<void>;
  loginWithOAuth: (provider: 'github' | 'google') => Promise<void>;
  logout: () => void;
  refreshToken: () => Promise<boolean>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export function useAuth() {
  const context = useContext(AuthContext);
  if (context === undefined) {
    throw new Error('useAuth must be used within an AuthProvider');
  }
  return context;
}

interface AuthProviderProps {
  children: ReactNode;
}

export function AuthProvider({ children }: AuthProviderProps) {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);
  const router = useRouter();

  useEffect(() => {
    const initializeAuth = async () => {
      const token = localStorage.getItem('access_token');
      const storedUser = localStorage.getItem('user');

      if (token && storedUser) {
        try {
          setUser(JSON.parse(storedUser));
        } catch (error) {
          localStorage.removeItem('access_token');
          localStorage.removeItem('refresh_token');
          localStorage.removeItem('user');
        }
      }
      setLoading(false);
    };

    initializeAuth();
  }, []);

  const login = async (email: string, password: string) => {
    const response = await api.post('/auth/login', {
      email,
      password,
    });

    const { access_token, refresh_token, user } = response.data.data;
    
    localStorage.setItem('access_token', access_token);
    localStorage.setItem('refresh_token', refresh_token);
    localStorage.setItem('user', JSON.stringify(user));
    
    setUser(user);
    router.push('/dashboard');
  };

  const register = async (username: string, email: string, password: string) => {
    await api.post('/auth/register', {
      username,
      email,
      password,
    });
  };

  const loginWithOAuth = async (provider: 'github' | 'google') => {
    // TODO: Implement OAuth flow once backend endpoints are ready
    // For now, show alert and redirect to login page
    alert(`${provider} OAuth is not yet implemented. Please use email/password for now.`);
    
    // Example of what the OAuth flow would look like:
    // const response = await api.get(`/auth/oauth/${provider}/redirect`);
    // window.location.href = response.data.redirect_url;
  };

  const logout = () => {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
    localStorage.removeItem('user');
    setUser(null);
    router.push('/login');
  };

  const refreshToken = async (): Promise<boolean> => {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) {
      return false;
    }

    try {
      const response = await api.post('/auth/refresh', {
        refresh_token: refreshToken,
      });

      const { access_token, refresh_token: newRefreshToken } = response.data.data;
      localStorage.setItem('access_token', access_token);
      localStorage.setItem('refresh_token', newRefreshToken);
      return true;
    } catch (error) {
      logout();
      return false;
    }
  };

  return (
    <AuthContext.Provider value={{ user, loading, login, register, loginWithOAuth, logout, refreshToken }}>
      {children}
    </AuthContext.Provider>
  );
}