import { useState, useEffect } from 'react';
import Calendar from 'react-calendar';
import 'react-calendar/dist/Calendar.css';
import { api } from '../services/api';
import type { CalendarMatch } from '../types';
import { format, parseISO, isSameDay, isAfter, startOfDay } from 'date-fns';
import { Calendar as CalendarIcon, Clock, MapPin, Plus, Edit2, Trash2, Users, Settings, X, Save } from 'lucide-react';
import MatchModal from '../components/matches/MatchModal';
import { useAuth } from '../hooks/useAuth';
import { usePermissions } from '../hooks/usePermissions';

interface MatchCategory {
  id: number;
  name: string;
  color: string;
  sort_order: number;
}

export default function MatchesPage() {
  const { user } = useAuth();
  const permissions = usePermissions(user);
  
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [matches, setMatches] = useState<CalendarMatch[]>([]);
  const [filteredMatches, setFilteredMatches] = useState<CalendarMatch[]>([]);
  const [selectedMatch, setSelectedMatch] = useState<CalendarMatch | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingMatch, setEditingMatch] = useState<CalendarMatch | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [categories, setCategories] = useState<MatchCategory[]>([]);
  const [selectedColors, setSelectedColors] = useState<Set<string>>(new Set());
  const [showCategoryManager, setShowCategoryManager] = useState(false);
  const [editingCategory, setEditingCategory] = useState<MatchCategory | null>(null);

  const currentUserId = user?.discord_id || '';
  const isAdmin = permissions.isAdmin;

  // Load categories on mount
  useEffect(() => {
    loadCategories();
  }, []);

  // Load matches for selected month
  useEffect(() => {
    loadMatches();
  }, [selectedDate]);

  // Initialize selected colors when categories load
  useEffect(() => {
    if (categories.length > 0) {
      const colors = new Set(categories.map(c => c.color));
      console.log('Initializing selectedColors:', Array.from(colors));
      setSelectedColors(colors);
    }
  }, [categories]);

  // Filter matches for selected date
  useEffect(() => {
    const dateStr = format(selectedDate, 'yyyy-MM-dd');
    console.log('Filtering matches for date:', dateStr);
    console.log('Total matches:', matches.length);
    console.log('Selected colors:', Array.from(selectedColors));
    
    const dayMatches = matches.filter(match => {
      // Parse dates and extract only the date part (ignore time and timezone)
      const matchStartDate = format(parseISO(match.start_date), 'yyyy-MM-dd');
      const matchEndDate = format(parseISO(match.end_date), 'yyyy-MM-dd');
      
      // Compare only the date strings
      const inRange = dateStr >= matchStartDate && dateStr <= matchEndDate;
      const colorMatch = selectedColors.has(match.color);
      
      console.log(`Match "${match.title}": inRange=${inRange}, color=${match.color}, colorMatch=${colorMatch}, start=${matchStartDate}, end=${matchEndDate}`);
      
      return inRange && colorMatch;
    });
    
    console.log('Filtered matches:', dayMatches.length);
    setFilteredMatches(dayMatches);
    
    // Reset selected match if it's not in the current day
    if (selectedMatch && !dayMatches.find(m => m.id === selectedMatch.id)) {
      setSelectedMatch(null);
    }
  }, [selectedDate, matches, selectedColors]);

  const loadMatches = async () => {
    try {
      setLoading(true);
      setError(null);
      const month = selectedDate.getMonth() + 1;
      const year = selectedDate.getFullYear();
      const response = await api.matches.getMatches({ month, year });
      console.log('API Response:', response);
      console.log('Response type:', typeof response);
      console.log('Response.matches:', response.matches);
      
      let matchList = [];
      if (Array.isArray(response)) {
        matchList = response;
      } else if (response && Array.isArray(response.matches)) {
        matchList = response.matches;
      } else if (response && response.data && Array.isArray(response.data)) {
        matchList = response.data;
      }
      
      console.log('Parsed matches:', matchList);
      console.log('Match count:', matchList.length);
      setMatches(matchList);
    } catch (err) {
      console.error('Failed to load matches:', err);
      setError('Failed to load matches. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleDateChange = (value: any) => {
    if (value instanceof Date) {
      setSelectedDate(value);
    }
  };

  const loadCategories = async () => {
    try {
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      const response = await fetch(`${API_BASE_URL}/api/match-categories`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) {
        throw new Error('Failed to load categories');
      }
      
      const data = await response.json();
      // Backend returns {categories: [...]}
      setCategories(data.categories || data || []);
    } catch (err) {
      console.error('Failed to load categories:', err);
      // Use default colors if categories fail to load
      setCategories([
        { id: 1, name: 'Competitive', color: '#E53935', sort_order: 1 },
        { id: 2, name: 'Casual', color: '#43A047', sort_order: 2 },
        { id: 3, name: 'Training', color: '#FB8C00', sort_order: 3 },
        { id: 4, name: 'Tournament', color: '#8E24AA', sort_order: 4 },
        { id: 5, name: 'Scrim', color: '#3949AB', sort_order: 5 },
        { id: 6, name: 'League', color: '#00ACC1', sort_order: 6 },
        { id: 7, name: 'Friendly', color: '#FDD835', sort_order: 7 },
        { id: 8, name: 'Championship', color: '#D81B60', sort_order: 8 },
      ]);
    }
  };

  const handleCreateMatch = () => {
    setEditingMatch(null);
    setIsModalOpen(true);
  };

  const handleEditMatch = (match: CalendarMatch) => {
    setEditingMatch(match);
    setIsModalOpen(true);
  };

  const handleDeleteMatch = async (matchId: number) => {
    if (!confirm('Are you sure you want to delete this match?')) return;
    
    try {
      await api.matches.deleteMatch(matchId);
      await loadMatches();
      if (selectedMatch?.id === matchId) {
        setSelectedMatch(null);
      }
    } catch (err) {
      console.error('Failed to delete match:', err);
      alert('Failed to delete match. Please try again.');
    }
  };

  const handleSaveMatch = async () => {
    setIsModalOpen(false);
    setEditingMatch(null);
    await loadMatches();
  };

  const canEditMatch = (match: CalendarMatch) => {
    return isAdmin || match.created_by === currentUserId;
  };

  const toggleColor = (color: string) => {
    const newColors = new Set(selectedColors);
    if (newColors.has(color)) {
      newColors.delete(color);
    } else {
      newColors.add(color);
    }
    setSelectedColors(newColors);
  };

  // Get matches for a specific date
  const getMatchesForDate = (date: Date): CalendarMatch[] => {
    const dateStr = format(date, 'yyyy-MM-dd');
    const filtered = matches.filter(match => {
      if (!match.start_date || !match.end_date) {
        console.warn('Match missing dates:', match);
        return false;
      }
      
      try {
        // Parse dates and extract only the date part (ignore time and timezone)
        const matchStartDate = format(parseISO(match.start_date), 'yyyy-MM-dd');
        const matchEndDate = format(parseISO(match.end_date), 'yyyy-MM-dd');
        
        // Compare only the date strings
        const inRange = dateStr >= matchStartDate && dateStr <= matchEndDate;
        const colorMatch = selectedColors.has(match.color);
        
        return inRange && colorMatch;
      } catch (err) {
        console.error('Error parsing match dates:', match, err);
        return false;
      }
    });
    
    if (filtered.length > 0 && dateStr === format(new Date(), 'yyyy-MM-dd')) {
      console.log('Matches for today:', filtered);
    }
    
    return filtered;
  };

  // Get upcoming matches (next 7 days)
  const getUpcomingMatches = (): CalendarMatch[] => {
    const today = startOfDay(new Date());
    return matches
      .filter(match => {
        const matchStart = parseISO(match.start_date);
        return isAfter(matchStart, today) || isSameDay(matchStart, today);
      })
      .filter(match => selectedColors.has(match.color))
      .sort((a, b) => parseISO(a.start_date).getTime() - parseISO(b.start_date).getTime())
      .slice(0, 5);
  };

  // Render matches in calendar tiles
  const tileContent = ({ date }: { date: Date }) => {
    const dayMatches = getMatchesForDate(date);
    if (dayMatches.length === 0) return null;

    return (
      <div className="match-indicators">
        {dayMatches.slice(0, 2).map((match) => (
          <div
            key={match.id}
            className="match-dot"
            style={{ backgroundColor: match.color }}
            title={match.title}
          >
            <img
              src={getAvatarUrl(match.creator_avatar, match.created_by)}
              alt={match.creator_name}
              className="match-creator-avatar"
            />
            <span className="match-title-preview">{match.title}</span>
          </div>
        ))}
        {dayMatches.length > 2 && (
          <div className="match-more">+{dayMatches.length - 2}</div>
        )}
      </div>
    );
  };

  // Highlight today's date
  const tileClassName = ({ date }: { date: Date }) => {
    return isSameDay(date, new Date()) ? 'today-tile' : '';
  };

  // Format match date range
  const formatDateRange = (match: CalendarMatch) => {
    const start = parseISO(match.start_date);
    const end = parseISO(match.end_date);
    
    if (format(start, 'yyyy-MM-dd') === format(end, 'yyyy-MM-dd')) {
      return format(start, 'MMMM d, yyyy');
    }
    
    return `${format(start, 'MMM d')} - ${format(end, 'MMM d, yyyy')}`;
  };

  // Get Discord avatar URL
  const getAvatarUrl = (avatar: string, userId: string) => {
    if (!avatar) return `https://cdn.discordapp.com/embed/avatars/${parseInt(userId) % 5}.png`;
    return `https://cdn.discordapp.com/avatars/${userId}/${avatar}.png`;
  };

  const upcomingMatches = getUpcomingMatches();

  return (
    <div className="p-4 sm:p-6 pt-16 lg:pt-4 sm:pt-6">
      <div className="mb-4 sm:mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
        <h1 className="text-2xl sm:text-3xl font-bold text-white">Matches Calendar</h1>
        <button
          onClick={handleCreateMatch}
          className="flex items-center gap-2 px-3 sm:px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm sm:text-base w-full sm:w-auto justify-center"
        >
          <Plus size={18} className="sm:w-5 sm:h-5" />
          <span>Create Match</span>
        </button>
      </div>

      {error && (
        <div className="mb-4 p-3 sm:p-4 bg-red-500/10 border border-red-500/20 rounded-lg text-red-400 text-sm sm:text-base">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 xl:grid-cols-5 gap-4 sm:gap-6">
        {/* Calendar Section */}
        <div className="xl:col-span-3">
          <div className="bg-gray-800 rounded-lg p-3 sm:p-6 shadow-lg">
            <Calendar
              onChange={handleDateChange}
              value={selectedDate}
              tileContent={tileContent}
              tileClassName={tileClassName}
              className="react-calendar-dark"
            />
            
            {/* Compact Color Legend */}
            <div className="mt-4 sm:mt-6 pt-3 sm:pt-4 border-t border-gray-700">
              <div className="flex items-center justify-between mb-2">
                <h3 className="text-white font-semibold text-xs sm:text-sm">Match Categories</h3>
                {isAdmin && (
                  <button
                    onClick={() => setShowCategoryManager(true)}
                    className="flex items-center gap-1.5 px-2 sm:px-3 py-1 sm:py-1.5 bg-gray-700 hover:bg-gray-600 rounded-lg transition-colors text-xs sm:text-sm"
                    title="Manage Categories"
                  >
                    <Settings size={14} className="sm:w-4 sm:h-4 text-blue-400" />
                    <span className="text-gray-300 hidden sm:inline">Manage</span>
                  </button>
                )}
              </div>
              <div className="flex flex-wrap gap-1.5 sm:gap-2">
                {categories.map((category) => (
                  <button
                    key={category.id}
                    onClick={() => toggleColor(category.color)}
                    className={`flex items-center gap-1 sm:gap-1.5 px-1.5 sm:px-2 py-0.5 sm:py-1 rounded-full text-[10px] sm:text-xs transition-all ${
                      selectedColors.has(category.color)
                        ? 'bg-gray-700 text-white'
                        : 'bg-gray-750 text-gray-500 opacity-60'
                    }`}
                  >
                    <div
                      className="w-2 h-2 sm:w-2.5 sm:h-2.5 rounded-full flex-shrink-0"
                      style={{ backgroundColor: category.color }}
                    />
                    <span className="truncate">{category.name}</span>
                  </button>
                ))}
              </div>
            </div>
          </div>
        </div>

        {/* Matches List Section */}
        <div className="xl:col-span-2">
          <div className="bg-gray-800 rounded-lg p-3 sm:p-6 shadow-lg">
            <div className="flex items-center gap-2 sm:gap-3 mb-3 sm:mb-4">
              <CalendarIcon className="text-blue-400" size={20} />
              <h2 className="text-lg sm:text-xl font-semibold text-white">
                {format(selectedDate, 'MMM d, yyyy')}
              </h2>
            </div>

            {loading ? (
              <div className="text-gray-400 text-center py-8 text-sm sm:text-base">Loading matches...</div>
            ) : filteredMatches.length === 0 ? (
              <div className="text-gray-400 text-center py-8 text-sm sm:text-base">
                No matches scheduled for this day.
              </div>
            ) : (
              <div className="space-y-3 max-h-[500px] sm:max-h-[600px] overflow-y-auto">
                {filteredMatches.map((match) => (
                  <div
                    key={match.id}
                    onClick={() => setSelectedMatch(match)}
                    className={`p-3 sm:p-4 rounded-lg border-l-4 cursor-pointer transition-all shadow-md ${
                      selectedMatch?.id === match.id
                        ? 'bg-gray-700 border-blue-500 shadow-lg shadow-blue-500/20'
                        : 'bg-gray-700/80 hover:bg-gray-700 border-gray-500 hover:shadow-lg'
                    }`}
                    style={{ borderLeftColor: match.color }}
                  >
                    <div className="flex justify-between items-start mb-2">
                      <h3 className="text-white font-semibold text-sm sm:text-lg">{match.title}</h3>
                      {canEditMatch(match) && (
                        <div className="flex gap-2">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleEditMatch(match);
                            }}
                            className="p-1 hover:bg-gray-600 rounded transition-colors"
                          >
                            <Edit2 size={16} className="text-blue-400" />
                          </button>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteMatch(match.id);
                            }}
                            className="p-1 hover:bg-gray-600 rounded transition-colors"
                          >
                            <Trash2 size={16} className="text-red-400" />
                          </button>
                        </div>
                      )}
                    </div>
                    
                    <div className="space-y-1 text-sm text-gray-400">
                      {!match.is_all_day && (
                        <div className="flex items-center gap-1">
                          <Clock size={14} />
                          <span>{match.start_time} - {match.end_time}</span>
                        </div>
                      )}
                      {match.is_all_day && (
                        <div className="flex items-center gap-1">
                          <Clock size={14} />
                          <span>Ganztägig</span>
                        </div>
                      )}
                      {match.location && (
                        <div className="flex items-center gap-1">
                          <MapPin size={14} />
                          <span>{match.location}</span>
                        </div>
                      )}
                      {match.guests && (
                        <div className="flex items-center gap-1">
                          <Users size={14} />
                          <span>{match.guests}</span>
                        </div>
                      )}
                    </div>

                    <div className="flex items-center gap-2 text-xs text-gray-500 mt-2">
                      <img
                        src={getAvatarUrl(match.creator_avatar, match.created_by)}
                        alt={match.creator_name}
                        className="w-5 h-5 rounded-full"
                      />
                      <span>{match.creator_name || 'Unknown'}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Match Details Section */}
            {selectedMatch && (
              <div className="mt-6 p-4 bg-gray-750 rounded-lg border border-gray-700">
                <h3 className="text-white font-semibold text-lg mb-3">Match Details</h3>
                <div className="space-y-2 text-gray-300">
                  {selectedMatch.description && (
                    <div>
                      <p className="text-sm text-gray-400">Description</p>
                      <p className="whitespace-pre-wrap">{selectedMatch.description}</p>
                    </div>
                  )}
                  <div>
                    <p className="text-sm text-gray-400">Date</p>
                    <p>{formatDateRange(selectedMatch)}</p>
                  </div>
                  {!selectedMatch.is_all_day && (
                    <div>
                      <p className="text-sm text-gray-400">Time</p>
                      <p>{selectedMatch.start_time} - {selectedMatch.end_time}</p>
                    </div>
                  )}
                  {selectedMatch.is_all_day && (
                    <div>
                      <p className="text-sm text-gray-400">Time</p>
                      <p>Ganztägig</p>
                    </div>
                  )}
                  {selectedMatch.location && (
                    <div>
                      <p className="text-sm text-gray-400">Location</p>
                      <p>{selectedMatch.location}</p>
                    </div>
                  )}
                  {selectedMatch.guests && (
                    <div>
                      <p className="text-sm text-gray-400">Guests</p>
                      <p>{selectedMatch.guests}</p>
                    </div>
                  )}
                  <div>
                    <p className="text-sm text-gray-400">Created by</p>
                    <div className="flex items-center gap-2 mt-1">
                      <img
                        src={getAvatarUrl(selectedMatch.creator_avatar, selectedMatch.created_by)}
                        alt={selectedMatch.creator_name}
                        className="w-6 h-6 rounded-full"
                      />
                      <span>{selectedMatch.creator_name || 'Unknown'}</span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Upcoming Matches Section */}
      {upcomingMatches.length > 0 && (
        <div className="mt-6 bg-gray-800 rounded-lg p-6 shadow-lg">
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <CalendarIcon size={24} className="text-blue-400" />
            Upcoming Matches
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
            {upcomingMatches.map((match) => (
              <div
                key={match.id}
                onClick={() => {
                  setSelectedDate(parseISO(match.start_date));
                  setSelectedMatch(match);
                }}
                className="p-4 rounded-lg border-l-4 cursor-pointer transition-all shadow-md bg-gray-700/80 hover:bg-gray-700 border-gray-500 hover:shadow-lg"
                style={{ borderLeftColor: match.color }}
              >
                <h3 className="text-white font-semibold mb-2">{match.title}</h3>
                <div className="space-y-1 text-sm text-gray-400">
                  <div className="flex items-center gap-1">
                    <CalendarIcon size={12} />
                    <span>{format(parseISO(match.start_date), 'MMM d')}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Clock size={12} />
                    <span>{match.start_time}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-xs text-gray-500 mt-2">
                  <img
                    src={getAvatarUrl(match.creator_avatar, match.created_by)}
                    alt={match.creator_name}
                    className="w-4 h-4 rounded-full"
                  />
                  <span>{match.creator_name}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      )}

      {/* Category Manager Modal */}
      {showCategoryManager && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-gray-800 rounded-lg p-6 w-full max-w-2xl max-h-[80vh] overflow-y-auto">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-xl font-bold text-white">Manage Match Categories</h2>
              <button
                onClick={() => {
                  setShowCategoryManager(false);
                  setEditingCategory(null);
                }}
                className="p-1 hover:bg-gray-700 rounded transition-colors"
              >
                <X size={20} className="text-gray-400" />
              </button>
            </div>

            <div className="space-y-3 mb-4">
              {categories.map((category) => (
                <div key={category.id} className="flex items-center gap-3 bg-gray-750 p-3 rounded-lg">
                  {editingCategory?.id === category.id ? (
                    <>
                      <input
                        type="text"
                        value={editingCategory.name}
                        onChange={(e) => setEditingCategory({ ...editingCategory, name: e.target.value })}
                        className="flex-1 bg-gray-700 text-white px-3 py-2 rounded"
                        placeholder="Category name"
                      />
                      <input
                        type="color"
                        value={editingCategory.color}
                        onChange={(e) => setEditingCategory({ ...editingCategory, color: e.target.value })}
                        className="w-12 h-10 rounded cursor-pointer"
                      />
                      <button
                        onClick={async () => {
                          try {
                            const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
                            await fetch(`${API_BASE_URL}/api/match-categories/${category.id}`, {
                              method: 'PUT',
                              headers: {
                                'Content-Type': 'application/json',
                                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                              },
                              body: JSON.stringify({
                                name: editingCategory.name,
                                color: editingCategory.color,
                                sort_order: category.sort_order,
                              }),
                            });
                            await loadCategories();
                            setEditingCategory(null);
                          } catch (err) {
                            console.error('Failed to update category:', err);
                          }
                        }}
                        className="p-2 bg-green-600 hover:bg-green-700 rounded transition-colors"
                      >
                        <Save size={16} className="text-white" />
                      </button>
                      <button
                        onClick={() => setEditingCategory(null)}
                        className="p-2 bg-gray-600 hover:bg-gray-700 rounded transition-colors"
                      >
                        <X size={16} className="text-white" />
                      </button>
                    </>
                  ) : (
                    <>
                      <div
                        className="w-6 h-6 rounded"
                        style={{ backgroundColor: category.color }}
                      />
                      <span className="flex-1 text-white">{category.name}</span>
                      <button
                        onClick={() => setEditingCategory(category)}
                        className="p-2 hover:bg-gray-700 rounded transition-colors"
                      >
                        <Edit2 size={16} className="text-blue-400" />
                      </button>
                      <button
                        onClick={async () => {
                          if (!confirm('Delete this category?')) return;
                          try {
                            const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
                            await fetch(`${API_BASE_URL}/api/match-categories/${category.id}`, {
                              method: 'DELETE',
                              headers: {
                                'Authorization': `Bearer ${localStorage.getItem('token')}`,
                              },
                            });
                            await loadCategories();
                          } catch (err) {
                            console.error('Failed to delete category:', err);
                          }
                        }}
                        className="p-2 hover:bg-gray-700 rounded transition-colors"
                      >
                        <Trash2 size={16} className="text-red-400" />
                      </button>
                    </>
                  )}
                </div>
              ))}
            </div>

            <button
              onClick={async () => {
                const name = prompt('New category name:');
                if (!name) return;
                const color = prompt('Color hex code:', '#E53935');
                if (!color) return;
                
                try {
                  const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
                  await fetch(`${API_BASE_URL}/api/match-categories`, {
                    method: 'POST',
                    headers: {
                      'Content-Type': 'application/json',
                      'Authorization': `Bearer ${localStorage.getItem('token')}`,
                    },
                    body: JSON.stringify({
                      name,
                      color,
                      sort_order: categories.length + 1,
                    }),
                  });
                  await loadCategories();
                } catch (err) {
                  console.error('Failed to create category:', err);
                }
              }}
              className="w-full py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
              + Add Category
            </button>
          </div>
        </div>
      )}

      {/* Match Modal */}
      {isModalOpen && (
        <MatchModal
          match={editingMatch}
          defaultDate={format(selectedDate, 'yyyy-MM-dd')}
          onClose={() => {
            setIsModalOpen(false);
            setEditingMatch(null);
          }}
          onSave={handleSaveMatch}
          categories={categories}
        />
      )}

      <style>{`
        .react-calendar-dark {
          background: transparent;
          border: none;
          color: white;
          width: 100%;
          font-size: 16px;
        }
        
        .react-calendar-dark .react-calendar__tile {
          color: #d1d5db;
          border-radius: 8px;
          padding: 20px 10px;
          min-height: 100px;
          position: relative;
          display: flex;
          flex-direction: column;
          align-items: center;
        }
        
        .react-calendar-dark .react-calendar__tile:enabled:hover,
        .react-calendar-dark .react-calendar__tile:enabled:focus {
          background-color: #374151;
        }
        
        .react-calendar-dark .react-calendar__tile--active {
          background-color: #3b82f6 !important;
          color: white;
        }
        
        .react-calendar-dark .today-tile {
          background-color: #1e40af !important;
          font-weight: bold;
          box-shadow: 0 0 0 2px #3b82f6 inset;
        }
        
        .react-calendar-dark .react-calendar__navigation button {
          color: white;
          font-size: 18px;
          min-width: 44px;
        }
        
        .react-calendar-dark .react-calendar__navigation button:enabled:hover,
        .react-calendar-dark .react-calendar__navigation button:enabled:focus {
          background-color: #374151;
        }
        
        .react-calendar-dark abbr[title] {
          text-decoration: none;
        }
        
        .react-calendar-dark .react-calendar__month-view__weekdays {
          font-size: 14px;
          font-weight: 600;
          color: #9ca3af;
        }
        
        .match-indicators {
          display: flex;
          flex-direction: column;
          gap: 2px;
          width: 100%;
          margin-top: 4px;
        }
        
        .match-dot {
          font-size: 10px;
          padding: 2px 4px;
          border-radius: 3px;
          overflow: hidden;
          text-overflow: ellipsis;
          white-space: nowrap;
          text-align: left;
          display: flex;
          align-items: center;
          gap: 3px;
        }
        
        .match-creator-avatar {
          width: 14px;
          height: 14px;
          border-radius: 50%;
          flex-shrink: 0;
          border: 1px solid rgba(255, 255, 255, 0.3);
        }
        
        .match-title-preview {
          color: white;
          font-weight: 500;
          text-shadow: 0 1px 2px rgba(0,0,0,0.3);
          overflow: hidden;
          text-overflow: ellipsis;
        }
        
        .match-more {
          font-size: 10px;
          color: #9ca3af;
          text-align: center;
          margin-top: 2px;
        }
      `}</style>
    </div>
  );
}
