import { Users, Calendar, ListTodo, TrendingUp } from 'lucide-react';
import StatCard from '../components/dashboard/StatCard';
import QuickActions from '../components/dashboard/QuickActions';
import UpcomingEvents from '../components/dashboard/UpcomingEvents';
import RecentActivity from '../components/dashboard/RecentActivity';
import ActiveMembers from '../components/dashboard/ActiveMembers';

export default function DashboardPage() {
  return (
    <div className="min-h-screen bg-[#0f1419] p-6">
      {/* Header */}
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-white mb-2">Dashboard</h1>
        <p className="text-gray-400">Welcome back to Entropy Gaming HQ</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-6">
        <StatCard
          title="Total Members"
          value={42}
          change="+3 this week"
          changeType="positive"
          icon={Users}
          iconColor="bg-cyan-500"
        />
        <StatCard
          title="Active Events"
          value={8}
          change="2 this week"
          changeType="neutral"
          icon={Calendar}
          iconColor="bg-blue-500"
        />
        <StatCard
          title="Open Tasks"
          value={15}
          change="5 overdue"
          changeType="negative"
          icon={ListTodo}
          iconColor="bg-green-500"
        />
        <StatCard
          title="Match Win Rate"
          value="72%"
          change="+8% from last month"
          changeType="positive"
          icon={TrendingUp}
          iconColor="bg-purple-500"
        />
      </div>

      {/* Quick Actions */}
      <div className="mb-6">
        <QuickActions />
      </div>

      {/* Main Content Grid */}
      <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
        {/* Left Column - 2/3 width */}
        <div className="lg:col-span-2 space-y-6">
          <UpcomingEvents />
          <RecentActivity />
        </div>

        {/* Right Column - 1/3 width */}
        <div className="lg:col-span-1">
          <ActiveMembers />
        </div>
      </div>
    </div>
  );
}
