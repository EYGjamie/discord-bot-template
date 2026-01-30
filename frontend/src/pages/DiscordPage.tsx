import React, { useEffect, useState } from 'react';
import { 
  Users, 
  Hash, 
  Volume2, 
  TrendingUp, 
  TrendingDown, 
  Minus,
  RefreshCw,
  Calendar,
  MessageSquare,
  Clock
} from 'lucide-react';
import { discordStatsService } from '../services/discordStatsService';
import type { DiscordStatistic, AdditionalStats } from '../types';

type TimeRange = 'day' | 'week' | 'month' | 'total';

const DiscordPage: React.FC = () => {
  const [currentStats, setCurrentStats] = useState<DiscordStatistic | null>(null);
  const [additionalStats, setAdditionalStats] = useState<AdditionalStats | null>(null);
  const [historicalStats, setHistoricalStats] = useState<DiscordStatistic[]>([]);
  const [timeRange, setTimeRange] = useState<TimeRange>('day');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadData();
  }, [timeRange]);

  const loadData = async () => {
    setLoading(true);
    setError(null);
    
    try {
      // Lade aktuelle Statistiken (und speichere sie)
      const response = await discordStatsService.getCurrentStats();
      setCurrentStats(response.current_stats);
      setAdditionalStats({
        user_max: response.user_max,
        total_messages: response.total_messages,
        total_voice_time: response.total_voice_time,
        avg_voice_time_day: response.avg_voice_time_day,
      });

      // Lade historische Daten für den gewählten Zeitraum
      if (timeRange !== 'total') {
        const historical = await discordStatsService.getStatsForPeriod(timeRange);
        setHistoricalStats(historical);
      } else {
        // Lade alle Daten für "Total"
        const historical = await discordStatsService.getHistoricalStats(1000);
        setHistoricalStats(historical);
      }
    } catch (err: any) {
      // Bessere Fehlermeldungen
      if (err.response?.status === 401) {
        setError('Nicht authentifiziert. Bitte melde dich an.');
      } else if (err.response?.status === 403) {
        setError('Keine Berechtigung. Diese Seite erfordert Moderator-Rechte.');
      } else if (err.code === 'ERR_NETWORK') {
        setError('Netzwerkfehler. Ist das Backend erreichbar?');
      } else {
        setError('Fehler beim Laden der Statistiken: ' + (err.message || 'Unbekannter Fehler'));
      }
      console.error('Error loading stats:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatSeconds = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const calculateChange = (current: number, previous: number) => {
    return discordStatsService.calculateChange(current, previous);
  };

  const getPreviousStats = (): DiscordStatistic | null => {
    if (historicalStats.length < 2) return null;
    // Zweite Statistik ist die vorherige (da sortiert nach timestamp DESC)
    return historicalStats[1];
  };

  const renderChangeIndicator = (change: { absolute: number; relative: number; direction: 'up' | 'down' | 'stable' }) => {
    const Icon = change.direction === 'up' ? TrendingUp : change.direction === 'down' ? TrendingDown : Minus;
    const colorClass = change.direction === 'up' ? 'text-green-500' : change.direction === 'down' ? 'text-red-500' : 'text-gray-500';
    
    return (
      <div className={`flex items-center gap-1 text-sm ${colorClass}`}>
        <Icon size={16} />
        <span>{change.absolute > 0 ? '+' : ''}{change.absolute}</span>
        <span>({change.relative > 0 ? '+' : ''}{change.relative.toFixed(1)}%)</span>
      </div>
    );
  };

  const renderStatCard = (
    title: string,
    value: number,
    icon: React.ReactNode,
    previousValue?: number
  ) => {
    const change = previousValue !== undefined ? calculateChange(value, previousValue) : null;

    return (
      <div className="bg-[#1a1f2e] rounded-lg shadow-lg p-6 border border-gray-800">
        <div className="flex items-center justify-between mb-4">
          <h3 className="text-gray-400 text-sm font-medium">{title}</h3>
          <div className="text-blue-400">{icon}</div>
        </div>
        <div className="flex items-end justify-between">
          <div className="text-3xl font-bold text-white">{value.toLocaleString()}</div>
          {change && renderChangeIndicator(change)}
        </div>
      </div>
    );
  };

  const renderLineChart = (
    data: DiscordStatistic[], 
    valueKey: keyof DiscordStatistic, 
    title: string,
    convertToHours: boolean = false
  ) => {
    if (data.length === 0) return null;

    // Sortiere nach Timestamp aufsteigend (älteste zuerst = links, neueste zuletzt = rechts)
    const sortedData = [...data].sort((a, b) => 
      new Date(a.timestamp).getTime() - new Date(b.timestamp).getTime()
    );

    const values = sortedData.map(stat => {
      let value = typeof stat[valueKey] === 'number' ? stat[valueKey] as number : 0;
      // Konvertiere Sekunden in Stunden für Voice Time
      if (convertToHours) {
        value = value / 3600;
      }
      return value;
    });
    
    const max = Math.max(...values);
    const min = Math.min(...values);
    const range = max - min || 1;

    const points = values.map((value, index) => {
      const x = (index / (values.length - 1 || 1)) * 100;
      const y = 100 - ((value - min) / range) * 100;
      return `${x},${y}`;
    }).join(' ');

    return (
      <div className="bg-[#1a1f2e] rounded-lg shadow-lg p-6 border border-gray-800">
        <h3 className="text-lg font-semibold mb-4 text-white">{title}</h3>
        <div className="relative h-64">
          <svg className="w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
            <polyline
              fill="none"
              stroke="rgb(96, 165, 250)"
              strokeWidth="0.5"
              points={points}
            />
            <polyline
              fill="rgba(96, 165, 250, 0.1)"
              stroke="none"
              points={`0,100 ${points} 100,100`}
            />
          </svg>
          <div className="absolute top-0 right-0 text-xs text-gray-500">
            {convertToHours ? `${max.toFixed(0)}h` : max.toLocaleString()}
          </div>
          <div className="absolute bottom-0 right-0 text-xs text-gray-500">
            {convertToHours ? `${min.toFixed(0)}h` : min.toLocaleString()}
          </div>
        </div>
        <div className="flex justify-between mt-2 text-xs text-gray-500">
          <span>{new Date(sortedData[0]?.timestamp || '').toLocaleDateString('de-DE', { day: '2-digit', month: 'short' })}</span>
          <span>{new Date(sortedData[sortedData.length - 1]?.timestamp || '').toLocaleDateString('de-DE', { day: '2-digit', month: 'short' })}</span>
        </div>
      </div>
    );
  };

  if (loading && !currentStats) {
    return (
      <div className="flex items-center justify-center h-64">
        <RefreshCw className="animate-spin text-blue-400" size={32} />
      </div>
    );
  }

  if (error) {
    return (
      <div className="bg-red-900/20 border border-red-800 text-red-400 px-4 py-3 rounded">
        {error}
      </div>
    );
  }

  const previousStats = getPreviousStats();

  return (
    <div className="p-6 space-y-6">
      {/* Header */}
      <div className="flex justify-between items-center">
        <div>
          <h1 className="text-3xl font-bold text-white">Discord Statistiken</h1>
          <p className="text-gray-400 mt-1">
            Überwache die Entwicklung deines Discord-Servers
          </p>
        </div>
        <button
          onClick={loadData}
          disabled={loading}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 text-white rounded-lg hover:bg-blue-700 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
        >
          <RefreshCw size={16} className={loading ? 'animate-spin' : ''} />
          Aktualisieren
        </button>
      </div>

      {/* Zeit-Range Auswahl */}
      <div className="flex gap-2">
        <button
          onClick={() => setTimeRange('day')}
          className={`px-4 py-2 rounded-lg transition-colors ${
            timeRange === 'day'
              ? 'bg-blue-600 text-white'
              : 'bg-[#1a1f2e] text-gray-400 hover:bg-[#252b3d] border border-gray-800'
          }`}
        >
          <Calendar size={16} className="inline mr-2" />
          24 Stunden
        </button>
        <button
          onClick={() => setTimeRange('week')}
          className={`px-4 py-2 rounded-lg transition-colors ${
            timeRange === 'week'
              ? 'bg-blue-600 text-white'
              : 'bg-[#1a1f2e] text-gray-400 hover:bg-[#252b3d] border border-gray-800'
          }`}
        >
          <Calendar size={16} className="inline mr-2" />
          7 Tage
        </button>
        <button
          onClick={() => setTimeRange('month')}
          className={`px-4 py-2 rounded-lg transition-colors ${
            timeRange === 'month'
              ? 'bg-blue-600 text-white'
              : 'bg-[#1a1f2e] text-gray-400 hover:bg-[#252b3d] border border-gray-800'
          }`}
        >
          <Calendar size={16} className="inline mr-2" />
          30 Tage
        </button>
        <button
          onClick={() => setTimeRange('total')}
          className={`px-4 py-2 rounded-lg transition-colors ${
            timeRange === 'total'
              ? 'bg-blue-600 text-white'
              : 'bg-[#1a1f2e] text-gray-400 hover:bg-[#252b3d] border border-gray-800'
          }`}
        >
          <Calendar size={16} className="inline mr-2" />
          Gesamt
        </button>
      </div>

      {/* Aktuelle Statistiken - Cards */}
      {currentStats && additionalStats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6">
          {renderStatCard(
            'Mitglieder',
            currentStats.member_count,
            <Users size={24} />,
            previousStats?.member_count
          )}
          {renderStatCard(
            'User Max',
            additionalStats.user_max,
            <Users size={24} />
          )}
          {renderStatCard(
            'Total Messages',
            additionalStats.total_messages,
            <MessageSquare size={24} />
          )}
          {renderStatCard(
            'Channels',
            currentStats.total_channels,
            <Hash size={24} />,
            previousStats?.total_channels
          )}
        </div>
      )}

      {/* Voice Statistiken */}
      {additionalStats && currentStats && (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          <div className="bg-[#1a1f2e] rounded-lg shadow-lg p-6 border border-gray-800">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-gray-400 text-sm font-medium">Total Voice Time</h3>
              <div className="text-blue-400"><Clock size={24} /></div>
            </div>
            <div className="text-3xl font-bold text-white">{formatSeconds(additionalStats.total_voice_time)}</div>
          </div>
          <div className="bg-[#1a1f2e] rounded-lg shadow-lg p-6 border border-gray-800">
            <div className="flex items-center justify-between mb-4">
              <h3 className="text-gray-400 text-sm font-medium">Ø Voice Time / Tag</h3>
              <div className="text-blue-400"><Clock size={24} /></div>
            </div>
            <div className="text-3xl font-bold text-white">{formatSeconds(additionalStats.avg_voice_time_day)}</div>
          </div>
          {renderStatCard(
            'User in Voice',
            currentStats.voice_user_count,
            <Volume2 size={24} />,
            previousStats?.voice_user_count
          )}
        </div>
      )}

      {/* Channel Details */}
      {currentStats && (
        <div className="bg-[#1a1f2e] rounded-lg shadow-lg p-6 border border-gray-800">
          <h3 className="text-lg font-semibold mb-4 text-white">Channel-Übersicht</h3>
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            <div className="flex items-center justify-between p-4 bg-[#252b3d] rounded border border-gray-700">
              <span className="text-gray-400">Text Channels</span>
              <span className="text-2xl font-bold text-white">{currentStats.text_channels}</span>
            </div>
            <div className="flex items-center justify-between p-4 bg-[#252b3d] rounded border border-gray-700">
              <span className="text-gray-400">Voice Channels</span>
              <span className="text-2xl font-bold text-white">{currentStats.voice_channels}</span>
            </div>
            <div className="flex items-center justify-between p-4 bg-[#252b3d] rounded border border-gray-700">
              <span className="text-gray-400">Kategorien</span>
              <span className="text-2xl font-bold text-white">{currentStats.category_channels}</span>
            </div>
          </div>
        </div>
      )}

      {/* Grafiken */}
      {historicalStats.length > 1 && (
        <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
          {renderLineChart(historicalStats, 'member_count', 'Mitglieder-Entwicklung')}
          {renderLineChart(historicalStats, 'total_voice_time', 'Voice-Time-Entwicklung (Stunden)', true)}
          {currentStats?.role_id && renderLineChart(historicalStats, 'role_member_count', 'Rollen-Mitglieder-Entwicklung')}
        </div>
      )}

      {/* Letzte Aktualisierung */}
      {currentStats && (
        <div className="text-center text-sm text-gray-500">
          Letzte Aktualisierung: {new Date(currentStats.timestamp).toLocaleString('de-DE')}
          {currentStats.source === 'scheduled' && ' (Automatisch)'}
          {currentStats.source === 'manual' && ' (Manuell)'}
        </div>
      )}
    </div>
  );
};

export default DiscordPage;
