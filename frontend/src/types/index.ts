export interface User {
  id: string;
  discord_id: string;
  username: string;
  email: string;
  avatar: string;
  discriminator: string;
  is_admin: boolean;
  created_at: string;
  updated_at: string;
}

export interface DashboardStats {
  total_users: number;
  total_teams: number;
  open_tickets: number;
  active_members: number;
}

export interface Activity {
  id: string;
  type: string;
  user_id: string;
  message: string;
  timestamp: string;
}

export interface Event {
  id: string;
  title: string;
  description: string;
  date: string;
  time: string;
  platform: string;
  participants: number;
  status: 'upcoming' | 'confirmed' | 'completed';
}

export interface Member {
  id: string;
  name: string;
  role: string;
  status: 'online' | 'offline';
}
