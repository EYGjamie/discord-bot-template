import { useState, useEffect } from 'react';
import type { User } from '../types';
import { api } from '../services/api';

export const useAuth = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in by checking localStorage token
    const token = localStorage.getItem('token');
    if (token) {
      // Fetch real user data from API
      api.auth.getCurrentUser()
        .then((userData) => {
          setUser(userData);
          // Store discord_user_id for API calls that require authentication
          if (userData.discord_id) {
            localStorage.setItem('discord_user_id', userData.discord_id);
          }
        })
        .catch((error) => {
          console.error('Failed to fetch user data:', error);
          // Token might be invalid or expired
          localStorage.removeItem('token');
          localStorage.removeItem('discord_user_id');
          setUser(null);
        })
        .finally(() => {
          setLoading(false);
        });
    } else {
      setLoading(false);
    }
  }, []);

  const login = () => {
    window.location.href = 'http://localhost:8080/api/auth/discord/login';
  };

  const logout = () => {
    localStorage.removeItem('token');
    localStorage.removeItem('discord_user_id');
    setUser(null);
  };

  return { user, loading, login, logout, isAuthenticated: !!user };
};
