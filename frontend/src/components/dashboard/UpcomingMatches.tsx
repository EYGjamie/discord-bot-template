import { useState, useEffect } from 'react';
import { Swords, Calendar as CalendarIcon, Clock, MapPin, Users } from 'lucide-react';
import { api } from '../../services/api';
import type { CalendarMatch } from '../../types';
import { format, parseISO, isAfter, startOfDay, isSameDay } from 'date-fns';

// Helper to format time without seconds (HH:MM:SS -> HH:MM)
const formatTime = (time: string | undefined): string => {
  if (!time) return '';
  return time.substring(0, 5); // Take only HH:MM
};

// Get Discord avatar URL
const getAvatarUrl = (avatar: string, userId: string) => {
  if (!avatar) return `https://cdn.discordapp.com/embed/avatars/${parseInt(userId) % 5}.png`;
  return `https://cdn.discordapp.com/avatars/${userId}/${avatar}.png`;
};

export default function UpcomingMatches() {
  const [matches, setMatches] = useState<CalendarMatch[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    loadMatches();
  }, []);

  const loadMatches = async () => {
    try {
      setLoading(true);
      const today = new Date();
      const month = today.getMonth() + 1;
      const year = today.getFullYear();
      const response = await api.matches.getMatches({ month, year });
      
      let matchList = [];
      if (Array.isArray(response)) {
        matchList = response;
      } else if (response && Array.isArray(response.matches)) {
        matchList = response.matches;
      } else if (response && response.data && Array.isArray(response.data)) {
        matchList = response.data;
      }

      // Filter to upcoming matches only (next 7 days)
      const todayStart = startOfDay(new Date());
      const upcomingMatches = matchList
        .filter((match: CalendarMatch) => {
          const matchStart = parseISO(match.start_date);
          return isAfter(matchStart, todayStart) || isSameDay(matchStart, todayStart);
        })
        .sort((a: CalendarMatch, b: CalendarMatch) => parseISO(a.start_date).getTime() - parseISO(b.start_date).getTime())
        .slice(0, 5);
      
      setMatches(upcomingMatches);
    } catch (err) {
      console.error('Failed to load matches:', err);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
        <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4 flex items-center gap-2">
          <Swords size={20} className="text-purple-400" />
          Upcoming Matches
        </h2>
        <div className="text-gray-400 text-center py-8">Loading matches...</div>
      </div>
    );
  }

  return (
    <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
      <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4 flex items-center gap-2">
        <Swords size={20} className="text-purple-400" />
        Upcoming Matches
      </h2>
      
      {matches.length === 0 ? (
        <div className="text-gray-400 text-center py-8 text-sm">
          No upcoming matches scheduled.
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3">
          {matches.map((match) => (
            <div
              key={match.id}
              className="p-3 sm:p-4 rounded-lg border-l-4 cursor-pointer transition-all shadow-md bg-gray-700/80 hover:bg-gray-700 hover:shadow-lg"
              style={{ borderLeftColor: match.color }}
            >
              <h3 className="text-white font-semibold text-sm mb-2 line-clamp-2">{match.title}</h3>
              <div className="space-y-1 text-xs text-gray-400">
                <div className="flex items-center gap-1">
                  <CalendarIcon size={12} />
                  <span>{format(parseISO(match.start_date), 'MMM d')}</span>
                </div>
                {!match.is_all_day && (
                  <div className="flex items-center gap-1">
                    <Clock size={12} />
                    <span>{formatTime(match.start_time)}</span>
                  </div>
                )}
                {match.is_all_day && (
                  <div className="flex items-center gap-1">
                    <Clock size={12} />
                    <span>All Day</span>
                  </div>
                )}
                {match.location && (
                  <div className="flex items-center gap-1">
                    <MapPin size={12} />
                    <span className="truncate">{match.location}</span>
                  </div>
                )}
                {match.guests && (
                  <div className="flex items-center gap-1">
                    <Users size={12} />
                    <span className="truncate">{match.guests}</span>
                  </div>
                )}
              </div>
              <div className="flex items-center gap-2 text-xs text-gray-500 mt-2 pt-2 border-t border-gray-600">
                <img
                  src={getAvatarUrl(match.creator_avatar, match.created_by)}
                  alt={match.creator_name}
                  className="w-4 h-4 rounded-full"
                />
                <span className="truncate">{match.creator_name}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
