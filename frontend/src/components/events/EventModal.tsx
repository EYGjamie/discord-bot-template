import { useState } from 'react';
import { X } from 'lucide-react';
import { api } from '../../services/api';
import type { CalendarEvent } from '../../types';

interface EventModalProps {
  event?: CalendarEvent | null;
  defaultDate?: string;
  onClose: () => void;
  onSave: () => void;
  categories?: Array<{ id: number; name: string; color: string; sort_order: number }>;
}

export default function EventModal({ event, defaultDate, onClose, onSave, categories = [] }: EventModalProps) {
  const [formData, setFormData] = useState({
    title: event?.title || '',
    description: event?.description || '',
    start_date: event?.start_date || defaultDate || '',
    end_date: event?.end_date || defaultDate || '',
    start_time: event?.start_time || '09:00',
    end_time: event?.end_time || '10:00',
    color: event?.color || '#4285F4',
    location: event?.location || '',
    guests: event?.guests || '',
  });
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.title.trim()) {
      setError('Title is required');
      return;
    }
    
    if (!formData.start_date) {
      setError('Start date is required');
      return;
    }

    // If end_date not set, use start_date
    const submitData = {
      ...formData,
      end_date: formData.end_date || formData.start_date,
    };
    
    setLoading(true);
    setError(null);
    
    try {
      if (event) {
        await api.events.updateEvent(event.id, submitData);
      } else {
        await api.events.createEvent(submitData);
      }
      onSave();
    } catch (err) {
      console.error('Failed to save event:', err);
      setError('Failed to save event. Please try again.');
    } finally {
      setLoading(false);
    }
  };

  const handleChange = (e: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement>) => {
    setFormData(prev => ({
      ...prev,
      [e.target.name]: e.target.value,
    }));
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-800 rounded-lg shadow-xl w-full max-w-2xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex justify-between items-center p-6 border-b border-gray-700">
          <h2 className="text-2xl font-bold text-white">
            {event ? 'Edit Event' : 'Create Event'}
          </h2>
          <button
            onClick={onClose}
            className="text-gray-400 hover:text-white transition-colors"
          >
            <X size={24} />
          </button>
        </div>

        {/* Form */}
        <form onSubmit={handleSubmit} className="p-6 space-y-4">
          {error && (
            <div className="p-3 bg-red-500/10 border border-red-500/20 rounded text-red-400 text-sm">
              {error}
            </div>
          )}

          {/* Title */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Title <span className="text-red-400">*</span>
            </label>
            <input
              type="text"
              name="title"
              value={formData.title}
              onChange={handleChange}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Event title"
              required
            />
          </div>

          {/* Description */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Description
            </label>
            <textarea
              name="description"
              value={formData.description}
              onChange={handleChange}
              rows={4}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
              placeholder="Event description (optional)"
            />
          </div>

          {/* Date Range */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Start Date <span className="text-red-400">*</span>
              </label>
              <input
                type="date"
                name="start_date"
                value={formData.start_date}
                onChange={handleChange}
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                required
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                End Date
              </label>
              <input
                type="date"
                name="end_date"
                value={formData.end_date}
                onChange={handleChange}
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="Same as start date if empty"
              />
            </div>
          </div>

          {/* Time Range */}
          <div className="grid grid-cols-2 gap-4">
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Start Time
              </label>
              <input
                type="time"
                name="start_time"
                value={formData.start_time}
                onChange={handleChange}
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                End Time
              </label>
              <input
                type="time"
                name="end_time"
                value={formData.end_time}
                onChange={handleChange}
                className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>
          </div>

          {/* Location */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Location
            </label>
            <input
              type="text"
              name="location"
              value={formData.location}
              onChange={handleChange}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Event location (optional)"
            />
          </div>

          {/* Guests */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Guests
            </label>
            <input
              type="text"
              name="guests"
              value={formData.guests}
              onChange={handleChange}
              className="w-full px-4 py-2 bg-gray-700 border border-gray-600 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
              placeholder="Guest names (comma-separated)"
            />
            <p className="text-xs text-gray-500 mt-1">Example: John Doe, Jane Smith, Bob Johnson</p>
          </div>

          {/* Color Picker */}
          <div>
            <label className="block text-sm font-medium text-gray-300 mb-2">
              Category
            </label>
            <div className="grid grid-cols-4 gap-3">
              {(categories.length > 0 ? categories : [
                { id: 1, name: 'Blue', color: '#4285F4', sort_order: 1 },
                { id: 2, name: 'Red', color: '#DB4437', sort_order: 2 },
                { id: 3, name: 'Yellow', color: '#F4B400', sort_order: 3 },
                { id: 4, name: 'Green', color: '#0F9D58', sort_order: 4 },
                { id: 5, name: 'Purple', color: '#AB47BC', sort_order: 5 },
                { id: 6, name: 'Orange', color: '#FF6D00', sort_order: 6 },
                { id: 7, name: 'Cyan', color: '#00ACC1', sort_order: 7 },
                { id: 8, name: 'Gray', color: '#616161', sort_order: 8 },
              ]).map((colorOption) => (
                <button
                  key={colorOption.color}
                  type="button"
                  onClick={() => setFormData(prev => ({ ...prev, color: colorOption.color }))}
                  className={`flex flex-col items-center gap-2 p-3 rounded-lg border-2 transition-all ${
                    formData.color === colorOption.color
                      ? 'border-white bg-gray-700'
                      : 'border-gray-600 hover:border-gray-500 bg-gray-750'
                  }`}
                >
                  <div
                    className="w-12 h-12 rounded-lg"
                    style={{ backgroundColor: colorOption.color }}
                  />
                  <span className="text-xs text-gray-300 font-medium">{colorOption.name}</span>
                </button>
              ))}
            </div>
          </div>

          {/* Footer */}
          <div className="flex justify-end gap-3 pt-4 border-t border-gray-700">
            <button
              type="button"
              onClick={onClose}
              className="px-4 py-2 text-gray-300 hover:text-white transition-colors"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={loading}
            >
              {loading ? 'Saving...' : event ? 'Save Changes' : 'Create Event'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
