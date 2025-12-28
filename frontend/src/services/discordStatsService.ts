import axios from 'axios';
import type { DiscordStatistic, DiscordStatsResponse } from '../types';

const API_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
const BOT_API_URL = import.meta.env.VITE_BOT_API_URL || 'http://localhost:8090';

// Helper to get Authorization headers
const getAuthHeaders = () => {
  const token = localStorage.getItem('token');
  return {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json',
  };
};

export const discordStatsService = {
  // Holt aktuelle Statistiken und speichert sie (triggert Bot API Calls)
  async getCurrentStats(): Promise<DiscordStatsResponse> {
    const response = await axios.get<DiscordStatsResponse>(
      `${API_URL}/api/discord/stats/current`,
      { headers: getAuthHeaders() }
    );
    return response.data;
  },
  
  // Direkter Zugriff auf Bot API (ohne Credentials)
  async getBotStats(): Promise<any> {
    const guildId = import.meta.env.VITE_DISCORD_GUILD_ID;
    const response = await axios.get(
      `${BOT_API_URL}/api/guild/member-count?guild_id=${guildId}`
    );
    return response.data;
  },

  // Holt historische Statistiken
  async getHistoricalStats(limit?: number, since?: number): Promise<DiscordStatistic[]> {
    const params = new URLSearchParams();
    if (limit) params.append('limit', limit.toString());
    if (since) params.append('since', since.toString());

    const response = await axios.get<DiscordStatistic[]>(
      `${API_URL}/api/discord/stats/historical?${params.toString()}`,
      { headers: getAuthHeaders() }
    );
    return response.data;
  },

  // Holt Statistiken in einem Zeitbereich
  async getStatsInRange(startTime: number, endTime: number): Promise<DiscordStatistic[]> {
    const params = new URLSearchParams({
      start: startTime.toString(),
      end: endTime.toString(),
    });

    const response = await axios.get<DiscordStatistic[]>(
      `${API_URL}/api/discord/stats/range?${params.toString()}`,
      { headers: getAuthHeaders() }
    );
    return response.data;
  },

  // Berechnet Änderungen zwischen zwei Statistiken
  calculateChange(current: number, previous: number): { absolute: number; relative: number; direction: 'up' | 'down' | 'stable' } {
    const absolute = current - previous;
    const relative = previous > 0 ? (absolute / previous) * 100 : 0;
    
    let direction: 'up' | 'down' | 'stable' = 'stable';
    if (absolute > 0) direction = 'up';
    else if (absolute < 0) direction = 'down';

    return { absolute, relative, direction };
  },

  // Holt Statistiken für verschiedene Zeiträume
  async getStatsForPeriod(period: 'hour' | 'day' | 'week' | 'month'): Promise<DiscordStatistic[]> {
    const now = Date.now();
    let startTime: number;

    switch (period) {
      case 'hour':
        startTime = now - 60 * 60 * 1000;
        break;
      case 'day':
        startTime = now - 24 * 60 * 60 * 1000;
        break;
      case 'week':
        startTime = now - 7 * 24 * 60 * 60 * 1000;
        break;
      case 'month':
        startTime = now - 30 * 24 * 60 * 60 * 1000;
        break;
    }

    return this.getStatsInRange(Math.floor(startTime / 1000), Math.floor(now / 1000));
  },
};
