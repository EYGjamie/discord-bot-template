import { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { formatDistanceToNow } from 'date-fns';

interface ActivityItem {
  id: number;
  user_id: string;
  user_name: string;
  user_avatar: string;
  action_type: string;
  resource_type: string;
  resource_id: string;
  timestamp: string;
}

// Get Discord avatar URL
const getAvatarUrl = (avatar: string, userId: string): string => {
  if (avatar) return `https://cdn.discordapp.com/avatars/${userId}/${avatar}.png?size=64`;
  return `https://cdn.discordapp.com/embed/avatars/${parseInt(userId) % 5}.png`;
};

// Map action types to display strings
const getActionMessage = (activity: ActivityItem): string => {
  const resourceType = activity.resource_type.toLowerCase();
  
  switch (activity.action_type) {
    case 'CREATE':
      if (resourceType === 'event') return 'created an event';
      if (resourceType === 'task') return 'created a task';
      if (resourceType === 'board') return 'created a board';
      if (resourceType === 'match') return 'created a match';
      return `created ${resourceType}`;
    case 'UPDATE':
      if (resourceType === 'event') return 'updated an event';
      if (resourceType === 'task') return 'updated a task';
      if (resourceType === 'board') return 'updated a board';
      if (resourceType === 'match') return 'updated a match';
      return `updated ${resourceType}`;
    case 'DELETE':
      if (resourceType === 'event') return 'deleted an event';
      if (resourceType === 'task') return 'deleted a task';
      if (resourceType === 'board') return 'deleted a board';
      if (resourceType === 'match') return 'deleted a match';
      return `deleted ${resourceType}`;
    case 'WARN_CREATE':
      return 'issued a warning';
    case 'NOTE_CREATE':
      return 'added a note';
    case 'LOGIN':
      return 'logged in';
    default:
      return activity.action_type.toLowerCase().replace('_', ' ');
  }
};

// Map action types to colors
const getActivityColor = (actionType: string): string => {
  switch (actionType) {
    case 'CREATE':
      return 'bg-green-500/10 text-green-400 border-green-500/20';
    case 'UPDATE':
      return 'bg-blue-500/10 text-blue-400 border-blue-500/20';
    case 'DELETE':
      return 'bg-red-500/10 text-red-400 border-red-500/20';
    case 'WARN_CREATE':
      return 'bg-amber-500/10 text-amber-400 border-amber-500/20';
    case 'NOTE_CREATE':
      return 'bg-purple-500/10 text-purple-400 border-purple-500/20';
    case 'LOGIN':
      return 'bg-cyan-500/10 text-cyan-400 border-cyan-500/20';
    default:
      return 'bg-gray-500/10 text-gray-400 border-gray-500/20';
  }
};

const getActionLabel = (actionType: string): string => {
  switch (actionType) {
    case 'CREATE':
      return 'Created';
    case 'UPDATE':
      return 'Updated';
    case 'DELETE':
      return 'Deleted';
    case 'WARN_CREATE':
      return 'Warning';
    case 'NOTE_CREATE':
      return 'Note';
    case 'LOGIN':
      return 'Login';
    default:
      return actionType.charAt(0).toUpperCase() + actionType.slice(1).toLowerCase();
  }
};

export default function RecentActivity() {
  const [activities, setActivities] = useState<ActivityItem[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadActivity();
  }, []);

  const loadActivity = async () => {
    try {
      const response = await api.dashboard.getRecentActivity();
      setActivities(response.activities || []);
    } catch (err) {
      console.error('Failed to load recent activity:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatTimestamp = (timestamp: string): string => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'recently';
    }
  };

  if (loading) {
    return (
      <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
        <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4">Recent Activity</h2>
        <div className="text-gray-400 text-center py-4 text-sm">Loading...</div>
      </div>
    );
  }

  // Calculate height for 5 items (each item is approximately 52px)
  const maxVisibleHeight = 5 * 52;

  return (
    <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
      <div className="flex items-center justify-between mb-3 sm:mb-4">
        <h2 className="text-white text-base sm:text-lg font-semibold">Recent Activity</h2>
        {activities.length > 5 && (
          <span className="text-xs text-gray-500">
            {activities.length} entries
          </span>
        )}
      </div>
      
      {activities.length === 0 ? (
        <div className="text-gray-400 text-center py-8 text-sm">
          No recent activity.
        </div>
      ) : (
        <div 
          className="space-y-3 overflow-y-auto scrollbar-thin scrollbar-thumb-gray-700 scrollbar-track-transparent pr-1"
          style={{ maxHeight: `${maxVisibleHeight}px` }}
        >
          {activities.map((activity) => (
            <div key={activity.id} className="flex items-start gap-2 sm:gap-3">
              <img
                src={getAvatarUrl(activity.user_avatar, activity.user_id)}
                alt={activity.user_name}
                className="w-7 h-7 sm:w-8 sm:h-8 rounded-full bg-gray-700 flex-shrink-0"
              />
              <div className="flex-1 min-w-0">
                <p className="text-gray-300 text-xs sm:text-sm">
                  <span className="font-medium text-white">{activity.user_name}</span>{' '}
                  {getActionMessage(activity)}
                </p>
                <p className="text-gray-500 text-xs mt-1">{formatTimestamp(activity.timestamp)}</p>
              </div>
              <span className={`hidden sm:flex px-2 py-1 text-xs font-medium rounded border ${getActivityColor(activity.action_type)} flex-shrink-0`}>
                {getActionLabel(activity.action_type)}
              </span>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
