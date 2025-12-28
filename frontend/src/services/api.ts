const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

// Helper to get Authorization header with JWT token
const getAuthHeaders = (): HeadersInit => {
  const token = localStorage.getItem('token');
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };
  
  if (token) {
    headers['Authorization'] = `Bearer ${token}`;
  }
  
  return headers;
};

// Helper to get user ID from localStorage
const getUserId = (): string => {
  return localStorage.getItem('discord_user_id') || '';
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
      
      return fetch(`${API_BASE_URL}/api/members?${query.toString()}`, {
        headers: getAuthHeaders(),
      }).then(res => res.json());
    },
    getMemberById: (id: string) => fetch(`${API_BASE_URL}/api/members/${id}`, {
      headers: getAuthHeaders(),
    }).then(res => res.json()),
    getMemberStats: (id: string) => fetch(`${API_BASE_URL}/api/members/${id}/stats`, {
      headers: getAuthHeaders(),
    }).then(res => res.json()),
  },

  moderation: {
    createWarn: (data: { user_id: string; reason: string }) => 
      fetch(`${API_BASE_URL}/api/moderation/warns`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    createNote: (data: { user_id: string; reason: string }) => 
      fetch(`${API_BASE_URL}/api/moderation/notes`, {
        method: 'POST',
        headers: getAuthHeaders(),
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    deleteWarn: (id: number) => 
      fetch(`${API_BASE_URL}/api/moderation/warns/${id}`, {
        method: 'DELETE',
        headers: getAuthHeaders(),
      }).then(res => res.json()),
    
    deleteNote: (id: number) => 
      fetch(`${API_BASE_URL}/api/moderation/notes/${id}`, {
        method: 'DELETE',
        headers: getAuthHeaders(),
      }).then(res => res.json()),
  },

  events: {
    getEvents: (params?: { month?: number; year?: number }) => {
      const query = new URLSearchParams();
      if (params?.month) query.set('month', params.month.toString());
      if (params?.year) query.set('year', params.year.toString());
      
      return fetch(`${API_BASE_URL}/api/events?${query.toString()}`).then(res => res.json());
    },
    
    getEventById: (id: number) => 
      fetch(`${API_BASE_URL}/api/events/${id}`).then(res => res.json()),
    
    createEvent: (data: { title: string; description: string; start_date: string; end_date: string; start_time: string; end_time: string; color: string; location: string; guests: string }) =>
      fetch(`${API_BASE_URL}/api/events`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    updateEvent: (id: number, data: { title: string; description: string; start_date: string; end_date: string; start_time: string; end_time: string; color: string; location: string; guests: string }) =>
      fetch(`${API_BASE_URL}/api/events/${id}`, {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    deleteEvent: (id: number) =>
      fetch(`${API_BASE_URL}/api/events/${id}`, {
        method: 'DELETE',
        headers: { 
          'X-User-ID': getUserId(),
        },
      }).then(res => res.json()),
  },

  matches: {
    getMatches: (params?: { month?: number; year?: number }) => {
      const query = new URLSearchParams();
      if (params?.month) query.set('month', params.month.toString());
      if (params?.year) query.set('year', params.year.toString());
      
      return fetch(`${API_BASE_URL}/api/matches?${query.toString()}`).then(res => res.json());
    },
    
    getMatchById: (id: number) => 
      fetch(`${API_BASE_URL}/api/matches/${id}`).then(res => res.json()),
    
    createMatch: (data: { title: string; description: string; start_date: string; end_date: string; start_time: string; end_time: string; color: string; location: string; guests: string }) =>
      fetch(`${API_BASE_URL}/api/matches`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    updateMatch: (id: number, data: { title: string; description: string; start_date: string; end_date: string; start_time: string; end_time: string; color: string; location: string; guests: string }) =>
      fetch(`${API_BASE_URL}/api/matches/${id}`, {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    deleteMatch: (id: number) =>
      fetch(`${API_BASE_URL}/api/matches/${id}`, {
        method: 'DELETE',
        headers: { 
          'X-User-ID': getUserId(),
        },
      }).then(res => res.json()),
  },
};
