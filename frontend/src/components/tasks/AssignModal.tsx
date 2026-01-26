import React, { useState, useEffect } from 'react';
import { X, Search, User } from 'lucide-react';
import { api } from '../../services/api';

interface Member {
  id: string;
  display_name: string;
  name: string;
  avatar_url: string;
}

interface AssignModalProps {
  currentAssignees: string[];
  onClose: () => void;
  onAssign: (assignees: string[]) => void;
}

const AssignModal: React.FC<AssignModalProps> = ({ currentAssignees, onClose, onAssign }) => {
  const [members, setMembers] = useState<Member[]>([]);
  const [filteredMembers, setFilteredMembers] = useState<Member[]>([]);
  const [searchTerm, setSearchTerm] = useState('');
  const [selectedAssignees, setSelectedAssignees] = useState<string[]>(currentAssignees);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetchMembers();
  }, []);

  useEffect(() => {
    if (searchTerm.trim() === '') {
      setFilteredMembers(members);
    } else {
      const filtered = members.filter(
        (member) =>
          member.display_name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          member.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
          member.id.includes(searchTerm)
      );
      setFilteredMembers(filtered);
    }
  }, [searchTerm, members]);

  const fetchMembers = async () => {
    try {
      const response = await api.get('/api/members');
      setMembers(response.data);
      setFilteredMembers(response.data);
    } catch (error) {
      console.error('Failed to fetch members:', error);
    } finally {
      setLoading(false);
    }
  };

  const toggleAssignee = (userId: string) => {
    if (selectedAssignees.includes(userId)) {
      setSelectedAssignees(selectedAssignees.filter((id) => id !== userId));
    } else {
      setSelectedAssignees([...selectedAssignees, userId]);
    }
  };

  const handleAssign = () => {
    onAssign(selectedAssignees);
    onClose();
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-[60] p-4">
      <div className="bg-[#1a1d29] border border-white/10 rounded-2xl w-full max-w-2xl max-h-[80vh] overflow-hidden flex flex-col">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-white/10">
          <h2 className="text-xl font-bold text-white">Assign Users</h2>
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
              placeholder="Search by name or user ID..."
              className="w-full pl-10 pr-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </div>
        </div>

        {/* Selected Count */}
        <div className="px-6 py-3 bg-blue-500/10 border-b border-white/10">
          <p className="text-sm text-blue-400">
            {selectedAssignees.length} user{selectedAssignees.length !== 1 ? 's' : ''} selected
          </p>
        </div>

        {/* Members List */}
        <div className="flex-1 overflow-y-auto p-4">
          {loading ? (
            <div className="text-center py-8 text-gray-400">Loading members...</div>
          ) : filteredMembers.length === 0 ? (
            <div className="text-center py-8 text-gray-400">No members found</div>
          ) : (
            <div className="space-y-2">
              {filteredMembers.map((member) => {
                const isSelected = selectedAssignees.includes(member.id);
                return (
                  <button
                    key={member.id}
                    onClick={() => toggleAssignee(member.id)}
                    className={`w-full flex items-center gap-3 p-3 rounded-lg transition-colors ${
                      isSelected
                        ? 'bg-blue-500/20 border border-blue-500'
                        : 'bg-white/5 border border-white/10 hover:bg-white/10'
                    }`}
                  >
                    {member.avatar_url ? (
                      <img
                        src={member.avatar_url}
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
                      <div className="text-xs text-gray-400">{member.name}</div>
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
        <div className="flex items-center justify-end gap-2 p-4 border-t border-white/10">
          <button
            onClick={onClose}
            className="px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
          >
            Cancel
          </button>
          <button
            onClick={handleAssign}
            className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
          >
            Assign Selected
          </button>
        </div>
      </div>
    </div>
  );
};

export default AssignModal;
