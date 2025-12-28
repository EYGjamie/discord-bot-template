export interface User {
  id: string;
  discord_id: string;
  username: string;
  email: string;
  avatar: string;
  avatar_url?: string;
  discriminator: string;
  is_admin: boolean;
  is_moderator: boolean;
  created_at: string;
  updated_at: string;
  global_name?: string;
  display_name?: string;
  nick?: string;
  joined_at?: string;
  roles?: Role[];
}

export interface Role {
  id: string;
  name: string;
  color: number;
  position: number;
}

export interface Permission {
  level: 'public' | 'member' | 'moderator' | 'admin';
}

export interface PermissionCheck {
  canViewMembers: boolean;
  canModerate: boolean;
  canViewAuditLogs: boolean;
  canManageSettings: boolean;
  isAdmin: boolean;
  isModerator: boolean;
  isMember: boolean;
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

export interface CalendarEvent {
  id: number;
  guild_id: string;
  title: string;
  description: string;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  start_time: string; // HH:MM
  end_time: string;   // HH:MM
  is_all_day: boolean; // Ganztägiges Event
  color: string;      // Hex color
  location: string;
  guests: string;     // Comma-separated list of guest names
  created_by: string;
  creator_name: string;
  creator_avatar: string;
  created_at: string;
  updated_at: string;
}

export interface CreateEventRequest {
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  start_time: string;
  end_time: string;
  is_all_day: boolean;
  color: string;
  location: string;
  guests: string;
}

export interface CalendarMatch {
  id: number;
  guild_id: string;
  title: string;
  description: string;
  start_date: string; // YYYY-MM-DD
  end_date: string;   // YYYY-MM-DD
  start_time: string; // HH:MM
  end_time: string;   // HH:MM
  is_all_day: boolean; // Ganztägiges Match
  color: string;      // Hex color
  location: string;
  guests: string;     // Comma-separated list of guest names
  created_by: string;
  creator_name: string;
  creator_avatar: string;
  created_at: string;
  updated_at: string;
}

export interface CreateMatchRequest {
  title: string;
  description: string;
  start_date: string;
  end_date: string;
  start_time: string;
  end_time: string;
  is_all_day: boolean;
  color: string;
  location: string;
  guests: string;
}

export interface DiscordStatistic {
  id: number;
  guild_id: string;
  member_count: number;
  role_member_count: number;
  role_id: string;
  total_channels: number;
  text_channels: number;
  voice_channels: number;
  category_channels: number;
  voice_user_count: number;
  active_voice_channels: number;
  total_voice_time: number;
  timestamp: string;
  source: string;
  created_at: string;
}

export interface AdditionalStats {
  user_max: number;
  total_messages: number;
  total_voice_time: number;
  avg_voice_time_day: number;
}

export interface DiscordStatsResponse {
  current_stats: DiscordStatistic;
  user_max: number;
  total_messages: number;
  total_voice_time: number;
  avg_voice_time_day: number;
}

export interface StatisticChange {
  absolute: number;
  relative: number;
  direction: 'up' | 'down' | 'stable';
}

export interface StatisticComparison {
  current: DiscordStatistic;
  previous?: DiscordStatistic;
  memberChange?: StatisticChange;
  roleMemberChange?: StatisticChange;
  channelChange?: StatisticChange;
  voiceUserChange?: StatisticChange;
}
