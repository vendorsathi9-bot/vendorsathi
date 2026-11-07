import React, { createContext, useContext, useState, useEffect } from 'react';
import { ApiClient, User } from '../api/apiClient';

interface AuthContextType {
  user: User | null;
  loading: boolean;
  login: (phone: string, pass: string) => Promise<void>;
  logout: () => Promise<void>;
  isAdmin: boolean;
}

const AuthContext = createContext<AuthContextType>({} as AuthContextType);

export const AuthProvider: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const init = async () => {
      const token = ApiClient.getToken();
      if (!token) {
        setLoading(false);
        return;
      }
      try {
        const res = await ApiClient.me();
        setUser(res.user);
      } catch (err) {
        ApiClient.setToken(null);
      } finally {
        setLoading(false);
      }
    };

    init();

    const handleUnauthorized = () => {
      setUser(null);
    };

    window.addEventListener('auth:unauthorized', handleUnauthorized);
    return () => window.removeEventListener('auth:unauthorized', handleUnauthorized);
  }, []);

  const login = async (phone: string, pass: string) => {
    const data = await ApiClient.login(phone, pass);
    setUser(data.user);
  };

  const logout = async () => {
    await ApiClient.logout();
    setUser(null);
  };

  const isAdmin = user ? user.roles?.some(r => r === 'ADMIN' || r === 'SUPER_ADMIN') : false;

  return (
    <AuthContext.Provider value={{ user, loading, login, logout, isAdmin }}>
      {children}
    </AuthContext.Provider>
  );
};

export const useAuth = () => useContext(AuthContext);
