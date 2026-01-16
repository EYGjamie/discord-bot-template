import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { ArrowLeft, Plus, Trash2, Shield, User, Users } from 'lucide-react';
import { boardsService } from '../services/tasks';
import type { Board, BoardPermission } from '../types/tasks';

const BoardSettingsPage: React.FC = () => {
  const { boardId } = useParams<{ boardId: string }>();
  const navigate = useNavigate();
  const [board, setBoard] = useState<Board | null>(null);
  const [permissions, setPermissions] = useState<BoardPermission[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showAddModal, setShowAddModal] = useState(false);

  useEffect(() => {
    if (boardId) {
      loadBoard();
      loadPermissions();
    }
  }, [boardId]);

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

  const handleAddPermission = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    
    const roleId = formData.get('role_id') as string;
    const userId = formData.get('user_id') as string;
    const canView = formData.get('can_view') === 'on';
    const canCreate = formData.get('can_create') === 'on';

    try {
      await boardsService.setPermission(Number(boardId), {
        role_id: roleId || undefined,
        user_id: userId || undefined,
        can_view: canView,
        can_create: canCreate,
      });
      setShowAddModal(false);
      loadPermissions();
    } catch (err: any) {
      setError(err?.message || 'Failed to add permission');
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
              <div className="flex items-center gap-4">
                {permission.role_id ? (
                  <div className="flex items-center gap-2">
                    <Users className="text-purple-400" size={20} />
                    <div>
                      <div className="text-white font-semibold">Role</div>
                      <div className="text-sm text-gray-400">ID: {permission.role_id}</div>
                    </div>
                  </div>
                ) : (
                  <div className="flex items-center gap-2">
                    <User className="text-blue-400" size={20} />
                    <div>
                      <div className="text-white font-semibold">User</div>
                      <div className="text-sm text-gray-400">ID: {permission.user_id}</div>
                    </div>
                  </div>
                )}
                <div className="flex gap-2 ml-8">
                  {permission.can_view && (
                    <span className="px-2 py-1 bg-green-500/20 text-green-400 rounded text-xs font-semibold">
                      Can View
                    </span>
                  )}
                  {permission.can_create && (
                    <span className="px-2 py-1 bg-blue-500/20 text-blue-400 rounded text-xs font-semibold">
                      Can Create
                    </span>
                  )}
                </div>
              </div>
              <button
                onClick={() => handleDeletePermission(permission.id)}
                className="p-2 hover:bg-red-500/20 text-red-400 rounded-lg transition-colors"
              >
                <Trash2 size={18} />
              </button>
            </div>
          ))
        )}
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
                  Role ID (optional)
                </label>
                <input
                  type="text"
                  name="role_id"
                  placeholder="Discord Role ID"
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-500 mt-1">Leave empty to set user permission</p>
              </div>
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  User ID (optional)
                </label>
                <input
                  type="text"
                  name="user_id"
                  placeholder="Discord User ID"
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <p className="text-xs text-gray-500 mt-1">Leave empty to set role permission</p>
              </div>
              <div className="space-y-2">
                <label className="flex items-center gap-2 text-white cursor-pointer">
                  <input
                    type="checkbox"
                    name="can_view"
                    defaultChecked
                    className="w-4 h-4 rounded bg-white/5 border-white/10"
                  />
                  <span>Can View</span>
                </label>
                <label className="flex items-center gap-2 text-white cursor-pointer">
                  <input
                    type="checkbox"
                    name="can_create"
                    className="w-4 h-4 rounded bg-white/5 border-white/10"
                  />
                  <span>Can Create Tasks</span>
                </label>
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
                  Add Permission
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
