import React, { createContext, useContext, useEffect, useState, ReactNode } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import apiClient, { tokenManager } from '../lib/apiClient';

interface User {
  id: string;
  username: string;
  email: string;
  firstName: string;
  lastName: string;
}

interface AuthContextType {
  user: User | null;
  isAuthenticated: boolean;
  isLoading: boolean;
  login: (username: string, password: string) => Promise<void>;
  logout: () => Promise<void>;
  refreshAuth: () => Promise<void>;
}

const AuthContext = createContext<AuthContextType | undefined>(undefined);

export const useAuth = () => {
  const context = useContext(AuthContext);
  if (!context) {
    throw new Error('useAuth must be used within AuthProvider');
  }
  return context;
};

interface AuthProviderProps {
  children: ReactNode;
}

export const AuthProvider: React.FC<AuthProviderProps> = ({ children }) => {
  const [user, setUser] = useState<User | null>(null);
  const queryClient = useQueryClient();

  // Query to check current auth status and get user info
  const { data: currentUser, isLoading: isLoadingUser } = useQuery({
    queryKey: ['auth', 'me'],
    queryFn: async () => {
      const response = await apiClient.get('/users/me');
      return response.data;
    },
    enabled: false, // Don't auto-fetch - we'll trigger manually after refresh
    retry: false,
  });

  // Mutation for login
  const loginMutation = useMutation({
    mutationFn: async ({ username, password }: { username: string; password: string }) => {
      const response = await apiClient.post('/auth/login', { username, password });
      return response.data;
    },
    onSuccess: async (data) => {
      // Store access token in memory
      tokenManager.setAccessToken(data.token);
      
      // Store auth state indicator in localStorage (not the token itself)
      localStorage.setItem('auth_state', 'authenticated');
      
      // Fetch user info
      await queryClient.invalidateQueries({ queryKey: ['auth', 'me'] });
    },
  });

  // Mutation for logout
  const logoutMutation = useMutation({
    mutationFn: async () => {
      await apiClient.post('/auth/logout');
    },
    onSuccess: () => {
      // Clear auth state
      tokenManager.clearAuth();
      setUser(null);
      queryClient.clear(); // Clear all cached queries
    },
  });

  // Function to refresh auth on app load
  const refreshAuth = async () => {
    try {
      // Attempt to refresh token (cookie is sent automatically)
      const response = await apiClient.post('/auth/refresh');
      const { token } = response.data;
      
      // Store new access token
      tokenManager.setAccessToken(token);
      
      // Fetch user info
      const userResponse = await apiClient.get('/users/me');
      setUser(userResponse.data);
      
      localStorage.setItem('auth_state', 'authenticated');
    } catch (error) {
      // Refresh failed - user is not authenticated
      tokenManager.clearAuth();
      setUser(null);
      localStorage.removeItem('auth_state');
    }
  };

  // On app load, check if user should be authenticated
  useEffect(() => {
    const authState = localStorage.getItem('auth_state');
    if (authState === 'authenticated') {
      // Silently attempt to refresh and restore session
      refreshAuth();
    }
  }, []);

  // Update user state when currentUser query updates
  useEffect(() => {
    if (currentUser) {
      setUser(currentUser);
    }
  }, [currentUser]);

  const login = async (username: string, password: string) => {
    await loginMutation.mutateAsync({ username, password });
    // After successful login, fetch user
    await queryClient.invalidateQueries({ queryKey: ['auth', 'me'] });
  };

  const logout = async () => {
    await logoutMutation.mutateAsync();
  };

  const value: AuthContextType = {
    user,
    isAuthenticated: !!user,
    isLoading: isLoadingUser || loginMutation.isPending || logoutMutation.isPending,
    login,
    logout,
    refreshAuth,
  };

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
};

