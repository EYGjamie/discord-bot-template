const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';

export const api = {
  auth: {
    loginUrl: () => `${API_BASE_URL}/api/auth/discord/login`,
    logout: () => fetch(`${API_BASE_URL}/api/auth/logout`, { method: 'POST' }),
    getCurrentUser: () => fetch(`${API_BASE_URL}/api/me`).then(res => res.json()),
  },
  
  dashboard: {
    getStats: () => fetch(`${API_BASE_URL}/api/dashboard/stats`).then(res => res.json()),
    getActivity: () => fetch(`${API_BASE_URL}/api/dashboard/activity`).then(res => res.json()),
  },
};
