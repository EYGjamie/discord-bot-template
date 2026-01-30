import { useState, useEffect } from 'react';
import { api } from '../../services/api';
import { formatDistanceToNow } from 'date-fns';

interface ActiveUser {
  id: string;
  display_name: string;
  avatar: string;
  avatar_url: string;
  top_role: string;
  top_role_name: string | null;
  top_role_color: string | null;
  last_active: string;
  is_online: boolean;
}

// Get Discord avatar URL
const getAvatarUrl = (avatar: string, avatarUrl: string, userId: string): string => {
  if (avatarUrl) return avatarUrl;
  if (avatar) return `https://cdn.discordapp.com/avatars/${userId}/${avatar}.png?size=64`;
  return `https://cdn.discordapp.com/embed/avatars/${parseInt(userId) % 5}.png`;
};

export default function ActiveMembers() {
  const [users, setUsers] = useState<ActiveUser[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadActiveUsers();
    // Refresh every 30 seconds
    const interval = setInterval(loadActiveUsers, 30000);
    return () => clearInterval(interval);
  }, []);

  const loadActiveUsers = async () => {
    try {
      const response = await api.dashboard.getActiveUsers();
      setUsers(response.users || []);
    } catch (err) {
      console.error('Failed to load active users:', err);
    } finally {
      setLoading(false);
    }
  };

  const formatLastSeen = (timestamp: string): string => {
    try {
      return formatDistanceToNow(new Date(timestamp), { addSuffix: true });
    } catch {
      return 'recently';
    }
  };

  if (loading) {
    return (
      <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
        <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4">Active Members</h2>
        <div className="text-gray-400 text-center py-4 text-sm">Loading...</div>
      </div>
    );
  }

  return (
    <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
      <div className="flex items-center justify-between mb-3 sm:mb-4">
        <h2 className="text-white text-base sm:text-lg font-semibold">Active Members</h2>
      </div>
      
      {users.length === 0 ? (
        <div className="text-gray-400 text-center py-8 text-sm">
          No active members right now.
        </div>
      ) : (
        <div className="space-y-3">
          {users.map((user) => (
            <div key={user.id} className="flex items-center gap-2 sm:gap-3">
              <div className="relative flex-shrink-0">
                <img
                  src={getAvatarUrl(user.avatar, user.avatar_url, user.id)}
                  alt={user.display_name}
                  className="w-9 h-9 sm:w-10 sm:h-10 rounded-full bg-gray-700"
                />
                <div 
                  className={`absolute bottom-0 right-0 w-2.5 h-2.5 sm:w-3 sm:h-3 rounded-full border-2 border-[#1a1f2e] ${
                    user.is_online ? 'bg-green-500' : 'bg-gray-500'
                  }`} 
                />
              </div>
              <div className="flex-1 min-w-0">
                <p className="text-white font-medium text-xs sm:text-sm truncate">{user.display_name}</p>
                {user.is_online ? (
                  user.top_role_name && (
                    <p 
                      className="text-xs truncate"
                      style={{ color: user.top_role_color || '#9ca3af' }}
                    >
                      {user.top_role_name}
                    </p>
                  )
                ) : (
                  <p className="text-xs text-gray-500 truncate">
                    {formatLastSeen(user.last_active)}
                  </p>
                )}
              </div>
              {!user.is_online && user.top_role_name && (
                <span 
                  className="hidden sm:block text-xs px-2 py-0.5 rounded truncate max-w-[80px]"
                  style={{ 
                    color: user.top_role_color || '#9ca3af',
                    backgroundColor: `${user.top_role_color || '#9ca3af'}15`
                  }}
                >
                  {user.top_role_name}
                </span>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
