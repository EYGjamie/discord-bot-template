import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Plus, Trash2, Shield, User, Users, Search, Edit, Save, Palette } from 'lucide-react';
import { boardsService } from '../services/tasks';
import { api } from '../services/api';
import type { Board, BoardPermission } from '../types/tasks';
import type { Role, MemberInfo, RolesAndMembersResponse } from '../types/discord';

const BOARD_COLORS = [
  { name: 'Blau', value: '#3b82f6' },
  { name: 'Grün', value: '#10b981' },
  { name: 'Gelb', value: '#f59e0b' },
  { name: 'Rot', value: '#ef4444' },
  { name: 'Lila', value: '#8b5cf6' },
  { name: 'Pink', value: '#ec4899' },
  { name: 'Türkis', value: '#06b6d4' },
  { name: 'Orange', value: '#f97316' },
  { name: 'Indigo', value: '#6366f1' },
  { name: 'Lime', value: '#84cc16' },
];

const BoardSettingsPage: React.FC = () => {
  const { boardId } = useParams<{ boardId: string }>();
  const navigate = useNavigate();
  const [board, setBoard] = useState<Board | null>(null);
  const [permissions, setPermissions] = useState<BoardPermission[]>([]);
  const [roles, setRoles] = useState<Role[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);
  const [showEditModal, setShowEditModal] = useState(false);
  const [editingPermission, setEditingPermission] = useState<BoardPermission | null>(null);
  const [permissionType, setPermissionType] = useState<'role' | 'user'>('role');
  const [searchQuery, setSearchQuery] = useState('');
  const [searchResults, setSearchResults] = useState<MemberInfo[]>([]);
  const [selectedMember, setSelectedMember] = useState<MemberInfo | null>(null);
  const [activeTab, setActiveTab] = useState<'general' | 'permissions'>('general');
  const [boardForm, setBoardForm] = useState({
    name: '',
    description: '',
    color: '#3b82f6',
  });
  const [savingBoard, setSavingBoard] = useState(false);

  useEffect(() => {
    if (boardId) {
      loadBoard();
      loadPermissions();
      loadRolesAndMembers();
    }
  }, [boardId]);

  useEffect(() => {
    if (searchQuery.length >= 2 && permissionType === 'user') {
      searchMembers();
    } else {
      setSearchResults([]);
    }
  }, [searchQuery, permissionType]);

  const loadBoard = async () => {
    try {
      const data = await boardsService.getById(Number(boardId));
      setBoard(data);
      setBoardForm({
        name: data.name,
        description: data.description || '',
        color: data.color || '#3b82f6',
      });
    } catch (err: any) {
      setError(err?.message || 'Failed to load board');
    }
  };

  const loadPermissions = async () => {
    try {
      setLoading(true);
      const data = await boardsService.getPermissions(Number(boardId));
      setPermissions(data || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load permissions');
      setPermissions([]);
    } finally {
      setLoading(false);
    }
  };

  const loadRolesAndMembers = async () => {
    try {
      const data: RolesAndMembersResponse = await api.get('/api/discord/roles-and-members');
      setRoles(data.roles || []);
    } catch (err: any) {
      console.error('Failed to load roles and members:', err);
    }
  };

  const searchMembers = async () => {
    try {
      const results: MemberInfo[] = await api.get(`/api/discord/members/search?q=${encodeURIComponent(searchQuery)}`);
      setSearchResults(results || []);
    } catch (err: any) {
      console.error('Failed to search members:', err);
    }
  };

  const handleAddPermission = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    
    const roleId = permissionType === 'role' ? (formData.get('role_id') as string) : undefined;
    const userId = permissionType === 'user' ? (selectedMember?.user_id) : undefined;
    const canViewBoard = formData.get('can_view_board') === 'on';
    const canViewTaskList = formData.get('can_view_task_list') === 'on';
    const canViewTasks = formData.get('can_view_tasks') === 'on';
    const canEditTasks = formData.get('can_edit_tasks') === 'on';
    const canEditBoard = formData.get('can_edit_board') === 'on';

    if (!roleId && !userId) {
      setError('Please select a role or user');
      return;
    }

    try {
      await boardsService.setPermission(Number(boardId), {
        role_id: roleId,
        user_id: userId,
        can_view_board: canViewBoard,
        can_view_task_list: canViewTaskList,
        can_view_tasks: canViewTasks,
        can_edit_tasks: canEditTasks,
        can_edit_board: canEditBoard,
      });
      setShowAddModal(false);
      setSelectedMember(null);
      setSearchQuery('');
      loadPermissions();
    } catch (err: any) {
      setError(err?.message || 'Failed to add permission');
    }
  };

  const handleEditPermission = (permission: BoardPermission) => {
    setEditingPermission(permission);
    setShowEditModal(true);
  };

  const handleUpdatePermission = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    if (!editingPermission) return;

    const formData = new FormData(e.currentTarget);
    const canViewBoard = formData.get('can_view_board') === 'on';
    const canViewTaskList = formData.get('can_view_task_list') === 'on';
    const canViewTasks = formData.get('can_view_tasks') === 'on';
    const canEditTasks = formData.get('can_edit_tasks') === 'on';
    const canEditBoard = formData.get('can_edit_board') === 'on';

    try {
      await boardsService.updatePermission(Number(boardId), editingPermission.id, {
        can_view_board: canViewBoard,
        can_view_task_list: canViewTaskList,
        can_view_tasks: canViewTasks,
        can_edit_tasks: canEditTasks,
        can_edit_board: canEditBoard,
      });
      setShowEditModal(false);
      setEditingPermission(null);
      loadPermissions();
    } catch (err: any) {
      setError(err?.message || 'Failed to update permission');
    }
  };

  const handleDeletePermission = async (permissionId: number) => {
    if (!confirm('Are you sure you want to delete this permission?')) return;

    try {
      await boardsService.deletePermission(Number(boardId), permissionId);
      loadPermissions();
    } catch (err: any) {
      setError(err?.message || 'Failed to delete permission');
    }
  };

  const handleBoardUpdate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!board) return;

    try {
      setSavingBoard(true);
      setError(null);
      await boardsService.update(board.id, boardForm);
      await loadBoard();
      setSavingBoard(false);
    } catch (err: any) {
      setError(err?.message || 'Failed to update board');
      setSavingBoard(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!board) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg">
          Board not found
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      {/* Header */}
      <div className="flex items-center gap-4 mb-8">
        <button
          onClick={() => navigate(`/tasks/boards/${boardId}`)}
          className="p-2 hover:bg-white/10 rounded-lg transition-colors text-white"
        >
          <ArrowLeft size={24} />
        </button>
        <div>
          <h1 className="text-3xl font-bold text-white">{board.name} - Einstellungen</h1>
          <p className="text-gray-400 mt-1">Board-Einstellungen und Berechtigungen verwalten</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {/* Tabs */}
      <div className="flex gap-2 mb-6 border-b border-white/10">
        <button
          onClick={() => setActiveTab('general')}
          className={`px-6 py-3 font-medium transition-colors relative ${
            activeTab === 'general'
              ? 'text-blue-400'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          Allgemein
          {activeTab === 'general' && (
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-400" />
          )}
        </button>
        <button
          onClick={() => setActiveTab('permissions')}
          className={`px-6 py-3 font-medium transition-colors relative ${
            activeTab === 'permissions'
              ? 'text-blue-400'
              : 'text-gray-400 hover:text-white'
          }`}
        >
          <div className="flex items-center gap-2">
            <Shield size={18} />
            Berechtigungen
          </div>
          {activeTab === 'permissions' && (
            <div className="absolute bottom-0 left-0 right-0 h-0.5 bg-blue-400" />
          )}
        </button>
      </div>

      {/* General Settings Tab */}
      {activeTab === 'general' && (
        <div className="bg-white/5 border border-white/10 rounded-2xl p-8">
          <h2 className="text-2xl font-bold text-white mb-6">Board-Einstellungen</h2>
          
          <form onSubmit={handleBoardUpdate} className="space-y-6">
            {/* Name */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Board-Name *
              </label>
              <input
                type="text"
                value={boardForm.name}
                onChange={(e) => setBoardForm({ ...boardForm, name: e.target.value })}
                required
                className="w-full px-4 py-3 bg-[#0d0f15] border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                placeholder="z.B. Sprint Planning"
              />
            </div>

            {/* Description */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-2">
                Beschreibung
              </label>
              <textarea
                value={boardForm.description}
                onChange={(e) => setBoardForm({ ...boardForm, description: e.target.value })}
                rows={4}
                className="w-full px-4 py-3 bg-[#0d0f15] border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 resize-none"
                placeholder="Beschreiben Sie das Board..."
              />
            </div>

            {/* Color Picker */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-3 flex items-center gap-2">
                <Palette size={18} />
                Board-Farbe
              </label>
              <div className="grid grid-cols-5 md:grid-cols-10 gap-3">
                {BOARD_COLORS.map((color) => (
                  <button
                    key={color.value}
                    type="button"
                    onClick={() => setBoardForm({ ...boardForm, color: color.value })}
                    className={`relative h-16 rounded-lg transition-all ${
                      boardForm.color === color.value
                        ? 'ring-2 ring-white ring-offset-2 ring-offset-[#1e2228] scale-110'
                        : 'hover:scale-105'
                    }`}
                    style={{ backgroundColor: color.value }}
                    title={color.name}
                  >
                    {boardForm.color === color.value && (
                      <div className="absolute inset-0 flex items-center justify-center">
                        <span className="text-white text-2xl">✓</span>
                      </div>
                    )}
                  </button>
                ))}
              </div>
              <p className="text-sm text-gray-400 mt-2">
                Diese Farbe wird als Akzent im Board verwendet
              </p>
            </div>

            {/* Preview */}
            <div>
              <label className="block text-sm font-medium text-gray-300 mb-3">
                Vorschau
              </label>
              <div className="bg-white/5 border border-white/10 rounded-xl overflow-hidden flex">
                <div
                  className="w-1.5 flex-shrink-0"
                  style={{ backgroundColor: boardForm.color }}
                />
                <div className="flex-1 p-4 flex items-center gap-4">
                  <div
                    className="w-12 h-12 rounded-xl flex items-center justify-center text-2xl"
                    style={{ backgroundColor: boardForm.color + '20', color: boardForm.color }}
                  >
                    📋
                  </div>
                  <div className="flex-1">
                    <h3 className="text-lg font-bold text-white mb-1">
                      {boardForm.name || 'Board-Name'}
                    </h3>
                    <p className="text-sm text-gray-400 line-clamp-2">
                      {boardForm.description || 'Keine Beschreibung'}
                    </p>
                  </div>
                </div>
              </div>
            </div>

            {/* Actions */}
            <div className="flex items-center justify-end gap-3 pt-6 border-t border-white/10">
              <button
                type="submit"
                disabled={savingBoard || !boardForm.name.trim()}
                className="flex items-center gap-2 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              >
                <Save size={18} />
                {savingBoard ? 'Speichert...' : 'Änderungen speichern'}
              </button>
            </div>
          </form>
        </div>
      )}

      {/* Permissions Tab */}
      {activeTab === 'permissions' && (
        <>
          {/* Info Box */}
          <div className="bg-blue-500/10 border border-blue-500/20 rounded-xl p-4 mb-6">
            <div className="flex items-start gap-3">
              <Shield className="text-blue-400 mt-1" size={20} />
              <div>
                <h3 className="text-white font-semibold mb-1">Berechtigungssystem</h3>
                <p className="text-sm text-gray-400">
                  Wenn keine Berechtigungen gesetzt sind, haben alle authentifizierten Benutzer vollen Zugriff auf dieses Board.
                  Sobald Sie Berechtigungen hinzufügen, können nur Benutzer/Rollen mit expliziten Berechtigungen auf das Board zugreifen.
                </p>
              </div>
            </div>
          </div>

          {/* Add Permission Button */}
          <div className="flex justify-between items-center mb-6">
            <h2 className="text-xl font-bold text-white">Board Permissions</h2>
            <button
              onClick={() => setShowAddModal(true)}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
              <Plus size={20} />
              Add Permission
            </button>
          </div>

          {/* Permissions List */}
          <div className="space-y-3">
        {permissions.length === 0 ? (
          <div className="bg-white/5 border border-white/10 rounded-xl p-8 text-center">
            <Shield className="mx-auto mb-4 text-gray-500" size={48} />
            <p className="text-gray-400">
              No specific permissions set. All users have full access.
            </p>
          </div>
        ) : (
          permissions.map((permission) => (
            <div
              key={permission.id}
              className="bg-white/5 border border-white/10 rounded-xl p-4 flex items-center justify-between hover:bg-white/10 transition-colors"
            >
              <div className="flex items-center gap-4 flex-1">
                {permission.role_id ? (
                  <div className="flex items-center gap-3">
                    <Users className="text-purple-400" size={24} />
                    <div>
                      <div className="text-white font-bold text-lg">{permission.role_name || 'Unknown Role'}</div>
                      <div className="text-xs text-gray-500">ID: {permission.role_id}</div>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center gap-3">
                    {permission.user_avatar_url ? (
                      <img 
                        src={permission.user_avatar_url} 
                        alt={permission.user_display_name || permission.user_name || 'User'}
                        className="w-10 h-10 rounded-full object-cover"
                      />
                    ) : (
                      <User className="text-blue-400" size={24} />
                    )}
                    <div>
                      <div className="text-white font-bold text-lg">{permission.user_display_name || permission.user_name || 'Unknown User'}</div>
                      <div className="text-xs text-gray-500">@{permission.user_name} • ID: {permission.user_id}</div>
                    </div>
                  </div>
                )}
                <div className="flex gap-2 ml-8 flex-wrap">
                  {permission.can_view_board && (
                    <span className="px-2 py-1 bg-green-500/20 text-green-400 rounded text-xs font-semibold">
                      View Board
                    </span>
                  )}
                  {permission.can_view_task_list && (
                    <span className="px-2 py-1 bg-blue-500/20 text-blue-400 rounded text-xs font-semibold">
                      View Task List
                    </span>
                  )}
                  {permission.can_view_tasks && (
                    <span className="px-2 py-1 bg-purple-500/20 text-purple-400 rounded text-xs font-semibold">
                      View Tasks
                    </span>
                  )}
                  {permission.can_edit_tasks && (
                    <span className="px-2 py-1 bg-orange-500/20 text-orange-400 rounded text-xs font-semibold">
                      Edit Tasks
                    </span>
                  )}
                  {permission.can_edit_board && (
                    <span className="px-2 py-1 bg-red-500/20 text-red-400 rounded text-xs font-semibold">
                      Edit Board
                    </span>
                  )}
                </div>
              </div>
              <div className="flex items-center gap-2">
                <button
                  onClick={() => handleEditPermission(permission)}
                  className="p-2 hover:bg-blue-500/20 text-blue-400 rounded-lg transition-colors"
                  title="Edit permissions"
                >
                  <Edit size={18} />
                </button>
                <button
                  onClick={() => handleDeletePermission(permission.id)}
                  className="p-2 hover:bg-red-500/20 text-red-400 rounded-lg transition-colors"
                  title="Delete permission"
                >
                  <Trash2 size={18} />
                </button>
              </div>
            </div>
          ))
        )}
      </div>

      {/* Add Permission Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md max-h-[90vh] overflow-y-auto">
            <div className="p-6 border-b border-white/10">
              <h2 className="text-xl font-bold text-white">Add Permission</h2>
            </div>
            <form onSubmit={handleAddPermission} className="p-6 space-y-4">
              {/* Permission Type Selector */}
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Permission Type
                </label>
                <div className="flex gap-2">
                  <button
                    type="button"
                    onClick={() => setPermissionType('role')}
                    className={`flex-1 px-4 py-2 rounded-lg transition-colors ${
                      permissionType === 'role'
                        ? 'bg-blue-600 text-white'
                        : 'bg-white/5 text-gray-400 hover:bg-white/10'
                    }`}
                  >
                    <Users size={16} className="inline mr-2" />
                    Role
                  </button>
                  <button
                    type="button"
                    onClick={() => setPermissionType('user')}
                    className={`flex-1 px-4 py-2 rounded-lg transition-colors ${
                      permissionType === 'user'
                        ? 'bg-blue-600 text-white'
                        : 'bg-white/5 text-gray-400 hover:bg-white/10'
                    }`}
                  >
                    <User size={16} className="inline mr-2" />
                    User
                  </button>
                </div>
              </div>

              {/* Role Selection */}
              {permissionType === 'role' && (
                <div>
                  <label className="block text-sm font-semibold text-white mb-2">
                    Select Role
                  </label>
                  <select
                    name="role_id"
                    required
                    className="w-full px-4 py-2.5 bg-gray-800 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 cursor-pointer hover:bg-gray-750 transition-colors"
                    style={{
                      colorScheme: 'dark',
                    }}
                  >
                    <option value="" className="bg-gray-800 text-gray-400">Choose a role...</option>
                    {roles.map((role) => (
                      <option key={role.id} value={role.id} className="bg-gray-800 text-white py-2">
                        {role.name}
                      </option>
                    ))}
                  </select>
                </div>
              )}

              {/* User Search */}
              {permissionType === 'user' && (
                <div>
                  <label className="block text-sm font-semibold text-white mb-2">
                    Search User
                  </label>
                  <div className="relative">
                    <div className="relative">
                      <Search className="absolute left-3 top-3 text-gray-500" size={16} />
                      <input
                        type="text"
                        value={searchQuery}
                        onChange={(e) => setSearchQuery(e.target.value)}
                        placeholder="Search by username..."
                        className="w-full pl-10 pr-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                      />
                    </div>
                    {searchResults.length > 0 && (
                      <div className="absolute z-50 w-full mt-2 bg-gray-800 border border-white/20 rounded-lg shadow-2xl max-h-48 overflow-y-auto top-full">
                        {searchResults.map((member) => (
                          <button
                            key={member.user_id}
                            type="button"
                            onClick={() => {
                              setSelectedMember(member);
                              setSearchQuery(member.display_name);
                              setSearchResults([]);
                            }}
                            className="w-full px-4 py-2.5 text-left hover:bg-blue-600/20 hover:border-l-2 hover:border-blue-500 transition-all text-white first:rounded-t-lg last:rounded-b-lg flex items-center gap-3"
                          >
                            {member.avatar ? (
                              <img 
                                src={`https://cdn.discordapp.com/avatars/${member.user_id}/${member.avatar}.png?size=64`}
                                alt={member.display_name}
                                className="w-8 h-8 rounded-full object-cover flex-shrink-0"
                              />
                            ) : (
                              <div className="w-8 h-8 rounded-full bg-blue-600 flex items-center justify-center flex-shrink-0">
                                <User size={16} className="text-white" />
                              </div>
                            )}
                            <div className="flex-1 min-w-0">
                              <div className="font-semibold truncate">{member.display_name}</div>
                              <div className="text-xs text-gray-400 truncate">{member.username}</div>
                            </div>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                  {selectedMember && (
                    <div className="mt-2 p-3 bg-blue-500/20 border border-blue-500/30 rounded-lg flex items-center gap-3">
                      {selectedMember.avatar ? (
                        <img 
                          src={`https://cdn.discordapp.com/avatars/${selectedMember.user_id}/${selectedMember.avatar}.png?size=64`}
                          alt={selectedMember.display_name}
                          className="w-10 h-10 rounded-full object-cover flex-shrink-0"
                        />
                      ) : (
                        <div className="w-10 h-10 rounded-full bg-blue-600 flex items-center justify-center flex-shrink-0">
                          <User size={20} className="text-white" />
                        </div>
                      )}
                      <div className="flex-1 min-w-0">
                        <div className="text-blue-300 font-semibold truncate">{selectedMember.display_name}</div>
                        <div className="text-xs text-blue-400/70 truncate">@{selectedMember.username}</div>
                      </div>
                    </div>
                  )}
                </div>
              )}

              {/* Permissions */}
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Permissions
                </label>
                <div className="space-y-2">
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_view_board"
                      defaultChecked
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can View Board</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_view_task_list"
                      defaultChecked
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can View Task Titles</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_view_tasks"
                      defaultChecked
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can View Full Tasks</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_edit_tasks"
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can Edit Tasks</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_edit_board"
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can Edit Board</span>
                  </label>
                </div>
              </div>

              <div className="flex gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowAddModal(false);
                    setSelectedMember(null);
                    setSearchQuery('');
                  }}
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                >
                  Add Permission
                </button>
              </div>
            </form>
          </div>
        </div>
      )}

      {/* Edit Permission Modal */}
      {showEditModal && editingPermission && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md">
            <div className="p-6 border-b border-white/10">
              <h2 className="text-xl font-bold text-white">Edit Permission</h2>
              <p className="text-sm text-gray-400 mt-1">
                {editingPermission.role_id 
                  ? `Role: ${editingPermission.role_name || editingPermission.role_id}`
                  : `User: ${editingPermission.user_display_name || editingPermission.user_name || editingPermission.user_id}`
                }
              </p>
            </div>
            <form onSubmit={handleUpdatePermission} className="p-6 space-y-4">
              {/* Permission Checkboxes */}
              <div>
                <label className="block text-sm font-semibold text-white mb-3">
                  Permissions
                </label>
                <div className="space-y-3">
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_view_board"
                      defaultChecked={editingPermission.can_view_board}
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can View Board</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_view_task_list"
                      defaultChecked={editingPermission.can_view_task_list}
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can View Task Titles</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_view_tasks"
                      defaultChecked={editingPermission.can_view_tasks}
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can View Full Tasks</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_edit_tasks"
                      defaultChecked={editingPermission.can_edit_tasks}
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can Edit Tasks</span>
                  </label>
                  <label className="flex items-center gap-2 text-white cursor-pointer">
                    <input
                      type="checkbox"
                      name="can_edit_board"
                      defaultChecked={editingPermission.can_edit_board}
                      className="w-4 h-4 rounded bg-white/5 border-white/10"
                    />
                    <span>Can Edit Board</span>
                  </label>
                </div>
              </div>

              <div className="flex gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => {
                    setShowEditModal(false);
                    setEditingPermission(null);
                  }}
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                >
                  Update
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
      </>
      )}
    </div>
  );
};

export default BoardSettingsPage;
