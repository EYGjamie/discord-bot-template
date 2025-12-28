import type { Activity } from '../../types';

const mockActivities: Activity[] = [
  {
    id: '1',
    type: 'task',
    user_id: 'AL',
    message: 'completed task Update roster page',
    timestamp: '2 min ago'
  },
  {
    id: '2',
    type: 'event',
    user_id: 'MA',
    message: 'scheduled event Weekly Scrim vs Team Alpha',
    timestamp: '15 min ago'
  },
  {
    id: '3',
    type: 'member',
    user_id: 'JA',
    message: 'added member NightOwl#1234',
    timestamp: '1 hour ago'
  },
  {
    id: '4',
    type: 'match',
    user_id: 'SA',
    message: 'won match Ranked 5v5 - Valorant',
    timestamp: '3 hours ago'
  },
  {
    id: '5',
    type: 'sync',
    user_id: 'BO',
    message: 'synced data Discord member roles',
    timestamp: '5 hours ago'
  },
];

const activityTypeColors: Record<string, string> = {
  task: 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20',
  event: 'bg-amber-500/10 text-amber-400 border-amber-500/20',
  member: 'bg-green-500/10 text-green-400 border-green-500/20',
  match: 'bg-purple-500/10 text-purple-400 border-purple-500/20',
  sync: 'bg-blue-500/10 text-blue-400 border-blue-500/20',
};

export default function RecentActivity() {
  return (
    <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
      <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4">Recent Activity</h2>
      <div className="space-y-3">
        {mockActivities.map((activity) => (
          <div key={activity.id} className="flex items-start gap-2 sm:gap-3">
            <div className="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-gray-700 flex items-center justify-center text-white text-xs font-medium flex-shrink-0">
              {activity.user_id}
            </div>
            <div className="flex-1 min-w-0">
              <p className="text-gray-300 text-xs sm:text-sm">
                <span className="font-medium text-white">{activity.user_id}</span>{' '}
                {activity.message}
              </p>
              <p className="text-gray-500 text-xs mt-1">{activity.timestamp}</p>
            </div>
            <span className={`hidden sm:flex px-2 py-1 text-xs font-medium rounded border ${activityTypeColors[activity.type]} flex-shrink-0`}>
              {activity.type.charAt(0).toUpperCase() + activity.type.slice(1)}
            </span>
          </div>
        ))}
      </div>
    </div>
  );
}
