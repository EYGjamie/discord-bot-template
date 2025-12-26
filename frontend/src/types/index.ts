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

export interface Member {
  id: string;
  name: string;
  global_name: string;
  display_name: string;
  bot: boolean;
  avatar: string;
  avatar_url: string;
  mention: string;
  created_at: string;
  nick: string;
  joined_at: string;
  top_role: string;
  top_role_name?: string;
  top_role_color?: string;
  timed_out_until?: string;
  premium_since?: string;
  updated_at: string;
}

export interface MemberStats {
  total_messages: number;
  total_voice_time: number; // in seconds
  top_text_channel?: {
    id: string;
    name: string;
    message_count: number;
  };
  top_voice_channel?: {
    id: string;
    name: string;
    duration: number; // in seconds
  };
  muted_duration: number;
  deafen_duration: number;
  stream_duration: number;
  join_count: number;
  total_joins: number;
  total_leaves: number;
  total_warns: number;
  roles: Array<{
    id: string;
    name: string;
    color: string;
  }>;
  warns: Array<{
    id: number;
    moderator_id: string;
    moderator_name: string;
    reason: string;
    created_at: string;
  }>;
  notes: Array<{
    id: number;
    moderator_id: string;
    moderator_name: string;
    reason: string;
    created_at: string;
  }>;
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
