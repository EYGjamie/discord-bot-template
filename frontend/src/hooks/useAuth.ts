import { useState, useEffect } from 'react';
import type { User } from '../types';

export const useAuth = () => {
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    // Check if user is logged in by checking localStorage token
    const token = localStorage.getItem('token');
    if (token) {
      // For now: Mock user data when token exists
      // TODO: Fetch real user data from API with token
      const mockUser = {
        id: '123456789',
        discord_id: '123456789',
        username: 'TestUser',
        email: 'test@example.com',
        avatar: 'https://cdn.discordapp.com/avatars/123456789/avatar.png',
        discriminator: '0001',
        is_admin: false,
        created_at: new Date().toISOString(),
        updated_at: new Date().toISOString(),
      };
      
      // Store discord_user_id for API calls that require authentication
      localStorage.setItem('discord_user_id', mockUser.discord_id);
      setUser(mockUser);
    }
    setLoading(false);
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
