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
    getStats: () => fetch(`${API_BASE_URL}/api/dashboard/stats`, {
      headers: getAuthHeaders(),
    }).then(res => res.json()),
    getActiveUsers: () => fetch(`${API_BASE_URL}/api/dashboard/active-users`, {
      headers: getAuthHeaders(),
    }).then(res => res.json()),
    getRecentActivity: () => fetch(`${API_BASE_URL}/api/dashboard/recent-activity`, {
      headers: getAuthHeaders(),
    }).then(res => res.json()),
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
    
    createEvent: (data: { title: string; description: string; start_date: string; end_date: string; start_time: string; end_time: string; color: string; location: string; tags: string[] }) =>
      fetch(`${API_BASE_URL}/api/events`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    updateEvent: (id: number, data: { title: string; description: string; start_date: string; end_date: string; start_time: string; end_time: string; color: string; location: string; tags: string[] }) =>
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

    // Event Guests
    getGuests: (eventId: number) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/guests`).then(res => res.json()),
    
    inviteGuest: (eventId: number, userId: string) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/guests`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify({ user_id: userId }),
      }).then(res => res.json()),
    
    removeGuest: (eventId: number, guestId: number) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/guests/${guestId}`, {
        method: 'DELETE',
        headers: { 'X-User-ID': getUserId() },
      }),
    
    updateGuestRSVP: (eventId: number, guestId: number, status: string) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/guests/${guestId}/rsvp`, {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify({ status }),
      }).then(res => res.json()),

    // Event Checklist
    getChecklist: (eventId: number) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/checklist`).then(res => res.json()),
    
    createChecklistItem: (eventId: number, text: string) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/checklist`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify({ text }),
      }).then(res => res.json()),
    
    updateChecklistItem: (eventId: number, itemId: number, data: { text?: string; is_completed?: boolean }) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/checklist/${itemId}`, {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify(data),
      }).then(res => res.json()),
    
    deleteChecklistItem: (eventId: number, itemId: number) =>
      fetch(`${API_BASE_URL}/api/events/${eventId}/checklist/${itemId}`, {
        method: 'DELETE',
        headers: { 'X-User-ID': getUserId() },
      }),
  },

  // Event Labels (global per guild)
  eventLabels: {
    getAll: () => fetch(`${API_BASE_URL}/api/event-labels`).then(res => res.json()),
    
    create: (name: string, color: string) =>
      fetch(`${API_BASE_URL}/api/event-labels`, {
        method: 'POST',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify({ name, color }),
      }).then(res => res.json()),
    
    update: (labelId: number, name: string, color: string) =>
      fetch(`${API_BASE_URL}/api/event-labels/${labelId}`, {
        method: 'PUT',
        headers: { 
          'Content-Type': 'application/json',
          'X-User-ID': getUserId(),
        },
        body: JSON.stringify({ name, color }),
      }).then(res => res.json()),
    
    delete: (labelId: number) =>
      fetch(`${API_BASE_URL}/api/event-labels/${labelId}`, {
        method: 'DELETE',
        headers: { 'X-User-ID': getUserId() },
      }),
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

  // Generic HTTP client for additional API calls
  get: async (url: string) => {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      method: 'GET',
      headers: getAuthHeaders(),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP ${response.status}`);
    }
    const text = await response.text();
    return text ? JSON.parse(text) : null;
  },

  post: async (url: string, data?: any) => {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      method: 'POST',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP ${response.status}`);
    }
    const text = await response.text();
    return text ? JSON.parse(text) : null;
  },

  put: async (url: string, data?: any) => {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      method: 'PUT',
      headers: getAuthHeaders(),
      body: JSON.stringify(data),
    });
    if (!response.ok) {
      const errorText = await response.text();
      throw new Error(errorText || `HTTP ${response.status}`);
    }
    const text = await response.text();
    return text ? JSON.parse(text) : null;
  },

  delete: async (url: string) => {
    const response = await fetch(`${API_BASE_URL}${url}`, {
      method: 'DELETE',
      headers: getAuthHeaders(),
    });
    if (!response.ok && response.status !== 204) throw response;
    if (response.status === 204) return;
    return response.json();
  },
};
