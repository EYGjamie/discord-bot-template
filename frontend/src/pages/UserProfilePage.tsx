import { useState, useEffect } from 'react';
import { useParams, Link, useNavigate } from 'react-router-dom';
import {
  ArrowLeft,
  MessageSquare,
  Mic,
  Clock,
  Calendar,
  Crown,
  Users,
  Hash,
  MicOff,
  Volume2,
  Video,
  AlertTriangle,
  FileText,
  Plus,
  Trash2,
} from 'lucide-react';
import { api } from '../services/api';
import type { Member, MemberStats } from '../types';

export default function UserProfilePage() {
  const { userId } = useParams<{ userId: string }>();
  const navigate = useNavigate();
  const [member, setMember] = useState<Member | null>(null);
  const [stats, setStats] = useState<MemberStats | null>(null);
  const [loading, setLoading] = useState(true);
  const [showWarnModal, setShowWarnModal] = useState(false);
  const [showNoteModal, setShowNoteModal] = useState(false);
  const [newReason, setNewReason] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (userId) {
      fetchMemberData(userId);
    }
  }, [userId]);

  const fetchMemberData = async (id: string) => {
    try {
      setLoading(true);
      
      // Fetch member data and stats in parallel
      const [memberData, statsData] = await Promise.all([
        api.members.getMemberById(id),
        api.members.getMemberStats(id),
      ]);

      setMember(memberData);
      setStats(statsData);
    } catch (error) {
      console.error('Failed to fetch member data:', error);
      setMember(null);
      setStats(null);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateWarn = async () => {
    if (!userId || !newReason.trim()) {
      alert('Please fill in all fields');
      return;
    }

    try {
      setSubmitting(true);
      await api.moderation.createWarn({
        user_id: userId,
        reason: newReason,
      });
      setShowWarnModal(false);
      setNewReason('');
      // Refresh stats
      if (userId) {
        const statsData = await api.members.getMemberStats(userId);
        setStats(statsData);
      }
    } catch (error) {
      console.error('Failed to create warn:', error);
      alert('Failed to create warn');
    } finally {
      setSubmitting(false);
    }
  };

  const handleCreateNote = async () => {
    if (!userId || !newReason.trim()) {
      alert('Please fill in all fields');
      return;
    }

    try {
      setSubmitting(true);
      await api.moderation.createNote({
        user_id: userId,
        reason: newReason,
      });
      setShowNoteModal(false);
      setNewReason('');
      // Refresh stats
      if (userId) {
        const statsData = await api.members.getMemberStats(userId);
        setStats(statsData);
      }
    } catch (error) {
      console.error('Failed to create note:', error);
      alert('Failed to create note');
    } finally {
      setSubmitting(false);
    }
  };

  const handleDeleteWarn = async (warnId: number) => {
    if (!confirm('Are you sure you want to delete this warn?')) return;

    try {
      await api.moderation.deleteWarn(warnId);
      // Refresh stats
      if (userId) {
        const statsData = await api.members.getMemberStats(userId);
        setStats(statsData);
      }
    } catch (error) {
      console.error('Failed to delete warn:', error);
      alert('Failed to delete warn');
    }
  };

  const handleDeleteNote = async (noteId: number) => {
    if (!confirm('Are you sure you want to delete this note?')) return;

    try {
      await api.moderation.deleteNote(noteId);
      // Refresh stats
      if (userId) {
        const statsData = await api.members.getMemberStats(userId);
        setStats(statsData);
      }
    } catch (error) {
      console.error('Failed to delete note:', error);
      alert('Failed to delete note');
    }
  };

  const formatDuration = (seconds: number): string => {
    const hours = Math.floor(seconds / 3600);
    const minutes = Math.floor((seconds % 3600) / 60);
    return `${hours}h ${minutes}m`;
  };

  const formatDate = (dateString: string | null): string => {
    if (!dateString) return 'N/A';
    return new Date(dateString).toLocaleDateString('de-DE', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getAvatarUrl = () => {
    if (member?.avatar_url) return member.avatar_url;
    return `https://cdn.discordapp.com/embed/avatars/${parseInt(userId || '0') % 5}.png`;
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center h-full">
        <div className="text-gray-400">Loading member profile...</div>
      </div>
    );
  }

  if (!member || !stats) {
    return (
      <div className="flex flex-col items-center justify-center h-full gap-4">
        <div className="text-gray-400">Member not found</div>
        <Link
          to="/members"
          className="px-4 py-2 bg-cyan-500 text-white rounded-lg hover:bg-cyan-600 transition-colors"
        >
          Back to Members
        </Link>
      </div>
    );
  }

  return (
    <div className="space-y-6 p-6">
      {/* Back Button */}
      <button
        onClick={() => navigate('/members')}
        className="flex items-center gap-2 text-gray-400 hover:text-white transition-colors px-4 py-2 bg-[#1a1f2e] rounded-lg border border-gray-800 hover:border-cyan-500"
      >
        <ArrowLeft className="w-5 h-5" />
        Back to Members
      </button>

      {/* Profile Header */}
      <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
        <div className="flex flex-col md:flex-row gap-6">
          {/* Avatar */}
          <div className="flex-shrink-0">
            <img
              src={getAvatarUrl()}
              alt={member.display_name}
              className="w-32 h-32 rounded-full border-4 border-gray-700"
            />
          </div>

          {/* Basic Info */}
          <div className="flex-1 space-y-4">
            <div>
              <h1 className="text-3xl font-bold text-white">
                {member.display_name || member.name}
              </h1>
              {member.nick && member.nick !== member.display_name && (
                <p className="text-gray-400 mt-1">@{member.name}</p>
              )}
              <p className="text-gray-500 text-sm mt-1">{member.mention}</p>
            </div>

            {/* Roles */}
            <div className="flex flex-wrap gap-2">
              {stats.roles.map(role => (
                <div
                  key={role.id}
                  className="px-3 py-1 rounded-full text-sm font-medium border"
                  style={{
                    backgroundColor: role.color ? `${role.color}15` : '#4b556315',
                    color: role.color || '#9ca3af',
                    borderColor: role.color ? `${role.color}40` : '#4b556340',
                  }}
                >
                  {role.name}
                </div>
              ))}
            </div>

            {/* Dates */}
            <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
              <div className="flex items-center gap-2 text-gray-400">
                <Calendar className="w-4 h-4" />
                <div>
                  <p className="text-xs">Account Created</p>
                  <p className="text-white text-sm">{formatDate(member.created_at)}</p>
                </div>
              </div>
              <div className="flex items-center gap-2 text-gray-400">
                <Users className="w-4 h-4" />
                <div>
                  <p className="text-xs">Joined Server</p>
                  <p className="text-white text-sm">{formatDate(member.joined_at)}</p>
                </div>
              </div>
              {member.premium_since && (
                <div className="flex items-center gap-2 text-gray-400">
                  <Crown className="w-4 h-4 text-pink-500" />
                  <div>
                    <p className="text-xs">Server Booster</p>
                    <p className="text-white text-sm">{formatDate(member.premium_since)}</p>
                  </div>
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Statistics Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
        {/* Total Messages */}
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Total Messages</p>
              <p className="text-2xl font-bold text-white mt-1">
                {stats.total_messages.toLocaleString()}
              </p>
            </div>
            <MessageSquare className="w-8 h-8 text-cyan-500" />
          </div>
        </div>

        {/* Total Joins */}
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Server Joins</p>
              <p className="text-2xl font-bold text-white mt-1">{stats.total_joins}</p>
            </div>
            <Users className="w-8 h-8 text-green-500" />
          </div>
        </div>

        {/* Total Leaves */}
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Server Leaves</p>
              <p className="text-2xl font-bold text-white mt-1">{stats.total_leaves}</p>
            </div>
            <Clock className="w-8 h-8 text-orange-500" />
          </div>
        </div>

        {/* Total Warns */}
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center justify-between">
            <div>
              <p className="text-gray-400 text-sm">Total Warns</p>
              <p className="text-2xl font-bold text-white mt-1">{stats.total_warns}</p>
            </div>
            <MessageSquare className="w-8 h-8 text-red-500" />
          </div>
        </div>
      </div>

      {/* Top Text Channel */}
      {stats.top_text_channel && (
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center gap-3 mb-4">
            <Hash className="w-6 h-6 text-cyan-500" />
            <h2 className="text-xl font-bold text-white">Top Text Channel</h2>
          </div>
          <div className="space-y-3">
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Channel</span>
              <span className="text-white font-medium">#{stats.top_text_channel.name}</span>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-gray-400">Messages</span>
              <span className="text-white font-medium">
                {stats.top_text_channel.message_count.toLocaleString()}
              </span>
            </div>
            <div className="w-full bg-gray-700 rounded-full h-2 mt-2">
              <div
                className="bg-cyan-500 h-2 rounded-full"
                style={{
                  width: `${Math.min(100, (stats.top_text_channel.message_count / stats.total_messages) * 100)}%`,
                }}
              />
            </div>
            <p className="text-gray-500 text-sm">
              {stats.total_messages > 0 ? ((stats.top_text_channel.message_count / stats.total_messages) * 100).toFixed(1) : 0}%
              of total messages
            </p>
          </div>
        </div>
      )}

      {/* Voice Activity Details */}
      <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
        <div className="flex items-center gap-3 mb-4">
          <Mic className="w-6 h-6 text-green-500" />
          <h2 className="text-xl font-bold text-white">Voice Activity Details</h2>
        </div>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {/* Total Voice Time */}
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Mic className="w-5 h-5 text-green-500" />
              <p className="text-gray-400 text-sm">Total Voice Time</p>
            </div>
            <p className="text-white font-medium text-xl">{formatDuration(stats.total_voice_time)}</p>
          </div>
          
          {/* Voice Channel Joins */}
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Users className="w-5 h-5 text-cyan-500" />
              <p className="text-gray-400 text-sm">Voice Channel Joins</p>
            </div>
            <p className="text-white font-medium text-xl">{stats.join_count}</p>
          </div>
          
          {/* Muted Time */}
          <div>
            <div className="flex items-center gap-2 mb-2">
              <MicOff className="w-5 h-5 text-red-500" />
              <p className="text-gray-400 text-sm">Muted Time</p>
            </div>
            <p className="text-white font-medium text-xl">{formatDuration(stats.muted_duration)}</p>
          </div>
          
          {/* Deafened Time */}
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Volume2 className="w-5 h-5 text-gray-500" />
              <p className="text-gray-400 text-sm">Deafened Time</p>
            </div>
            <p className="text-white font-medium text-xl">{formatDuration(stats.deafen_duration)}</p>
          </div>
          
          {/* Streaming Time */}
          <div>
            <div className="flex items-center gap-2 mb-2">
              <Video className="w-5 h-5 text-purple-500" />
              <p className="text-gray-400 text-sm">Streaming Time</p>
            </div>
            <p className="text-white font-medium text-xl">{formatDuration(stats.stream_duration)}</p>
          </div>
          
          {/* Top Voice Channel */}
          {stats.top_voice_channel && (
            <div>
              <div className="flex items-center gap-2 mb-2">
                <Mic className="w-5 h-5 text-green-500" />
                <p className="text-gray-400 text-sm">Most Active Channel</p>
              </div>
              <p className="text-white font-medium text-lg">{stats.top_voice_channel.name}</p>
              <p className="text-gray-400 text-sm">{formatDuration(stats.top_voice_channel.duration)}</p>
            </div>
          )}
        </div>
      </div>

      {/* Moderation Section */}
      <div className="grid grid-cols-1 lg:grid-cols-2 gap-6">
        {/* Warns */}
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <AlertTriangle className="w-6 h-6 text-red-500" />
              <h2 className="text-xl font-bold text-white">Warns ({stats.warns.length})</h2>
            </div>
            <button
              onClick={() => setShowWarnModal(true)}
              className="flex items-center gap-2 px-3 py-1.5 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors text-sm"
            >
              <Plus className="w-4 h-4" />
              Add Warn
            </button>
          </div>
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {stats.warns.length === 0 ? (
              <p className="text-gray-400 text-sm">No warns recorded</p>
            ) : (
              stats.warns.map(warn => (
                <div key={warn.id} className="bg-[#0f1419] rounded-lg p-4 border border-red-900/30 relative group">
                  <button
                    onClick={() => handleDeleteWarn(warn.id)}
                    className="absolute top-3 right-3 p-1.5 bg-red-600/20 hover:bg-red-600 text-red-400 hover:text-white rounded transition-colors opacity-0 group-hover:opacity-100"
                    title="Delete warn"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                  <div className="flex items-start justify-between mb-2 pr-10">
                    <span className="text-red-400 text-xs font-medium">WARN #{warn.id}</span>
                    <span className="text-gray-500 text-xs">
                      {new Date(warn.created_at).toLocaleDateString('de-DE', {
                        day: '2-digit',
                        month: '2-digit',
                        year: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                  </div>
                  <p className="text-white text-sm mb-2">{warn.reason}</p>
                  <p className="text-gray-400 text-xs">
                    Moderator: <span className="text-cyan-400">{warn.moderator_name}</span>
                  </p>
                </div>
              ))
            )}
          </div>
        </div>

        {/* Notes */}
        <div className="bg-[#1a1f2e] rounded-lg p-6 border border-gray-800">
          <div className="flex items-center justify-between mb-4">
            <div className="flex items-center gap-3">
              <FileText className="w-6 h-6 text-blue-500" />
              <h2 className="text-xl font-bold text-white">Notes ({stats.notes.length})</h2>
            </div>
            <button
              onClick={() => setShowNoteModal(true)}
              className="flex items-center gap-2 px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm"
            >
              <Plus className="w-4 h-4" />
              Add Note
            </button>
          </div>
          <div className="space-y-3 max-h-96 overflow-y-auto">
            {stats.notes.length === 0 ? (
              <p className="text-gray-400 text-sm">No notes recorded</p>
            ) : (
              stats.notes.map(note => (
                <div key={note.id} className="bg-[#0f1419] rounded-lg p-4 border border-blue-900/30 relative group">
                  <button
                    onClick={() => handleDeleteNote(note.id)}
                    className="absolute top-3 right-3 p-1.5 bg-blue-600/20 hover:bg-blue-600 text-blue-400 hover:text-white rounded transition-colors opacity-0 group-hover:opacity-100"
                    title="Delete note"
                  >
                    <Trash2 className="w-4 h-4" />
                  </button>
                  <div className="flex items-start justify-between mb-2 pr-10">
                    <span className="text-blue-400 text-xs font-medium">NOTE #{note.id}</span>
                    <span className="text-gray-500 text-xs">
                      {new Date(note.created_at).toLocaleDateString('de-DE', {
                        day: '2-digit',
                        month: '2-digit',
                        year: 'numeric',
                        hour: '2-digit',
                        minute: '2-digit',
                      })}
                    </span>
                  </div>
                  <p className="text-white text-sm mb-2">{note.reason}</p>
                  <p className="text-gray-400 text-xs">
                    Moderator: <span className="text-cyan-400">{note.moderator_name}</span>
                  </p>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      {/* Warn Modal */}
      {showWarnModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1a1f2e] rounded-lg p-6 w-full max-w-md border border-gray-800">
            <h3 className="text-xl font-bold text-white mb-4">Add Warn</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Reason</label>
                <textarea
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  className="w-full bg-[#0f1419] border border-gray-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-cyan-500 min-h-[100px]"
                  placeholder="Enter reason for warn"
                />
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => {
                    setShowWarnModal(false);
                    setNewReason('');
                  }}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
                  disabled={submitting}
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateWarn}
                  className="flex-1 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50"
                  disabled={submitting || !newReason.trim()}
                >
                  {submitting ? 'Creating...' : 'Create Warn'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Note Modal */}
      {showNoteModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-[#1a1f2e] rounded-lg p-6 w-full max-w-md border border-gray-800">
            <h3 className="text-xl font-bold text-white mb-4">Add Note</h3>
            <div className="space-y-4">
              <div>
                <label className="block text-sm font-medium text-gray-400 mb-2">Note</label>
                <textarea
                  value={newReason}
                  onChange={(e) => setNewReason(e.target.value)}
                  className="w-full bg-[#0f1419] border border-gray-700 rounded-lg px-4 py-2 text-white focus:outline-none focus:border-cyan-500 min-h-[100px]"
                  placeholder="Enter note text"
                />
              </div>
              <div className="flex gap-3">
                <button
                  onClick={() => {
                    setShowNoteModal(false);
                    setNewReason('');
                  }}
                  className="flex-1 px-4 py-2 bg-gray-700 hover:bg-gray-600 text-white rounded-lg transition-colors"
                  disabled={submitting}
                >
                  Cancel
                </button>
                <button
                  onClick={handleCreateNote}
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50"
                  disabled={submitting || !newReason.trim()}
                >
                  {submitting ? 'Creating...' : 'Create Note'}
                </button>
              </div>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}