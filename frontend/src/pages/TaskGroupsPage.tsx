import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { taskGroupsService } from '../services/tasks';
import type { TaskGroup } from '../types/tasks';
import { Plus, Shield, Settings } from 'lucide-react';

const TaskGroupsPage: React.FC = () => {
  const navigate = useNavigate();
  const [groups, setGroups] = useState<TaskGroup[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [showCreateModal, setShowCreateModal] = useState(false);

  useEffect(() => {
    loadGroups();
  }, []);

  const loadGroups = async () => {
    try {
      setLoading(true);
      const data = await taskGroupsService.getAll();
      setGroups(data || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load groups');
      setGroups([]);
    } finally {
      setLoading(false);
    }
  };

  const handleCreateGroup = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    const formData = new FormData(e.currentTarget);
    
    try {
      await taskGroupsService.create({
        name: formData.get('name') as string,
        description: formData.get('description') as string,
        color: formData.get('color') as string || '#39d98a',
      });
      setShowCreateModal(false);
      loadGroups();
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to create group');
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white">Task Groups</h1>
          <p className="text-gray-400 mt-2">
            Manage permission groups for tasks
          </p>
        </div>
        <button
          onClick={() => setShowCreateModal(true)}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
        >
          <Plus size={20} />
          New Group
        </button>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
        {groups.map((group) => (
          <div
            key={group.id}
            className="bg-white/5 border border-white/10 rounded-2xl p-6 hover:bg-white/10 transition-all"
          >
            <div className="flex items-start justify-between mb-4">
              <div
                className="w-12 h-12 rounded-xl flex items-center justify-center"
                style={{ backgroundColor: group.color + '20', color: group.color }}
              >
                <Shield size={24} />
              </div>
              <button
                onClick={() => navigate(`/tasks/groups/${group.id}/permissions`)}
                className="p-2 hover:bg-white/10 rounded-lg transition-colors"
              >
                <Settings size={18} className="text-white" />
              </button>
            </div>
            <h3 className="text-lg font-bold text-white mb-2">{group.name}</h3>
            <p className="text-gray-400 text-sm line-clamp-2">
              {group.description || 'No description'}
            </p>
            <div className="mt-4 pt-4 border-t border-white/10">
              <button
                onClick={() => navigate(`/tasks/groups/${group.id}/permissions`)}
                className="text-blue-400 hover:text-blue-300 text-sm font-semibold transition-colors"
              >
                Manage Permissions →
              </button>
            </div>
          </div>
        ))}
      </div>

      {/* Create Group Modal */}
      {showCreateModal && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
          <div className="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-md">
            <div className="p-6 border-b border-white/10">
              <h2 className="text-xl font-bold text-white">Create Task Group</h2>
            </div>
            <form onSubmit={handleCreateGroup} className="p-6 space-y-4">
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Name *
                </label>
                <input
                  type="text"
                  name="name"
                  required
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="e.g., Project Management"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Description
                </label>
                <textarea
                  name="description"
                  rows={3}
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500"
                  placeholder="Describe this group"
                />
              </div>
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Color
                </label>
                <input
                  type="color"
                  name="color"
                  defaultValue="#39d98a"
                  className="w-full h-12 bg-white/5 border border-white/10 rounded-lg cursor-pointer"
                />
              </div>
              <div className="flex gap-2 pt-4">
                <button
                  type="button"
                  onClick={() => setShowCreateModal(false)}
                  className="flex-1 px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
                >
                  Cancel
                </button>
                <button
                  type="submit"
                  className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
                >
                  Create
                </button>
              </div>
            </form>
          </div>
        </div>
      )}
    </div>
  );
};

export default TaskGroupsPage;
