import { useState, useEffect } from 'react';
import { X, Plus, MapPin, Clock, Calendar, Tag, Users, CheckSquare, Loader2, Search, Check, XIcon, HelpCircle } from 'lucide-react';
import { api } from '../../services/api';
import type { CalendarEvent, EventGuest, EventLabel, EventChecklistItem, Member } from '../../types';

interface EventModalProps {
  event?: CalendarEvent | null;
  defaultDate?: string;
  onClose: () => void;
  onSave: () => void;
  categories?: Array<{ id: number; name: string; color: string; sort_order: number }>;
}

const LABEL_COLORS = [
  { value: 'green', bg: 'bg-green-600', text: 'text-white', name: 'Grün' },
  { value: 'yellow', bg: 'bg-yellow-500', text: 'text-gray-900', name: 'Gelb' },
  { value: 'orange', bg: 'bg-orange-600', text: 'text-white', name: 'Orange' },
  { value: 'red', bg: 'bg-red-600', text: 'text-white', name: 'Rot' },
  { value: 'purple', bg: 'bg-purple-600', text: 'text-white', name: 'Lila' },
  { value: 'blue', bg: 'bg-blue-600', text: 'text-white', name: 'Blau' },
  { value: 'pink', bg: 'bg-pink-600', text: 'text-white', name: 'Pink' },
  { value: 'teal', bg: 'bg-teal-600', text: 'text-white', name: 'Türkis' },
];

// Helper function to format date for input field (yyyy-MM-dd)
const formatDateForInput = (dateStr: string | undefined): string => {
  if (!dateStr) return '';
  // If already in yyyy-MM-dd format, return as is
  if (/^\d{4}-\d{2}-\d{2}$/.test(dateStr)) return dateStr;
  // Otherwise parse and format (handles ISO strings like "2026-01-31T00:00:00Z")
  try {
    const date = new Date(dateStr);
    return date.toISOString().split('T')[0];
  } catch {
    return '';
  }
};

export default function EventModal({ event, defaultDate, onClose, onSave, categories = [] }: EventModalProps) {
  const [formData, setFormData] = useState({
    title: event?.title || '',
    description: event?.description || '',
    start_date: formatDateForInput(event?.start_date) || defaultDate || '',
    end_date: formatDateForInput(event?.end_date) || defaultDate || '',
    start_time: event?.start_time || '09:00',
    end_time: event?.end_time || '10:00',
    is_all_day: event?.is_all_day || false,
    color: event?.color || '#4285F4',
    location: event?.location || '',
    tags: event?.tags || [],
  });
  
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  
  // Guests
  const [guests, setGuests] = useState<EventGuest[]>([]);
  const [guestSearchQuery, setGuestSearchQuery] = useState('');
  const [guestSearchResults, setGuestSearchResults] = useState<Member[]>([]);
  const [searchingGuests, setSearchingGuests] = useState(false);
  const [showGuestSearch, setShowGuestSearch] = useState(false);
  
  // Labels
  const [availableLabels, setAvailableLabels] = useState<EventLabel[]>([]);
  const [showLabelPicker, setShowLabelPicker] = useState(false);
  const [newLabelName, setNewLabelName] = useState('');
  const [newLabelColor, setNewLabelColor] = useState('blue');
  
  // Checklist
  const [checklist, setChecklist] = useState<EventChecklistItem[]>([]);
  const [newChecklistItem, setNewChecklistItem] = useState('');

  // Load data when editing
  useEffect(() => {
    loadLabels();
    if (event?.id) {
      loadGuests();
      loadChecklist();
    }
  }, [event?.id]);

  const loadLabels = async () => {
    try {
      const labels = await api.eventLabels.getAll();
      setAvailableLabels(labels || []);
    } catch (err) {
      console.error('Failed to load labels:', err);
    }
  };

  const loadGuests = async () => {
    if (!event?.id) return;
    try {
      const data = await api.events.getGuests(event.id);
      setGuests(data || []);
    } catch (err) {
      console.error('Failed to load guests:', err);
    }
  };

  const loadChecklist = async () => {
    if (!event?.id) return;
    try {
      const data = await api.events.getChecklist(event.id);
      setChecklist(data || []);
    } catch (err) {
      console.error('Failed to load checklist:', err);
    }
  };

  // Guest search
  const searchUsers = async (query: string) => {
    if (query.length < 2) {
      setGuestSearchResults([]);
      return;
    }
    setSearchingGuests(true);
    try {
      const response = await api.members.getMembers({ search: query, per_page: 10 });
      const members = response.members || [];
      // Filter out already invited guests
      const guestIds = guests.map(g => g.user_id);
      setGuestSearchResults(members.filter((m: Member) => !guestIds.includes(m.id)));
    } catch (err) {
      console.error('Failed to search users:', err);
    } finally {
      setSearchingGuests(false);
    }
  };

  useEffect(() => {
    const timeout = setTimeout(() => {
      if (guestSearchQuery) {
        searchUsers(guestSearchQuery);
      }
    }, 300);
    return () => clearTimeout(timeout);
  }, [guestSearchQuery]);

  const inviteGuest = async (member: Member) => {
    if (!event?.id) return;
    try {
      const newGuest = await api.events.inviteGuest(event.id, member.id);
      setGuests([...guests, newGuest]);
      setGuestSearchQuery('');
      setGuestSearchResults([]);
      setShowGuestSearch(false);
    } catch (err) {
      console.error('Failed to invite guest:', err);
    }
  };

  const removeGuest = async (guestId: number) => {
    if (!event?.id) return;
    try {
      await api.events.removeGuest(event.id, guestId);
      setGuests(guests.filter(g => g.id !== guestId));
    } catch (err) {
      console.error('Failed to remove guest:', err);
    }
  };

  // Labels
  const toggleLabel = (labelName: string) => {
    if (formData.tags.includes(labelName)) {
      setFormData({ ...formData, tags: formData.tags.filter(t => t !== labelName) });
    } else {
      setFormData({ ...formData, tags: [...formData.tags, labelName] });
    }
  };

  const createLabel = async () => {
    if (!newLabelName.trim()) return;
    try {
      const label = await api.eventLabels.create(newLabelName.trim(), newLabelColor);
      setAvailableLabels([...availableLabels, label]);
      setFormData({ ...formData, tags: [...formData.tags, label.name] });
      setNewLabelName('');
      setNewLabelColor('blue');
    } catch (err) {
      console.error('Failed to create label:', err);
    }
  };

  // Checklist
  const addChecklistItem = async () => {
    if (!newChecklistItem.trim() || !event?.id) return;
    try {
      const item = await api.events.createChecklistItem(event.id, newChecklistItem.trim());
      setChecklist([...checklist, item]);
      setNewChecklistItem('');
    } catch (err) {
      console.error('Failed to add checklist item:', err);
    }
  };

  const toggleChecklistItem = async (item: EventChecklistItem) => {
    if (!event?.id) return;
    try {
      const updated = await api.events.updateChecklistItem(event.id, item.id, { is_completed: !item.is_completed });
      setChecklist(checklist.map(c => c.id === item.id ? updated : c));
    } catch (err) {
      console.error('Failed to toggle checklist item:', err);
    }
  };

  const deleteChecklistItem = async (itemId: number) => {
    if (!event?.id) return;
    try {
      await api.events.deleteChecklistItem(event.id, itemId);
      setChecklist(checklist.filter(c => c.id !== itemId));
    } catch (err) {
      console.error('Failed to delete checklist item:', err);
    }
  };

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.title.trim()) {
      setError('Titel ist erforderlich');
      return;
    }
    
    if (!formData.start_date) {
      setError('Startdatum ist erforderlich');
      return;
    }

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
      setError('Fehler beim Speichern. Bitte versuche es erneut.');
    } finally {
      setLoading(false);
    }
  };

  const getLabelColor = (labelName: string) => {
    const label = availableLabels.find(l => l.name === labelName);
    const colorValue = label?.color || 'blue';
    return LABEL_COLORS.find(c => c.value === colorValue) || LABEL_COLORS[5];
  };

  const getRSVPIcon = (status: string) => {
    switch (status) {
      case 'accepted': return <Check size={14} className="text-green-400" />;
      case 'declined': return <XIcon size={14} className="text-red-400" />;
      default: return <HelpCircle size={14} className="text-yellow-400" />;
    }
  };

  const getRSVPText = (status: string) => {
    switch (status) {
      case 'accepted': return 'Zugesagt';
      case 'declined': return 'Abgesagt';
      default: return 'Ausstehend';
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-[#1e2228] rounded-lg w-full max-w-4xl max-h-[90vh] overflow-hidden border border-white/10 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <Calendar size={20} className="text-blue-400" />
            <h2 className="text-lg font-semibold text-white">
              {event ? 'Event bearbeiten' : 'Neues Event'}
            </h2>
          </div>
          <button onClick={onClose} className="p-1 hover:bg-white/10 rounded transition-colors">
            <X size={20} className="text-gray-400" />
          </button>
        </div>

        {/* Content */}
        <div className="flex flex-col md:flex-row max-h-[calc(90vh-140px)] overflow-hidden">
          {/* Main Form */}
          <form onSubmit={handleSubmit} className="flex-1 p-4 space-y-4 overflow-y-auto">
            {error && (
              <div className="p-3 bg-red-500/10 border border-red-500/20 rounded text-red-400 text-sm">
                {error}
              </div>
            )}

            {/* Title */}
            <div>
              <input
                type="text"
                value={formData.title}
                onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                className="w-full px-3 py-2 bg-transparent border-b border-white/20 text-white text-lg font-medium focus:outline-none focus:border-blue-500 placeholder-gray-500"
                placeholder="Event-Titel..."
                required
              />
            </div>

            {/* Description */}
            <div>
              <textarea
                value={formData.description}
                onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                rows={3}
                className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none text-sm"
                placeholder="Beschreibung (optional)"
              />
            </div>

            {/* Date & Time Row */}
            <div className="grid grid-cols-2 gap-4">
              <div>
                <label className="text-xs font-medium text-gray-400 mb-1 block flex items-center gap-1">
                  <Calendar size={12} /> Startdatum
                </label>
                <input
                  type="date"
                  value={formData.start_date}
                  onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                  className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  required
                />
              </div>
              <div>
                <label className="text-xs font-medium text-gray-400 mb-1 block">Enddatum</label>
                <input
                  type="date"
                  value={formData.end_date}
                  onChange={(e) => setFormData({ ...formData, end_date: e.target.value })}
                  className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                />
              </div>
            </div>

            {/* All Day Toggle */}
            <div className="flex items-center gap-2">
              <input
                type="checkbox"
                id="is_all_day"
                checked={formData.is_all_day}
                onChange={(e) => setFormData({ ...formData, is_all_day: e.target.checked })}
                className="w-4 h-4 rounded border-white/20 bg-[#0d0f15] text-blue-600"
              />
              <label htmlFor="is_all_day" className="text-sm text-gray-300 cursor-pointer">
                Ganztägig
              </label>
            </div>

            {/* Time Range */}
            {!formData.is_all_day && (
              <div className="grid grid-cols-2 gap-4">
                <div>
                  <label className="text-xs font-medium text-gray-400 mb-1 block flex items-center gap-1">
                    <Clock size={12} /> Startzeit
                  </label>
                  <input
                    type="time"
                    value={formData.start_time}
                    onChange={(e) => setFormData({ ...formData, start_time: e.target.value })}
                    className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                </div>
                <div>
                  <label className="text-xs font-medium text-gray-400 mb-1 block">Endzeit</label>
                  <input
                    type="time"
                    value={formData.end_time}
                    onChange={(e) => setFormData({ ...formData, end_time: e.target.value })}
                    className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                  />
                </div>
              </div>
            )}

            {/* Location */}
            <div>
              <label className="text-xs font-medium text-gray-400 mb-1 block flex items-center gap-1">
                <MapPin size={12} /> Ort
              </label>
              <input
                type="text"
                value={formData.location}
                onChange={(e) => setFormData({ ...formData, location: e.target.value })}
                className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                placeholder="Ort des Events (optional)"
              />
            </div>

            {/* Labels */}
            <div>
              <label className="text-xs font-medium text-gray-400 mb-2 block flex items-center gap-1">
                <Tag size={12} /> Labels
              </label>
              <div className="flex flex-wrap gap-2 mb-2">
                {formData.tags.map((tag) => {
                  const color = getLabelColor(tag);
                  return (
                    <span
                      key={tag}
                      className={`px-2 py-1 rounded text-xs font-medium ${color.bg} ${color.text} cursor-pointer hover:opacity-80`}
                      onClick={() => toggleLabel(tag)}
                    >
                      {tag} ×
                    </span>
                  );
                })}
                <button
                  type="button"
                  onClick={() => setShowLabelPicker(!showLabelPicker)}
                  className="px-2 py-1 rounded text-xs font-medium bg-white/10 text-gray-300 hover:bg-white/20 flex items-center gap-1"
                >
                  <Plus size={12} /> Label
                </button>
              </div>
              
              {showLabelPicker && (
                <div className="p-3 bg-[#0d0f15] border border-white/10 rounded-lg space-y-3">
                  <div className="flex flex-wrap gap-2">
                    {availableLabels.map((label) => {
                      const color = LABEL_COLORS.find(c => c.value === label.color) || LABEL_COLORS[5];
                      const isSelected = formData.tags.includes(label.name);
                      return (
                        <button
                          key={label.id}
                          type="button"
                          onClick={() => toggleLabel(label.name)}
                          className={`px-2 py-1 rounded text-xs font-medium ${color.bg} ${color.text} ${isSelected ? 'ring-2 ring-white' : 'opacity-70 hover:opacity-100'}`}
                        >
                          {label.name}
                        </button>
                      );
                    })}
                  </div>
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={newLabelName}
                      onChange={(e) => setNewLabelName(e.target.value)}
                      placeholder="Neues Label..."
                      className="flex-1 px-2 py-1 bg-[#1e2228] border border-white/10 rounded text-sm text-white"
                    />
                    <select
                      value={newLabelColor}
                      onChange={(e) => setNewLabelColor(e.target.value)}
                      className="px-2 py-1 bg-[#1e2228] border border-white/10 rounded text-sm text-white"
                    >
                      {LABEL_COLORS.map(c => (
                        <option key={c.value} value={c.value}>{c.name}</option>
                      ))}
                    </select>
                    <button
                      type="button"
                      onClick={createLabel}
                      disabled={!newLabelName.trim()}
                      className="px-3 py-1 bg-blue-600 hover:bg-blue-700 text-white rounded text-sm disabled:opacity-50"
                    >
                      <Plus size={14} />
                    </button>
                  </div>
                </div>
              )}
            </div>

            {/* Category Color */}
            <div>
              <label className="text-xs font-medium text-gray-400 mb-2 block">Kategorie-Farbe</label>
              <div className="grid grid-cols-8 gap-2">
                {(categories.length > 0 ? categories : [
                  { id: 1, name: 'Blue', color: '#4285F4', sort_order: 1 },
                  { id: 2, name: 'Red', color: '#DB4437', sort_order: 2 },
                  { id: 3, name: 'Yellow', color: '#F4B400', sort_order: 3 },
                  { id: 4, name: 'Green', color: '#0F9D58', sort_order: 4 },
                  { id: 5, name: 'Purple', color: '#AB47BC', sort_order: 5 },
                  { id: 6, name: 'Orange', color: '#FF6D00', sort_order: 6 },
                  { id: 7, name: 'Cyan', color: '#00ACC1', sort_order: 7 },
                  { id: 8, name: 'Gray', color: '#616161', sort_order: 8 },
                ]).map((c) => (
                  <button
                    key={c.color}
                    type="button"
                    onClick={() => setFormData({ ...formData, color: c.color })}
                    className={`w-8 h-8 rounded-full border-2 ${formData.color === c.color ? 'border-white' : 'border-transparent'}`}
                    style={{ backgroundColor: c.color }}
                    title={c.name}
                  />
                ))}
              </div>
            </div>

            {/* Checklist (only for existing events) */}
            {event?.id && (
              <div>
                <label className="text-xs font-medium text-gray-400 mb-2 block flex items-center gap-1">
                  <CheckSquare size={12} /> Checkliste
                </label>
                <div className="space-y-2">
                  {checklist.map((item) => (
                    <div key={item.id} className="flex items-center gap-2 p-2 bg-[#0d0f15] border border-white/10 rounded">
                      <input
                        type="checkbox"
                        checked={item.is_completed}
                        onChange={() => toggleChecklistItem(item)}
                        className="w-4 h-4 rounded border-white/20 bg-[#0d0f15] text-blue-600"
                      />
                      <span className={`flex-1 text-sm ${item.is_completed ? 'line-through text-gray-500' : 'text-gray-300'}`}>
                        {item.text}
                      </span>
                      <button
                        type="button"
                        onClick={() => deleteChecklistItem(item.id)}
                        className="text-gray-500 hover:text-red-400"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ))}
                  <div className="flex gap-2">
                    <input
                      type="text"
                      value={newChecklistItem}
                      onChange={(e) => setNewChecklistItem(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addChecklistItem())}
                      placeholder="Neuer Eintrag..."
                      className="flex-1 px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white text-sm"
                    />
                    <button
                      type="button"
                      onClick={addChecklistItem}
                      disabled={!newChecklistItem.trim()}
                      className="px-3 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg text-sm disabled:opacity-50"
                    >
                      <Plus size={14} />
                    </button>
                  </div>
                </div>
              </div>
            )}
          </form>

          {/* Sidebar - Guests (only for existing events) */}
          {event?.id && (
            <div className="w-full md:w-80 border-t md:border-t-0 md:border-l border-white/10 p-4 overflow-y-auto bg-[#151820]">
              <div className="flex items-center justify-between mb-3">
                <h3 className="text-sm font-medium text-gray-300 flex items-center gap-2">
                  <Users size={14} /> Gäste ({guests.length})
                </h3>
                <button
                  type="button"
                  onClick={() => setShowGuestSearch(!showGuestSearch)}
                  className="p-1 hover:bg-white/10 rounded text-blue-400"
                >
                  <Plus size={16} />
                </button>
              </div>

              {/* Guest Search */}
              {showGuestSearch && (
                <div className="mb-4 relative">
                  <div className="relative">
                    <Search size={14} className="absolute left-3 top-1/2 -translate-y-1/2 text-gray-500" />
                    <input
                      type="text"
                      value={guestSearchQuery}
                      onChange={(e) => setGuestSearchQuery(e.target.value)}
                      placeholder="User suchen..."
                      className="w-full pl-9 pr-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-white text-sm"
                    />
                  </div>
                  {(searchingGuests || guestSearchResults.length > 0) && (
                    <div className="absolute top-full left-0 right-0 mt-1 bg-[#1e2228] border border-white/10 rounded-lg shadow-xl z-10 max-h-48 overflow-y-auto">
                      {searchingGuests ? (
                        <div className="p-3 text-center text-gray-500">
                          <Loader2 size={16} className="animate-spin mx-auto" />
                        </div>
                      ) : (
                        guestSearchResults.map((member) => (
                          <button
                            key={member.id}
                            type="button"
                            onClick={() => inviteGuest(member)}
                            className="w-full flex items-center gap-2 p-2 hover:bg-white/5 text-left"
                          >
                            {member.avatar_url ? (
                              <img src={member.avatar_url} alt="" className="w-8 h-8 rounded-full" />
                            ) : (
                              <div className="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center text-white text-sm">
                                {member.display_name?.[0] || member.name?.[0] || '?'}
                              </div>
                            )}
                            <div>
                              <div className="text-sm text-white">{member.display_name || member.name}</div>
                              <div className="text-xs text-gray-500">@{member.name}</div>
                            </div>
                          </button>
                        ))
                      )}
                    </div>
                  )}
                </div>
              )}

              {/* Guest List with RSVP Status */}
              <div className="space-y-2">
                {guests.length === 0 ? (
                  <p className="text-xs text-gray-500 text-center py-4">
                    Noch keine Gäste eingeladen
                  </p>
                ) : (
                  guests.map((guest) => (
                    <div
                      key={guest.id}
                      className="flex items-center gap-2 p-2 bg-[#0d0f15] border border-white/10 rounded-lg"
                    >
                      {guest.user_avatar ? (
                        <img
                          src={`https://cdn.discordapp.com/avatars/${guest.user_id}/${guest.user_avatar}.png?size=32`}
                          alt=""
                          className="w-8 h-8 rounded-full"
                        />
                      ) : (
                        <div className="w-8 h-8 rounded-full bg-gray-600 flex items-center justify-center text-white text-xs">
                          {guest.user_display_name?.[0] || '?'}
                        </div>
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="text-sm text-white truncate">{guest.user_display_name}</div>
                        <div className="flex items-center gap-1 text-xs">
                          {getRSVPIcon(guest.rsvp_status)}
                          <span className={`${
                            guest.rsvp_status === 'accepted' ? 'text-green-400' :
                            guest.rsvp_status === 'declined' ? 'text-red-400' :
                            'text-yellow-400'
                          }`}>
                            {getRSVPText(guest.rsvp_status)}
                          </span>
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={() => removeGuest(guest.id)}
                        className="text-gray-500 hover:text-red-400 p-1"
                      >
                        <X size={14} />
                      </button>
                    </div>
                  ))
                )}
              </div>

              {/* RSVP Summary */}
              {guests.length > 0 && (
                <div className="mt-4 pt-4 border-t border-white/10">
                  <h4 className="text-xs font-medium text-gray-400 mb-2">Zusammenfassung</h4>
                  <div className="grid grid-cols-3 gap-2 text-center">
                    <div className="p-2 bg-green-500/10 rounded">
                      <div className="text-lg font-bold text-green-400">
                        {guests.filter(g => g.rsvp_status === 'accepted').length}
                      </div>
                      <div className="text-xs text-gray-500">Zugesagt</div>
                    </div>
                    <div className="p-2 bg-red-500/10 rounded">
                      <div className="text-lg font-bold text-red-400">
                        {guests.filter(g => g.rsvp_status === 'declined').length}
                      </div>
                      <div className="text-xs text-gray-500">Abgesagt</div>
                    </div>
                    <div className="p-2 bg-yellow-500/10 rounded">
                      <div className="text-lg font-bold text-yellow-400">
                        {guests.filter(g => g.rsvp_status === 'pending').length}
                      </div>
                      <div className="text-xs text-gray-500">Offen</div>
                    </div>
                  </div>
                </div>
              )}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex gap-3 p-4 border-t border-white/10">
          <button
            type="button"
            onClick={onClose}
            className="flex-1 px-4 py-2 bg-white/10 hover:bg-white/20 text-white rounded-lg transition-colors text-sm font-medium"
            disabled={loading}
          >
            Abbrechen
          </button>
          <button
            onClick={handleSubmit}
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm font-medium disabled:opacity-50"
            disabled={loading}
          >
            {loading ? <Loader2 size={16} className="animate-spin mx-auto" /> : (event ? 'Speichern' : 'Erstellen')}
          </button>
        </div>
      </div>
    </div>
  );
}
