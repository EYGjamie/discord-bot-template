import { useState, useEffect } from 'react';
import { Settings as SettingsIcon, Shield, MessageSquare, Volume2, Trash2, Plus, Save, X } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';
import { usePermissions } from '../hooks/usePermissions';

interface BotSettings {
  moderator_roles: string[];
  moderation_channel: string;
  log_message_edits: boolean;
  log_message_deletes: boolean;
  notification_users: string[];
  create_voice_settings: CreateVoiceSetting[];
  purge_settings: PurgeSetting[];
}

interface CreateVoiceSetting {
  id: number;
  guild_id: string;
  channel_id: string;
  default_user_limit: number;
  control_channel_id: string;
  control_message_id: string;
  created_at: string;
  updated_at: string;
}

interface PurgeSetting {
  id: number;
  guild_id: string;
  channel_id: string;
  schedule_time: string;
  enabled: boolean;
  last_run: string;
  created_at: string;
  updated_at: string;
}

interface DiscordRole {
  id: string;
  name: string;
  color: number;
  color_hex: string;
  position: number;
}

interface DiscordChannel {
  id: string;
  name: string;
  type: number; // 0 = Text, 2 = Voice
  position: number;
}

export default function SettingsPage() {
  const { user } = useAuth();
  const permissions = usePermissions(user);
  const [loading, setLoading] = useState(true);
  const [saving, setSaving] = useState(false);
  const [settings, setSettings] = useState<BotSettings | null>(null);
  
  // Discord Data
  const [availableRoles, setAvailableRoles] = useState<DiscordRole[]>([]);
  const [availableChannels, setAvailableChannels] = useState<DiscordChannel[]>([]);
  
  // Moderation Settings
  const [selectedRoles, setSelectedRoles] = useState<string[]>([]);
  const [moderationChannel, setModerationChannel] = useState('');
  const [logEdits, setLogEdits] = useState(false);
  const [logDeletes, setLogDeletes] = useState(false);

  // Create Voice Settings
  const [newCreateVoice, setNewCreateVoice] = useState({
    channel_id: '',
    default_user_limit: 0,
    control_channel_id: ''
  });

  // Purge Settings
  const [newPurge, setNewPurge] = useState({
    channel_id: '',
    schedule_time: '00:00',
    enabled: true
  });

  useEffect(() => {
    loadSettings();
    loadDiscordData();
  }, []);

  const loadDiscordData = async () => {
    try {
      const BOT_API_URL = import.meta.env.VITE_BOT_API_URL || 'http://localhost:8090';
      
      // Load Roles
      const rolesResponse = await fetch(`${BOT_API_URL}/api/guild/roles`);
      if (rolesResponse.ok) {
        const rolesData = await rolesResponse.json();
        console.log('Loaded roles:', rolesData);
        setAvailableRoles(rolesData.roles || []);
      } else {
        console.error('Failed to load roles:', rolesResponse.status);
      }

      // Load Channels
      const channelsResponse = await fetch(`${BOT_API_URL}/api/guild/channels`);
      if (channelsResponse.ok) {
        const channelsData = await channelsResponse.json();
        console.log('Loaded channels:', channelsData);
        setAvailableChannels(channelsData.channels || []);
      } else {
        console.error('Failed to load channels:', channelsResponse.status);
      }
    } catch (err) {
      console.error('Failed to load Discord data:', err);
    }
  };

  const loadSettings = async () => {
    try {
      setLoading(true);
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      const response = await fetch(`${API_BASE_URL}/api/bot-settings`, {
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });
      
      if (!response.ok) throw new Error('Failed to load settings');
      
      const data = await response.json();
      setSettings(data);
      setSelectedRoles(data.moderator_roles || []);
      setModerationChannel(data.moderation_channel || '');
      setLogEdits(data.log_message_edits || false);
      setLogDeletes(data.log_message_deletes || false);
    } catch (err) {
      console.error('Failed to load settings:', err);
    } finally {
      setLoading(false);
    }
  };

  const saveModerationSettings = async () => {
    try {
      setSaving(true);
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      
      await fetch(`${API_BASE_URL}/api/bot-settings/moderator-roles`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(selectedRoles),
      });

      await fetch(`${API_BASE_URL}/api/bot-settings/moderation`, {
        method: 'PUT',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify({
          channel_id: moderationChannel || null,
          log_message_edits: logEdits,
          log_message_deletes: logDeletes,
        }),
      });

      alert('Moderation settings saved successfully!');
      await loadSettings();
    } catch (err) {
      console.error('Failed to save settings:', err);
      alert('Failed to save settings');
    } finally {
      setSaving(false);
    }
  };

  const addCreateVoice = async () => {
    if (!newCreateVoice.channel_id) {
      alert('Please enter a voice channel ID');
      return;
    }

    try {
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      await fetch(`${API_BASE_URL}/api/bot-settings/create-voice`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(newCreateVoice),
      });

      setNewCreateVoice({ channel_id: '', default_user_limit: 0, control_channel_id: '' });
      await loadSettings();
    } catch (err) {
      console.error('Failed to add create voice setting:', err);
      alert('Failed to add create voice setting');
    }
  };

  const deleteCreateVoice = async (channelId: string) => {
    if (!confirm('Delete this create voice setting?')) return;

    try {
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      await fetch(`${API_BASE_URL}/api/bot-settings/create-voice/${channelId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      await loadSettings();
    } catch (err) {
      console.error('Failed to delete create voice setting:', err);
      alert('Failed to delete create voice setting');
    }
  };

  const addPurge = async () => {
    if (!newPurge.channel_id || !newPurge.schedule_time) {
      alert('Please enter channel ID and time');
      return;
    }

    try {
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      await fetch(`${API_BASE_URL}/api/bot-settings/purge`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
        body: JSON.stringify(newPurge),
      });

      setNewPurge({ channel_id: '', schedule_time: '00:00', enabled: true });
      await loadSettings();
    } catch (err) {
      console.error('Failed to add purge setting:', err);
      alert('Failed to add purge setting');
    }
  };

  const deletePurge = async (channelId: string) => {
    if (!confirm('Delete this purge schedule?')) return;

    try {
      const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080';
      await fetch(`${API_BASE_URL}/api/bot-settings/purge/${channelId}`, {
        method: 'DELETE',
        headers: {
          'Authorization': `Bearer ${localStorage.getItem('token')}`,
        },
      });

      await loadSettings();
    } catch (err) {
      console.error('Failed to delete purge setting:', err);
      alert('Failed to delete purge setting');
    }
  };

  const getRoleName = (roleId: string) => {
    const role = availableRoles.find(r => r.id === roleId);
    return role ? role.name : roleId;
  };

  const getRoleColor = (roleId: string) => {
    const role = availableRoles.find(r => r.id === roleId);
    return role ? role.color_hex : '#99aab5';
  };

  const getChannelName = (channelId: string) => {
    const channel = availableChannels.find(c => c.id === channelId);
    return channel ? channel.name : channelId;
  };

  const getChannelIcon = (channelId: string) => {
    const channel = availableChannels.find(c => c.id === channelId);
    if (!channel) return '#';
    return channel.type === 2 ? '🔊' : '#'; // 2 = Voice, sonst Text
  };

  if (!permissions.isAdmin) {
    return (
      <div className="min-h-screen bg-[#0f1419] p-6 pt-20">
        <div className="max-w-4xl mx-auto bg-gray-800 rounded-lg p-6 text-center">
          <Shield className="w-16 h-16 mx-auto mb-4 text-red-500" />
          <h2 className="text-2xl font-bold text-white mb-2">Access Denied</h2>
          <p className="text-gray-400">You need administrator permissions to access this page.</p>
        </div>
      </div>
    );
  }

  if (loading) {
    return (
      <div className="min-h-screen bg-[#0f1419] p-6 pt-20 flex items-center justify-center">
        <div className="text-white text-xl">Loading settings...</div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#0f1419] p-6 pt-20">
      <div className="max-w-6xl mx-auto">
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-white mb-2 flex items-center gap-3">
            <SettingsIcon className="w-8 h-8" />
            Bot Settings
          </h1>
          <p className="text-gray-400">Configure your Discord bot settings</p>
        </div>

        {/* Moderation Settings */}
        <div className="bg-gray-800 rounded-lg p-6 mb-6">
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <Shield className="w-6 h-6 text-blue-400" />
            Moderation Settings
          </h2>

          <div className="mb-6">
            <label className="block text-white text-sm font-medium mb-2">Moderator Roles</label>
            <div className="space-y-2 mb-3">
              {selectedRoles.map(roleId => (
                <div key={roleId} className="flex items-center justify-between bg-gray-700 p-3 rounded">
                  <div className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded-full" style={{ backgroundColor: getRoleColor(roleId) }}></div>
                    <span className="text-white font-medium">{getRoleName(roleId)}</span>
                  </div>
                  <button onClick={() => setSelectedRoles(selectedRoles.filter(r => r !== roleId))} className="text-red-400 hover:text-red-300">
                    <X size={18} />
                  </button>
                </div>
              ))}
            </div>
            <select
              onChange={(e) => {
                const roleId = e.target.value;
                if (roleId && !selectedRoles.includes(roleId)) {
                  setSelectedRoles([...selectedRoles, roleId]);
                  e.target.value = '';
                }
              }}
              className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600"
              defaultValue=""
            >
              <option value="" disabled>Select a role...</option>
              {availableRoles
                .filter(role => !selectedRoles.includes(role.id))
                .sort((a, b) => b.position - a.position)
                .map(role => (
                  <option key={role.id} value={role.id}>
                    {role.name}
                  </option>
                ))}
            </select>
          </div>

          <div className="mb-6">
            <label className="block text-white text-sm font-medium mb-2">Moderation Log Channel</label>
            {moderationChannel && (
              <div className="mb-2 flex items-center justify-between bg-gray-700 p-3 rounded">
                <span className="text-white">{getChannelIcon(moderationChannel)} {getChannelName(moderationChannel)}</span>
                <button onClick={() => setModerationChannel('')} className="text-red-400 hover:text-red-300">
                  <X size={18} />
                </button>
              </div>
            )}
            <select
              value={moderationChannel}
              onChange={(e) => setModerationChannel(e.target.value)}
              className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600"
            >
              <option value="">No channel selected</option>
              {availableChannels
                .filter(ch => ch.type === 0)
                .sort((a, b) => a.position - b.position)
                .map(channel => (
                  <option key={channel.id} value={channel.id}>
                    # {channel.name}
                  </option>
                ))}
            </select>
          </div>

          <div className="space-y-3 mb-6">
            <label className="flex items-center gap-3 text-white cursor-pointer">
              <input type="checkbox" checked={logEdits} onChange={(e) => setLogEdits(e.target.checked)} className="w-5 h-5 rounded" />
              <span>Log Message Edits</span>
            </label>
            <label className="flex items-center gap-3 text-white cursor-pointer">
              <input type="checkbox" checked={logDeletes} onChange={(e) => setLogDeletes(e.target.checked)} className="w-5 h-5 rounded" />
              <span>Log Message Deletes</span>
            </label>
          </div>

          <button onClick={saveModerationSettings} disabled={saving} className="w-full bg-blue-600 hover:bg-blue-700 text-white py-2 px-4 rounded flex items-center justify-center gap-2 transition-colors disabled:opacity-50">
            <Save size={18} />
            {saving ? 'Saving...' : 'Save Moderation Settings'}
          </button>
        </div>

        {/* Create Voice Settings */}
        <div className="bg-gray-800 rounded-lg p-6 mb-6">
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <Volume2 className="w-6 h-6 text-green-400" />
            Create Voice Channels
          </h2>

          {settings?.create_voice_settings && settings.create_voice_settings.length > 0 && (
            <div className="space-y-2 mb-4">
              {settings.create_voice_settings.map(cv => (
                <div key={cv.id} className="flex items-center justify-between bg-gray-700 p-3 rounded">
                  <div className="text-white">
                    <div className="font-medium">🔊 {getChannelName(cv.channel_id)}</div>
                    <div className="text-sm text-gray-400">Max Users: {cv.default_user_limit || 'Unlimited'} | Control: # {getChannelName(cv.control_channel_id)}</div>
                  </div>
                  <button onClick={() => deleteCreateVoice(cv.channel_id)} className="text-red-400 hover:text-red-300"><Trash2 size={18} /></button>
                </div>
              ))}
            </div>
          )}

          <div className="space-y-3">
            <div>
              <label className="block text-white text-sm font-medium mb-2">Voice Channel</label>
              <select
                value={newCreateVoice.channel_id}
                onChange={(e) => setNewCreateVoice({ ...newCreateVoice, channel_id: e.target.value })}
                className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600"
              >
                <option value="">Select a voice channel...</option>
                {availableChannels
                  .filter(ch => ch.type === 2)
                  .sort((a, b) => a.position - b.position)
                  .map(channel => (
                    <option key={channel.id} value={channel.id}>
                      🔊 {channel.name}
                    </option>
                  ))}
              </select>
            </div>
            <div>
              <label className="block text-white text-sm font-medium mb-2">Default User Limit (0 = Unlimited)</label>
              <input type="number" min="0" max="99" value={newCreateVoice.default_user_limit} onChange={(e) => setNewCreateVoice({ ...newCreateVoice, default_user_limit: parseInt(e.target.value) || 0 })} className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600" />
            </div>
            <div>
              <label className="block text-white text-sm font-medium mb-2">Control Panel Channel</label>
              <select
                value={newCreateVoice.control_channel_id}
                onChange={(e) => setNewCreateVoice({ ...newCreateVoice, control_channel_id: e.target.value })}
                className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600"
              >
                <option value="">Select a text channel...</option>
                {availableChannels
                  .filter(ch => ch.type === 0)
                  .sort((a, b) => a.position - b.position)
                  .map(channel => (
                    <option key={channel.id} value={channel.id}>
                      # {channel.name}
                    </option>
                  ))}
              </select>
            </div>
            <button onClick={addCreateVoice} className="w-full bg-green-600 hover:bg-green-700 text-white py-2 px-4 rounded flex items-center justify-center gap-2 transition-colors">
              <Plus size={18} />
              Add Create Voice Channel
            </button>
          </div>
        </div>

        {/* Purge Settings */}
        <div className="bg-gray-800 rounded-lg p-6 mb-6">
          <h2 className="text-2xl font-bold text-white mb-4 flex items-center gap-2">
            <MessageSquare className="w-6 h-6 text-orange-400" />
            Auto Purge Schedules
          </h2>

          {settings?.purge_settings && settings.purge_settings.length > 0 && (
            <div className="space-y-2 mb-4">
              {settings.purge_settings.map(purge => (
                <div key={purge.id} className="flex items-center justify-between bg-gray-700 p-3 rounded">
                  <div className="text-white">
                    <div className="font-medium"># {getChannelName(purge.channel_id)}</div>
                    <div className="text-sm text-gray-400">Daily at {purge.schedule_time} | {purge.enabled ? '✅ Enabled' : '❌ Disabled'}</div>
                  </div>
                  <button onClick={() => deletePurge(purge.channel_id)} className="text-red-400 hover:text-red-300"><Trash2 size={18} /></button>
                </div>
              ))}
            </div>
          )}

          <div className="space-y-3">
            <div>
              <label className="block text-white text-sm font-medium mb-2">Channel to Purge</label>
              <select
                value={newPurge.channel_id}
                onChange={(e) => setNewPurge({ ...newPurge, channel_id: e.target.value })}
                className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600"
              >
                <option value="">Select a channel...</option>
                {availableChannels
                  .filter(ch => ch.type === 0)
                  .sort((a, b) => a.position - b.position)
                  .map(channel => (
                    <option key={channel.id} value={channel.id}>
                      # {channel.name}
                    </option>
                  ))}
              </select>
            </div>
            <div>
              <label className="block text-white text-sm font-medium mb-2">Daily Time (24h format)</label>
              <input type="time" value={newPurge.schedule_time} onChange={(e) => setNewPurge({ ...newPurge, schedule_time: e.target.value })} className="w-full bg-gray-700 text-white px-4 py-2 rounded border border-gray-600" />
            </div>
            <label className="flex items-center gap-3 text-white cursor-pointer">
              <input type="checkbox" checked={newPurge.enabled} onChange={(e) => setNewPurge({ ...newPurge, enabled: e.target.checked })} className="w-5 h-5 rounded" />
              <span>Enabled</span>
            </label>
            <button onClick={addPurge} className="w-full bg-orange-600 hover:bg-orange-700 text-white py-2 px-4 rounded flex items-center justify-center gap-2 transition-colors">
              <Plus size={18} />
              Add Purge Schedule
            </button>
          </div>
        </div>
      </div>
    </div>
  );
}
