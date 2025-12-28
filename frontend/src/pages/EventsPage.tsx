import { useState, useEffect } from 'react';
import Calendar from 'react-calendar';
import 'react-calendar/dist/Calendar.css';
import { api } from '../services/api';
import type { CalendarEvent } from '../types';
import { format, parseISO, isSameDay, isAfter, startOfDay } from 'date-fns';
import { Calendar as CalendarIcon, Clock, MapPin, Plus, Edit2, Trash2, Users, Settings, X, Save } from 'lucide-react';
import EventModal from '../components/events/EventModal';
import { useAuth } from '../hooks/useAuth';
import { usePermissions } from '../hooks/usePermissions';

interface EventCategory {
  id: number;
  name: string;
  color: string;
  sort_order: number;
}

export default function EventsPage() {
  const { user } = useAuth();
  const permissions = usePermissions(user);
  
  const [selectedDate, setSelectedDate] = useState<Date>(new Date());
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [filteredEvents, setFilteredEvents] = useState<CalendarEvent[]>([]);
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);
  const [editingEvent, setEditingEvent] = useState<CalendarEvent | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [categories, setCategories] = useState<EventCategory[]>([]);
  const [selectedColors, setSelectedColors] = useState<Set<string>>(new Set());
  const [showCategoryManager, setShowCategoryManager] = useState(false);
  const [editingCategory, setEditingCategory] = useState<EventCategory | null>(null);

  const currentUserId = user?.discord_id || '';
  const isAdmin = permissions.isAdmin;

  // Load categories on mount
  useEffect(() => {
    loadCategories();
  }, []);

  // Load events for selected month
  useEffect(() => {
    loadEvents();
  }, [selectedDate]);

  // Initialize selected colors when categories load
  useEffect(() => {
    if (categories.length > 0) {
      const colors = new Set(categories.map(c => c.color));
      console.log('Initializing selectedColors:', Array.from(colors));
      setSelectedColors(colors);
    }
  }, [categories]);

  // Filter events for selected date
  useEffect(() => {
    const dateStr = format(selectedDate, 'yyyy-MM-dd');
    console.log('Filtering events for date:', dateStr);
    console.log('Total events:', events.length);
    console.log('Selected colors:', Array.from(selectedColors));
    
    const dayEvents = events.filter(event => {
      // Parse dates and extract only the date part (ignore time and timezone)
      const eventStartDate = format(parseISO(event.start_date), 'yyyy-MM-dd');
      const eventEndDate = format(parseISO(event.end_date), 'yyyy-MM-dd');
      
      // Compare only the date strings
      const inRange = dateStr >= eventStartDate && dateStr <= eventEndDate;
      const colorMatch = selectedColors.has(event.color);
      
      console.log(`Event "${event.title}": inRange=${inRange}, color=${event.color}, colorMatch=${colorMatch}, start=${eventStartDate}, end=${eventEndDate}`);
      
      return inRange && colorMatch;
    });
    
    console.log('Filtered events:', dayEvents.length);
    setFilteredEvents(dayEvents);
    
    // Reset selected event if it's not in the current day
    if (selectedEvent && !dayEvents.find(e => e.id === selectedEvent.id)) {
      setSelectedEvent(null);
    }
  }, [selectedDate, events, selectedColors]);

  const loadEvents = async () => {
    try {
      setLoading(true);
      setError(null);
      const month = selectedDate.getMonth() + 1;
      const year = selectedDate.getFullYear();
      const response = await api.events.getEvents({ month, year });
      console.log('API Response:', response);
      console.log('Response type:', typeof response);
      console.log('Response.events:', response.events);
      
      let eventList = [];
      if (Array.isArray(response)) {
        eventList = response;
      } else if (response && Array.isArray(response.events)) {
        eventList = response.events;
      } else if (response && response.data && Array.isArray(response.data)) {
        eventList = response.data;
      }
      
      console.log('Parsed events:', eventList);
      console.log('Event count:', eventList.length);
      setEvents(eventList);
    } catch (err) {
      console.error('Failed to load events:', err);
      setError('Failed to load events. Please try again.');
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
      const response = await fetch(`${API_BASE_URL}/api/event-categories`, {
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
        { id: 1, name: 'Primary', color: '#4285F4', sort_order: 1 },
        { id: 2, name: 'Success', color: '#0F9D58', sort_order: 2 },
        { id: 3, name: 'Warning', color: '#F4B400', sort_order: 3 },
        { id: 4, name: 'Danger', color: '#DB4437', sort_order: 4 },
        { id: 5, name: 'Purple', color: '#AB47BC', sort_order: 5 },
        { id: 6, name: 'Pink', color: '#E91E63', sort_order: 6 },
        { id: 7, name: 'Orange', color: '#FF6F00', sort_order: 7 },
        { id: 8, name: 'Teal', color: '#009688', sort_order: 8 },
      ]);
    }
  };

  const handleCreateEvent = () => {
    setEditingEvent(null);
    setIsModalOpen(true);
  };

  const handleEditEvent = (event: CalendarEvent) => {
    setEditingEvent(event);
    setIsModalOpen(true);
  };

  const handleDeleteEvent = async (eventId: number) => {
    if (!confirm('Are you sure you want to delete this event?')) return;
    
    try {
      await api.events.deleteEvent(eventId);
      await loadEvents();
      if (selectedEvent?.id === eventId) {
        setSelectedEvent(null);
      }
    } catch (err) {
      console.error('Failed to delete event:', err);
      alert('Failed to delete event. Please try again.');
    }
  };

  const handleSaveEvent = async () => {
    setIsModalOpen(false);
    setEditingEvent(null);
    await loadEvents();
  };

  const canEditEvent = (event: CalendarEvent) => {
    return isAdmin || event.created_by === currentUserId;
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

  // Get events for a specific date
  const getEventsForDate = (date: Date): CalendarEvent[] => {
    const dateStr = format(date, 'yyyy-MM-dd');
    const filtered = events.filter(event => {
      if (!event.start_date || !event.end_date) {
        console.warn('Event missing dates:', event);
        return false;
      }
      
      try {
        // Parse dates and extract only the date part (ignore time and timezone)
        const eventStartDate = format(parseISO(event.start_date), 'yyyy-MM-dd');
        const eventEndDate = format(parseISO(event.end_date), 'yyyy-MM-dd');
        
        // Compare only the date strings
        const inRange = dateStr >= eventStartDate && dateStr <= eventEndDate;
        const colorMatch = selectedColors.has(event.color);
        
        return inRange && colorMatch;
      } catch (err) {
        console.error('Error parsing event dates:', event, err);
        return false;
      }
    });
    
    if (filtered.length > 0 && dateStr === format(new Date(), 'yyyy-MM-dd')) {
      console.log('Events for today:', filtered);
    }
    
    return filtered;
  };

  // Get upcoming events (next 7 days)
  const getUpcomingEvents = (): CalendarEvent[] => {
    const today = startOfDay(new Date());
    return events
      .filter(event => {
        const eventStart = parseISO(event.start_date);
        return isAfter(eventStart, today) || isSameDay(eventStart, today);
      })
      .filter(event => selectedColors.has(event.color))
      .sort((a, b) => parseISO(a.start_date).getTime() - parseISO(b.start_date).getTime())
      .slice(0, 5);
  };

  // Render events in calendar tiles
  const tileContent = ({ date }: { date: Date }) => {
    const dayEvents = getEventsForDate(date);
    if (dayEvents.length === 0) return null;

    return (
      <div className="event-indicators">
        {dayEvents.slice(0, 2).map((event) => (
          <div
            key={event.id}
            className="event-dot"
            style={{ backgroundColor: event.color }}
            title={event.title}
          >
            <img
              src={getAvatarUrl(event.creator_avatar, event.created_by)}
              alt={event.creator_name}
              className="event-creator-avatar"
            />
            <span className="event-title-preview">{event.title}</span>
          </div>
        ))}
        {dayEvents.length > 2 && (
          <div className="event-more">+{dayEvents.length - 2}</div>
        )}
      </div>
    );
  };

  // Highlight today's date
  const tileClassName = ({ date }: { date: Date }) => {
    return isSameDay(date, new Date()) ? 'today-tile' : '';
  };

  // Format event date range
  const formatDateRange = (event: CalendarEvent) => {
    const start = parseISO(event.start_date);
    const end = parseISO(event.end_date);
    
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

  const upcomingEvents = getUpcomingEvents();

  return (
    <div className="p-4 sm:p-6 pt-16 lg:pt-4 sm:pt-6">
      <div className="mb-4 sm:mb-6 flex flex-col sm:flex-row justify-between items-start sm:items-center gap-3">
        <h1 className="text-2xl sm:text-3xl font-bold text-white">Events Calendar</h1>
        <button
          onClick={handleCreateEvent}
          className="flex items-center gap-2 px-3 sm:px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm sm:text-base w-full sm:w-auto justify-center"
        >
          <Plus size={18} className="sm:w-5 sm:h-5" />
          <span>Create Event</span>
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
                <h3 className="text-white font-semibold text-xs sm:text-sm">Event Categories</h3>
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

        {/* Events List Section */}
        <div className="xl:col-span-2">
          <div className="bg-gray-800 rounded-lg p-3 sm:p-6 shadow-lg">
            <div className="flex items-center gap-2 sm:gap-3 mb-3 sm:mb-4">
              <CalendarIcon className="text-blue-400" size={20} />
              <h2 className="text-lg sm:text-xl font-semibold text-white">
                {format(selectedDate, 'MMM d, yyyy')}
              </h2>
            </div>

            {loading ? (
              <div className="text-gray-400 text-center py-8 text-sm sm:text-base">Loading events...</div>
            ) : filteredEvents.length === 0 ? (
              <div className="text-gray-400 text-center py-8 text-sm sm:text-base">
                No events scheduled for this day.
              </div>
            ) : (
              <div className="space-y-3 max-h-[500px] sm:max-h-[600px] overflow-y-auto">
                {filteredEvents.map((event) => (
                  <div
                    key={event.id}
                    onClick={() => setSelectedEvent(event)}
                    className={`p-3 sm:p-4 rounded-lg border-l-4 cursor-pointer transition-all shadow-md ${
                      selectedEvent?.id === event.id
                        ? 'bg-gray-700 border-blue-500 shadow-lg shadow-blue-500/20'
                        : 'bg-gray-700/80 hover:bg-gray-700 border-gray-500 hover:shadow-lg'
                    }`}
                    style={{ borderLeftColor: event.color }}
                  >
                    <div className="flex justify-between items-start mb-2">
                      <h3 className="text-white font-semibold text-sm sm:text-lg">{event.title}</h3>
                      {canEditEvent(event) && (
                        <div className="flex gap-2">
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleEditEvent(event);
                            }}
                            className="p-1 hover:bg-gray-600 rounded transition-colors"
                          >
                            <Edit2 size={16} className="text-blue-400" />
                          </button>
                          <button
                            onClick={(e) => {
                              e.stopPropagation();
                              handleDeleteEvent(event.id);
                            }}
                            className="p-1 hover:bg-gray-600 rounded transition-colors"
                          >
                            <Trash2 size={16} className="text-red-400" />
                          </button>
                        </div>
                      )}
                    </div>
                    
                    <div className="space-y-1 text-sm text-gray-400">
                      {!event.is_all_day && (
                        <div className="flex items-center gap-1">
                          <Clock size={14} />
                          <span>{event.start_time} - {event.end_time}</span>
                        </div>
                      )}
                      {event.is_all_day && (
                        <div className="flex items-center gap-1">
                          <Clock size={14} />
                          <span>Ganztägig</span>
                        </div>
                      )}
                      {event.location && (
                        <div className="flex items-center gap-1">
                          <MapPin size={14} />
                          <span>{event.location}</span>
                        </div>
                      )}
                      {event.guests && (
                        <div className="flex items-center gap-1">
                          <Users size={14} />
                          <span>{event.guests}</span>
                        </div>
                      )}
                    </div>

                    <div className="flex items-center gap-2 text-xs text-gray-500 mt-2">
                      <img
                        src={getAvatarUrl(event.creator_avatar, event.created_by)}
                        alt={event.creator_name}
                        className="w-5 h-5 rounded-full"
                      />
                      <span>{event.creator_name || 'Unknown'}</span>
                    </div>
                  </div>
                ))}
              </div>
            )}

            {/* Event Details Section */}
            {selectedEvent && (
              <div className="mt-6 p-4 bg-gray-750 rounded-lg border border-gray-700">
                <h3 className="text-white font-semibold text-lg mb-3">Event Details</h3>
                <div className="space-y-2 text-gray-300">
                  {selectedEvent.description && (
                    <div>
                      <p className="text-sm text-gray-400">Description</p>
                      <p className="whitespace-pre-wrap">{selectedEvent.description}</p>
                    </div>
                  )}
                  <div>
                    <p className="text-sm text-gray-400">Date</p>
                    <p>{formatDateRange(selectedEvent)}</p>
                  </div>
                  {!selectedEvent.is_all_day && (
                    <div>
                      <p className="text-sm text-gray-400">Time</p>
                      <p>{selectedEvent.start_time} - {selectedEvent.end_time}</p>
                    </div>
                  )}
                  {selectedEvent.is_all_day && (
                    <div>
                      <p className="text-sm text-gray-400">Time</p>
                      <p>Ganztägig</p>
                    </div>
                  )}
                  {selectedEvent.location && (
                    <div>
                      <p className="text-sm text-gray-400">Location</p>
                      <p>{selectedEvent.location}</p>
                    </div>
                  )}
                  {selectedEvent.guests && (
                    <div>
                      <p className="text-sm text-gray-400">Guests</p>
                      <p>{selectedEvent.guests}</p>
                    </div>
                  )}
                  <div>
                    <p className="text-sm text-gray-400">Created by</p>
                    <div className="flex items-center gap-2 mt-1">
                      <img
                        src={getAvatarUrl(selectedEvent.creator_avatar, selectedEvent.created_by)}
                        alt={selectedEvent.creator_name}
                        className="w-6 h-6 rounded-full"
                      />
                      <span>{selectedEvent.creator_name || 'Unknown'}</span>
                    </div>
                  </div>
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* Upcoming Events Section */}
      {upcomingEvents.length > 0 && (
        <div className="mt-6 bg-gray-800 rounded-lg p-6 shadow-lg">
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <CalendarIcon size={24} className="text-blue-400" />
            Upcoming Events
          </h2>
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-4">
            {upcomingEvents.map((event) => (
              <div
                key={event.id}
                onClick={() => {
                  setSelectedDate(parseISO(event.start_date));
                  setSelectedEvent(event);
                }}
                className="p-4 rounded-lg border-l-4 cursor-pointer transition-all shadow-md bg-gray-700/80 hover:bg-gray-700 border-gray-500 hover:shadow-lg"
                style={{ borderLeftColor: event.color }}
              >
                <h3 className="text-white font-semibold mb-2">{event.title}</h3>
                <div className="space-y-1 text-sm text-gray-400">
                  <div className="flex items-center gap-1">
                    <CalendarIcon size={12} />
                    <span>{format(parseISO(event.start_date), 'MMM d')}</span>
                  </div>
                  <div className="flex items-center gap-1">
                    <Clock size={12} />
                    <span>{event.start_time}</span>
                  </div>
                </div>
                <div className="flex items-center gap-2 text-xs text-gray-500 mt-2">
                  <img
                    src={getAvatarUrl(event.creator_avatar, event.created_by)}
                    alt={event.creator_name}
                    className="w-4 h-4 rounded-full"
                  />
                  <span>{event.creator_name}</span>
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
              <h2 className="text-xl font-bold text-white">Manage Event Categories</h2>
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
                            await fetch(`${API_BASE_URL}/api/event-categories/${category.id}`, {
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
                            await fetch(`${API_BASE_URL}/api/event-categories/${category.id}`, {
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
                const color = prompt('Color hex code:', '#4285F4');
                if (!color) return;
                
                try {
                  const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
                  await fetch(`${API_BASE_URL}/api/event-categories`, {
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

      {/* Event Modal */}
      {isModalOpen && (
        <EventModal
          event={editingEvent}
          defaultDate={format(selectedDate, 'yyyy-MM-dd')}
          onClose={() => {
            setIsModalOpen(false);
            setEditingEvent(null);
          }}
          onSave={handleSaveEvent}
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
        
        .event-indicators {
          display: flex;
          flex-direction: column;
          gap: 2px;
          width: 100%;
          margin-top: 4px;
        }
        
        .event-dot {
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
        
        .event-creator-avatar {
          width: 14px;
          height: 14px;
          border-radius: 50%;
          flex-shrink: 0;
          border: 1px solid rgba(255, 255, 255, 0.3);
        }
        
        .event-title-preview {
          color: white;
          font-weight: 500;
          text-shadow: 0 1px 2px rgba(0,0,0,0.3);
          overflow: hidden;
          text-overflow: ellipsis;
        }
        
        .event-more {
          font-size: 10px;
          color: #9ca3af;
          text-align: center;
          margin-top: 2px;
        }
      `}</style>
    </div>
  );
}
