import { useState, useEffect } from 'react';
import { Calendar as CalendarIcon, Clock, MapPin, Users } from 'lucide-react';
import { api } from '../../services/api';
import type { CalendarEvent } from '../../types';
import { format, parseISO, isAfter, startOfDay, isSameDay } from 'date-fns';
import EventModal from '../events/EventModal';

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

interface EventCategory {
  id: number;
  name: string;
  color: string;
  sort_order: number;
}

export default function UpcomingEvents() {
  const [events, setEvents] = useState<CalendarEvent[]>([]);
  const [categories, setCategories] = useState<EventCategory[]>([]);
  const [loading, setLoading] = useState(true);
  const [selectedEvent, setSelectedEvent] = useState<CalendarEvent | null>(null);
  const [isModalOpen, setIsModalOpen] = useState(false);

  useEffect(() => {
    loadEvents();
    loadCategories();
  }, []);

  const loadEvents = async () => {
    try {
      setLoading(true);
      const today = new Date();
      const month = today.getMonth() + 1;
      const year = today.getFullYear();
      const response = await api.events.getEvents({ month, year });
      
      let eventList = [];
      if (Array.isArray(response)) {
        eventList = response;
      } else if (response && Array.isArray(response.events)) {
        eventList = response.events;
      } else if (response && response.data && Array.isArray(response.data)) {
        eventList = response.data;
      }

      // Filter to upcoming events only (next 7 days)
      const todayStart = startOfDay(new Date());
      const upcomingEvents = eventList
        .filter((event: CalendarEvent) => {
          const eventStart = parseISO(event.start_date);
          return isAfter(eventStart, todayStart) || isSameDay(eventStart, todayStart);
        })
        .sort((a: CalendarEvent, b: CalendarEvent) => parseISO(a.start_date).getTime() - parseISO(b.start_date).getTime())
        .slice(0, 5);
      
      setEvents(upcomingEvents);
    } catch (err) {
      console.error('Failed to load events:', err);
    } finally {
      setLoading(false);
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
      
      if (response.ok) {
        const data = await response.json();
        setCategories(data.categories || data || []);
      }
    } catch (err) {
      console.error('Failed to load categories:', err);
    }
  };

  const handleEventClick = (event: CalendarEvent) => {
    setSelectedEvent(event);
    setIsModalOpen(true);
  };

  const handleModalClose = () => {
    setIsModalOpen(false);
    setSelectedEvent(null);
  };

  const handleModalSave = () => {
    setIsModalOpen(false);
    setSelectedEvent(null);
    loadEvents();
  };

  if (loading) {
    return (
      <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
        <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4 flex items-center gap-2">
          <CalendarIcon size={20} className="text-blue-400" />
          Upcoming Events
        </h2>
        <div className="text-gray-400 text-center py-8">Loading events...</div>
      </div>
    );
  }

  return (
    <>
      <div className="bg-[#1a1f2e] rounded-lg p-4 sm:p-6 border border-gray-800">
        <h2 className="text-white text-base sm:text-lg font-semibold mb-3 sm:mb-4 flex items-center gap-2">
          <CalendarIcon size={20} className="text-blue-400" />
          Upcoming Events
        </h2>
        
        {events.length === 0 ? (
          <div className="text-gray-400 text-center py-8 text-sm">
            No upcoming events scheduled.
          </div>
        ) : (
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-5 gap-3">
            {events.map((event) => (
              <div
                key={event.id}
                onClick={() => handleEventClick(event)}
                className="p-3 sm:p-4 rounded-lg border-l-4 cursor-pointer transition-all shadow-md bg-gray-700/80 hover:bg-gray-700 hover:shadow-lg"
                style={{ borderLeftColor: event.color }}
              >
                <h3 className="text-white font-semibold text-sm mb-2 line-clamp-2">{event.title}</h3>
                <div className="space-y-1 text-xs text-gray-400">
                  <div className="flex items-center gap-1">
                    <CalendarIcon size={12} />
                    <span>{format(parseISO(event.start_date), 'MMM d')}</span>
                  </div>
                  {!event.is_all_day && (
                    <div className="flex items-center gap-1">
                      <Clock size={12} />
                      <span>{formatTime(event.start_time)}</span>
                    </div>
                  )}
                  {event.is_all_day && (
                    <div className="flex items-center gap-1">
                      <Clock size={12} />
                      <span>All Day</span>
                    </div>
                  )}
                  {event.location && (
                    <div className="flex items-center gap-1">
                      <MapPin size={12} />
                      <span className="truncate">{event.location}</span>
                    </div>
                  )}
                </div>
                <div className="flex items-center gap-2 text-xs text-gray-500 mt-2 pt-2 border-t border-gray-600">
                  <img
                    src={getAvatarUrl(event.creator_avatar, event.created_by)}
                    alt={event.creator_name}
                    className="w-4 h-4 rounded-full"
                  />
                  <span className="truncate">{event.creator_name}</span>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>

      {/* Event Modal */}
      {isModalOpen && (
        <EventModal
          event={selectedEvent}
          onClose={handleModalClose}
          onSave={handleModalSave}
          categories={categories}
        />
      )}
    </>
  );
}
