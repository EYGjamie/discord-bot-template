import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { taskGroupsService } from '../services/tasks';
import type { TaskGroup, TaskGroupPermission, PermissionLevel } from '../types/tasks';
import { ArrowLeft, Plus, Trash2, User, Users } from 'lucide-react';

const GroupPermissionsPage: React.FC = () => {
  const { groupId } = useParams<{ groupId: string }>();
  const navigate = useNavigate();
  const [group, setGroup] = useState<TaskGroup | null>(null);
  const [permissions, setPermissions] = useState<TaskGroupPermission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);

  useEffect(() => {
    if (groupId) {
      loadGroup();
      loadPermissions();
    }
  }, [groupId]);

  const loadGroup = async () => {
    try {
      const data = await taskGroupsService.getById(Number(groupId));
      setGroup(data);
    } catch (err: any) {
      setError(err?.message || 'Failed to load group');
    }
  };

  const loadPermissions = async () => {
    try {
      setLoading(true);
      const data = await taskGroupsService.getPermissions(Number(groupId));
      setPermissions(data || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load permissions');
      setPermissions([]);
    } finally {
      setLoading(false);
    }
  };

  const handleAddPermission = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    const type = formData.get('type') as string;

    try {
      await taskGroupsService.setPermission(Number(groupId), {
        role_id: type === 'role' ? formData.get('role_id') as string : undefined,
        user_id: type === 'user' ? formData.get('user_id') as string : undefined,
        permission: formData.get('permission') as PermissionLevel,
      });
      setShowAddModal(false);
      loadPermissions();
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to add permission');
    }
  };

  const handleDeletePermission = async (permissionId: number) => {
    if (!confirm('Are you sure you want to remove this permission?')) return;

    try {
      await taskGroupsService.deletePermission(Number(groupId), permissionId);
      loadPermissions();
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to delete permission');
    }
  };

  const getPermissionLabel = (level: PermissionLevel) => {
    const labels: Record<PermissionLevel, string> = {
      none: 'No Access',
      existence: 'See Existence',
      read_title: 'Read Title',
      read_content: 'Read Content',
      edit: 'Edit',
      delete: 'Full Access',
    };
    return labels[level];
  };

  const getPermissionColor = (level: PermissionLevel) => {
    const colors: Record<PermissionLevel, string> = {
      none: 'bg-gray-500/20 text-gray-400',
      existence: 'bg-yellow-500/20 text-yellow-400',
      read_title: 'bg-blue-500/20 text-blue-400',
      read_content: 'bg-green-500/20 text-green-400',
      edit: 'bg-purple-500/20 text-purple-400',
      delete: 'bg-red-500/20 text-red-400',
    };
    return colors[level];
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  if (!group) {
    return (
      <div className="container mx-auto px-4 py-8">
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg">
          Group not found
        </div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex items-center justify-between mb-8">
        <div className="flex items-center gap-4">
          <button
            onClick={() => navigate('/tasks/groups')}
            className="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <ArrowLeft size={20} className="text-white" />
          </button>
          <div>
            <h1 className="text-3xl font-bold text-white">{group.name}</h1>
            <p className="text-gray-400 mt-2">Manage permissions for this group</p>
          </div>
        </div>
        <button
          onClick={() => setShowAddModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
        >
          <Plus size={20} />
          Add Permission
        </button>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {/* Permission Levels Info */}
      <div className="bg-white/5 border border-white/10 rounded-2xl p-6 mb-6">
        <h3 className="text-lg font-bold text-white mb-4">Permission Levels</h3>
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-3 text-sm">
          <div className="bg-white/5 rounded-lg p-3">
            <div className="font-semibold text-yellow-400 mb-1">Existenz sehen</div>
            <div className="text-gray-400 text-xs">See assignee and due date only</div>
          </div>
          <div className="bg-white/5 rounded-lg p-3">
            <div className="font-semibold text-blue-400 mb-1">Titel Lesen</div>
            <div className="text-gray-400 text-xs">+ Task title</div>
          </div>
          <div className="bg-white/5 rounded-lg p-3">
            <div className="font-semibold text-green-400 mb-1">Inhalt lesen</div>
            <div className="text-gray-400 text-xs">+ Task details</div>
          </div>
          <div className="bg-white/5 rounded-lg p-3">
            <div className="font-semibold text-purple-400 mb-1">Bearbeiten</div>
            <div className="text-gray-400 text-xs">+ Edit and change status</div>
          </div>
          <div className="bg-white/5 rounded-lg p-3">
            <div className="font-semibold text-red-400 mb-1">Löschen</div>
            <div className="text-gray-400 text-xs">+ Delete tasks</div>
          </div>
        </div>
      </div>

      {/* Permissions List */}
      <div className="bg-white/5 border border-white/10 rounded-2xl overflow-hidden">
        <table className="w-full">
          <thead className="bg-white/5">
            <tr>
              <th className="text-left px-6 py-3 text-sm font-semibold text-gray-400">Type</th>
              <th className="text-left px-6 py-3 text-sm font-semibold text-gray-400">ID</th>
              <th className="text-left px-6 py-3 text-sm font-semibold text-gray-400">Permission</th>
              <th className="text-right px-6 py-3 text-sm font-semibold text-gray-400">Actions</th>
            </tr>
          </thead>
          <tbody className="divide-y divide-white/10">
            {permissions.length === 0 ? (
              <tr>
                <td colSpan={4} className="px-6 py-8 text-center text-gray-400">
                  No permissions set. Add one to get started.
                </td>
              </tr>
            ) : (
              permissions.map((perm) => (
                <tr key={perm.id} className="hover:bg-white/5">
                  <td className="px-6 py-4">
                    <div className="flex items-center gap-2">
                      {perm.role_id ? (
                        <>
                          <Users size={16} className="text-blue-400" />
                          <span className="text-white">Role</span>
                        </>
                      ) : (
                        <>
                          <User size={16} className="text-green-400" />
                          <span className="text-white">User</span>
                        </>
                      )}
                    </div>
                  </td>
                  <td className="px-6 py-4">
                    <code className="text-gray-400 text-sm">
                      {perm.role_id || perm.user_id}
                    </code>
                  </td>
                  <td className="px-6 py-4">
                    <span
                      className={`px-3 py-1 rounded-full text-xs font-semibold ${getPermissionColor(perm.permission)}`}
                    >
                      {getPermissionLabel(perm.permission)}
                    </span>
                  </td>
                  <td className="px-6 py-4 text-right">
                    <button
                      onClick={() => handleDeletePermission(perm.id)}
                      className="p-2 hover:bg-red-500/20 text-red-400 rounded-lg transition-colors"
                    >
                      <Trash2 size={16} />
                    </button>
                  </td>
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* Add Permission Modal */}
      {showAddModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md">
            <div className="p-6 border-b border-white/10">
              <h2 className="text-xl font-bold text-white">Add Permission</h2>
            </div>
            <form onSubmit={handleAddPermission} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Type *
                </label>
                <select
                  name="type"
                  required
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                  onChange={(e) => {
                    const form = e.currentTarget.form;
                    if (form) {
                      const roleInput = form.querySelector('[name="role_id"]') as HTMLInputElement;
                      const userInput = form.querySelector('[name="user_id"]') as HTMLInputElement;
                      if (e.target.value === 'role') {
                        roleInput.disabled = false;
                        userInput.disabled = true;
                        userInput.value = '';
                      } else {
                        roleInput.disabled = true;
                        userInput.disabled = false;
                        roleInput.value = '';
                      }
                    }
                  }}
                >
                  <option value="role">Discord Role</option>
                  <option value="user">Discord User</option>
                </select>
              </div>
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Role ID
                </label>
                <input
                  type="text"
                  name="role_id"
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  placeholder="Enter Discord role ID"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  User ID
                </label>
                <input
                  type="text"
                  name="user_id"
                  disabled
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  placeholder="Enter Discord user ID"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Permission Level *
                </label>
                <select
                  name="permission"
                  required
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500"
                >
                  <option value="existence">Existenz sehen (See assignee & due date)</option>
                  <option value="read_title">Titel Lesen (+ Title)</option>
                  <option value="read_content">Inhalt lesen (+ Details)</option>
                  <option value="edit">Bearbeiten (+ Edit)</option>
                  <option value="delete">Löschen (+ Delete)</option>
                </select>
              </div>
              <div className="flex gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => setShowAddModal(false)}
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                >
                  Add
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default GroupPermissionsPage;
