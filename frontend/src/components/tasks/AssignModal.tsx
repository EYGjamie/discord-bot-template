import React, { useState, useEffect } from 'react';
import { X, Search, User } from 'lucide-react';
import { api } from '../../services/api';

interface Member {
  user_id: string;
  display_name: string;
  username: string;
  avatar: string | null;
}

interface AssignModalProps {
  currentAssignees: string[];
  onClose: () => void;
  onAssign: (assignees: string[]) => void;
}

const AssignModal: React.FC<AssignModalProps> = ({ currentAssignees, onClose, onAssign }) => {
  const [members, setMembers] = useState<Member[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedAssignees, setSelectedAssignees] = useState<string[]>(currentAssignees);
  const [loading, setLoading] = useState(false);
  const [hasSearched, setHasSearched] = useState(false);

  useEffect(() => {
    // Debounce search - wait 300ms after user stops typing
    if (searchTerm.trim().length < 2) {
      setMembers([]);
      setHasSearched(false);
      return;
    }

    setLoading(true);
    const timeoutId = setTimeout(() => {
      searchMembers(searchTerm);
    }, 300);

    return () => clearTimeout(timeoutId);
  }, [searchTerm]);

  const searchMembers = async (query: string) => {
    try {
      const data = await api.get(`/api/discord/members/search?q=${encodeURIComponent(query)}`);
      console.log('Search results:', data);
      setMembers(Array.isArray(data) ? data : []);
      setHasSearched(true);
    } catch (error) {
      console.error('Failed to search members:', error);
      setMembers([]);
      setHasSearched(true);
    } finally {
      setLoading(false);
    }
  };

  const toggleAssignee = (userId: string) => {
    // Single-select mode: Only one assignee at a time
    if (selectedAssignees.includes(userId)) {
      // If clicking the same user, deselect them
      setSelectedAssignees([]);
    } else {
      // Select only this user
      setSelectedAssignees([userId]);
    }
  };

  const handleAssign = () => {
    onAssign(selectedAssignees);
    onClose();
  };

  const handleRemoveAll = () => {
    onAssign([]);
    onClose();
  };

  const getAvatarUrl = (avatar: string | null, userId: string) => {
    if (!avatar) return null;
    // Discord CDN avatar URL
    return `https://cdn.discordapp.com/avatars/${userId}/${avatar}.png?size=64`;
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] p-4">
      <div className="bg-[#1a1d29] border border-white/10 rounded-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-white/10">
          <h2 className="text-xl font-bold text-white">Benutzer zuweisen</h2>
          <button
            onClick={onClose}
            className="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <X size={20} className="text-white" />
          </button>
        </div>

        {/* Search */}
        <div className="p-4 border-b border-white/10">
          <div className="relative">
            <Search className="absolute left-3 top-1/2 transform -translate-y-1/2 text-gray-400" size={18} />
            <input
              type="text"
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              placeholder="Mindestens 2 Zeichen eingeben..."
              className="w-full pl-10 pr-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
              autoFocus
            />
          </div>
        </div>

        {/* Selected Count */}
        <div className="px-6 py-3 bg-blue-500/10 border-b border-white/10">
          <p className="text-sm text-blue-400">
            {selectedAssignees.length > 0 ? '1 Benutzer ausgewählt' : 'Kein Benutzer ausgewählt'}
          </p>
        </div>

        {/* Members List */}
        <div className="flex-1 overflow-y-auto p-4">
          {loading ? (
            <div className="text-center py-8 text-gray-400">Suche läuft...</div>
          ) : !hasSearched ? (
            <div className="text-center py-12 text-gray-400">
              <Search size={48} className="mx-auto mb-4 opacity-50" />
              <p>Geben Sie mindestens 2 Zeichen ein, um Benutzer zu suchen</p>
            </div>
          ) : members.length === 0 ? (
            <div className="text-center py-8 text-gray-400">Keine Benutzer gefunden</div>
          ) : (
            <div className="space-y-2">
              {members.map((member) => {
                const isSelected = selectedAssignees.includes(member.user_id);
                const avatarUrl = getAvatarUrl(member.avatar, member.user_id);
                return (
                  <button
                    key={member.user_id}
                    onClick={() => toggleAssignee(member.user_id)}
                    className={`w-full flex items-center gap-3 p-3 rounded-lg transition-colors ${
                      isSelected
                        ? 'bg-blue-500/20 border border-blue-500'
                        : 'bg-white/5 border border-white/10 hover:bg-white/10'
                    }`}
                  >
                    {avatarUrl ? (
                      <img
                        src={avatarUrl}
                        alt={member.display_name}
                        className="w-10 h-10 rounded-full"
                      />
                    ) : (
                      <div className="w-10 h-10 rounded-full bg-gray-700 flex items-center justify-center">
                        <User size={20} className="text-gray-400" />
                      </div>
                    )}
                    <div className="flex-1 text-left">
                      <div className="font-semibold text-white">{member.display_name}</div>
                      <div className="text-xs text-gray-400">{member.username}</div>
                    </div>
                    {isSelected && (
                      <div className="w-5 h-5 rounded-full bg-blue-500 flex items-center justify-center">
                        <span className="text-white text-xs">✓</span>
                      </div>
                    )}
                  </button>
                );
              })}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex items-center justify-between gap-2 p-4 border-t border-white/10">
          <button
            onClick={handleRemoveAll}
            className="px-4 py-2 bg-red-600/20 hover:bg-red-600/30 text-red-400 border border-red-600/30 rounded-lg transition-colors"
          >
            Alle entfernen
          </button>
          <div className="flex items-center gap-2">
            <button
              onClick={onClose}
              className="px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
            >
              Abbrechen
            </button>
            <button
              onClick={handleAssign}
              disabled={selectedAssignees.length === 0}
              className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
            >
              Benutzer zuweisen
            </button>
          </div>
        </div>
      </div>
    </div>
  );
};

export default AssignModal;
