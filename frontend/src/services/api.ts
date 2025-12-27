const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Helper to get current user's Discord ID from localStorage
// The user's discord_id should be stored during login
const getUserId = (): string => {
  const userId = localStorage.getItem('discord_user_id');
  if (!userId) {
    console.warn('No discord_user_id found in localStorage. User might not be authenticated.');
  }
  return userId || '';
};

export const api = {
  auth: {
    loginUrl: () => `${API_BASE_URL}/api/auth/discord/login`,
    logout: () => fetch(`${API_BASE_URL}/api/auth/logout`, { method: 'POST' }),
    getCurrentUser: () => {
      const token = localStorage.getItem('token');
      return fetch(`${API_BASE_URL}/api/me`, {
        headers: {
          'Authorization': `Bearer ${token}`,
        },
      }).then(res => {
        if (!res.ok) {
          throw new Error('Failed to fetch user data');
        }
        return res.json();
      });
    },
  },
  
  dashboard: {
    getStats: () => fetch(`${API_BASE_URL}/api/dashboard/stats`).then(res => res.json()),
    getActivity: () => fetch(`${API_BASE_URL}/api/dashboard/activity`).then(res => res.json()),
  },
  
  members: {
    getMembers: (params?: { page?: number; per_page?: number; search?: string; role?: string }) => {
      const query = new URLSearchParams();
      if (params?.page) query.set('page', params.page.toString());
      if (params?.per_page) query.set('per_page', params.per_page.toString());
      if (params?.search) query.set('search', params.search);
      if (params?.role) query.set('role', params.role);
      
      return fetch(`${API_BASE_URL}/api/members?${query.toString()}`).then(res => res.json());
    },
    getMemberById: (id: string) => fetch(`${API_BASE_URL}/api/members/${id}`).then(res => res.json()),
    getMemberStats: (id: string) => fetch(`${API_BASE_URL}/api/members/${id}/stats`).then(res => res.json()),
  },

  moderation: {
    createWarn: (data: { user_id: string; reason: string }) => 
      fetch(`${API_BASE_URL}/api/moderation/warns`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    createNote: (data: { user_id: string; reason: string }) => 
      fetch(`${API_BASE_URL}/api/moderation/notes`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    deleteWarn: (id: number) => 
      fetch(`${API_BASE_URL}/api/moderation/warns/${id}`, {
        method: 'DELETE',
      }).then(res => res.json()),
    
    deleteNote: (id: number) => 
      fetch(`${API_BASE_URL}/api/moderation/notes/${id}`, {
        method: 'DELETE',
      }).then(res => res.json()),
  },
};
