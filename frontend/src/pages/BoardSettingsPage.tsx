import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Plus, Trash2, Shield, User, Users, Search, Edit } from 'lucide-react';
import { boardsService } from '../services/tasks';
import { api } from '../services/api';
import type { Board, BoardPermission } from '../types/tasks';
import type { Role, MemberInfo, RolesAndMembersResponse } from '../types/discord';

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
          <h1 className="text-3xl font-bold text-white">{board.name} - Settings</h1>
          <p className="text-gray-400 mt-1">Manage board permissions</p>
        </div>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {/* Info Box */}
      <div className="bg-blue-500/10 border border-blue-500/20 rounded-xl p-4 mb-6">
        <div className="flex items-start gap-3">
          <Shield className="text-blue-400 mt-1" size={20} />
          <div>
            <h3 className="text-white font-semibold mb-1">Permission System</h3>
            <p className="text-sm text-gray-400">
              If no permissions are set, all authenticated users have full access to this board.
              Once you add permissions, only users/roles with explicit permissions can access the board.
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
                    <User className="text-blue-400" size={24} />
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
                            className="w-full px-4 py-2.5 text-left hover:bg-blue-600/20 hover:border-l-2 hover:border-blue-500 transition-all text-white first:rounded-t-lg last:rounded-b-lg"
                          >
                            <div className="font-semibold">{member.display_name}</div>
                            <div className="text-xs text-gray-400">{member.username}</div>
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                  {selectedMember && (
                    <div className="mt-2 p-2 bg-blue-500/20 border border-blue-500/30 rounded-lg text-sm text-blue-300">
                      Selected: {selectedMember.display_name}
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
    </div>
  );
};

export default BoardSettingsPage;
