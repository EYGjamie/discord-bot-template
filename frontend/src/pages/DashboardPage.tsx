import { useState, useEffect } from 'react';
import { Users, Calendar, ListTodo, TrendingUp } from 'lucide-react';
import StatCard from '../components/dashboard/StatCard';
import UpcomingEvents from '../components/dashboard/UpcomingEvents';
import UpcomingMatches from '../components/dashboard/UpcomingMatches';
import RecentActivity from '../components/dashboard/RecentActivity';
import ActiveMembers from '../components/dashboard/ActiveMembers';
import { api } from '../services/api';

interface DashboardStats {
  total_members: number;
  active_events: number;
  open_tasks: number;
  overdue_tasks: number;
  match_win_rate: number;
}

export default function DashboardPage() {
  const [stats, setStats] = useState<DashboardStats | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadStats();
  }, []);

  const loadStats = async () => {
    try {
      setLoading(true);
      const data = await api.dashboard.getStats();
      setStats(data);
    } catch (err) {
      console.error('Failed to load dashboard stats:', err);
    } finally {
      setLoading(false);
    }
  };

  const getTaskChangeText = () => {
    if (!stats) return '';
    if (stats.overdue_tasks > 0) {
      return `${stats.overdue_tasks} overdue`;
    }
    return 'No overdue tasks';
  };

  return (
    <div className="min-h-screen bg-[#0f1419] p-4 sm:p-6 pt-16 lg:pt-4 sm:pt-6">
      {/* Header */}
      <div className="mb-6 sm:mb-8">
        <h1 className="text-2xl sm:text-3xl font-bold text-white mb-2">Dashboard</h1>
        <p className="text-sm sm:text-base text-gray-400">Welcome back to Entropy Gaming HQ</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-4 sm:gap-6 mb-6">
        <StatCard
          title="Total Members"
          value={loading ? '...' : (stats?.total_members ?? 0)}
          change=""
          changeType="neutral"
          icon={Users}
          iconColor="bg-cyan-500"
        />
        <StatCard
          title="Active Events"
          value={loading ? '...' : (stats?.active_events ?? 0)}
          change=""
          changeType="neutral"
          icon={Calendar}
          iconColor="bg-blue-500"
        />
        <StatCard
          title="Open Tasks"
          value={loading ? '...' : (stats?.open_tasks ?? 0)}
          change={getTaskChangeText()}
          changeType={stats && stats.overdue_tasks > 0 ? 'negative' : 'neutral'}
          icon={ListTodo}
          iconColor="bg-green-500"
        />
        <StatCard
          title="Match Win Rate"
          value={`${stats?.match_win_rate ?? 72}%`}
          change="Coming soon"
          changeType="neutral"
          icon={TrendingUp}
          iconColor="bg-purple-500"
        />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 xl:grid-cols-3 gap-4 sm:gap-6">
        {/* Left Column - 2/3 width on large screens */}
        <div className="xl:col-span-2 space-y-4 sm:space-y-6">
          <UpcomingEvents />
          <UpcomingMatches />
          <RecentActivity />
        </div>

        {/* Right Column - 1/3 width on large screens */}
        <div className="xl:col-span-1">
          <ActiveMembers />
        </div>
      </div>
    </div>
  );
}
