import React, { useState } from 'react';
import { X, Lock, Trash2, Save } from 'lucide-react';
import type { FilteredTask, Task, TaskStatus, CreateTaskRequest, UpdateTaskRequest } from '../../types/tasks';
import { tasksService } from '../../services/tasks';

interface TaskModalProps {
  task?: FilteredTask | Task;
  boardId?: number;
  onClose: () => void;
  onUpdate: () => void;
}

const TaskModal: React.FC<TaskModalProps> = ({ task, boardId, onClose, onUpdate }) => {
  const isEditMode = !!task;
  const canReadContent = !task || (task.permission && ['read_content', 'edit', 'delete'].includes(task.permission));
  const canEdit = !task || (task.permission && ['edit', 'delete'].includes(task.permission));
  const canDelete = !task || task.permission === 'delete';

  const [formData, setFormData] = useState({
    title: (task && 'title' in task && task.title) ? task.title : '',
    description: (task && 'description' in task && task.description) ? task.description : '',
    status: task?.status || 'todo' as TaskStatus,
    assignee_id: task?.assignee_id || '',
    due_date: task?.due_date ? task.due_date.split('T')[0] : '',
    tags: (task && 'tags' in task && task.tags) ? JSON.parse(task.tags) : [],
    group_id: (task && 'group_id' in task) ? task.group_id : undefined,
  });
  const [tagInput, setTagInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canEdit && isEditMode) return;

    setLoading(true);
    setError(null);

    try {
      if (isEditMode && task) {
        const updateData: UpdateTaskRequest = {
          title: formData.title,
          description: formData.description,
          status: formData.status,
          assignee_id: formData.assignee_id || undefined,
          due_date: formData.due_date || undefined,
          tags: formData.tags,
        };
        await tasksService.update(task.id, updateData);
      } else if (boardId) {
        const createData: CreateTaskRequest = {
          board_id: boardId,
          group_id: formData.group_id,
          title: formData.title,
          description: formData.description,
          status: formData.status,
          assignee_id: formData.assignee_id || undefined,
          due_date: formData.due_date || undefined,
          tags: formData.tags,
        };
        await tasksService.create(createData);
      }
      onUpdate();
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to save task');
    } finally {
      setLoading(false);
    }
  };

  const handleDelete = async () => {
    if (!task || !canDelete) return;
    if (!confirm('Are you sure you want to delete this task?')) return;

    setLoading(true);
    try {
      await tasksService.delete(task.id);
      onUpdate();
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to delete task');
      setLoading(false);
    }
  };

  const addTag = () => {
    if (tagInput.trim() && !formData.tags.includes(tagInput.trim())) {
      setFormData({ ...formData, tags: [...formData.tags, tagInput.trim()] });
      setTagInput('');
    }
  };

  const removeTag = (tag: string) => {
    setFormData({ ...formData, tags: formData.tags.filter((t: string) => t !== tag) });
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-gray-900 border border-white/10 rounded-2xl w-full max-w-3xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-white/10">
          <div className="flex items-center gap-3">
            <h2 className="text-xl font-bold text-white">
              {isEditMode ? 'Task Details' : 'Create New Task'}
            </h2>
            {isEditMode && !canReadContent && (
              <div className="flex items-center gap-2 px-3 py-1 bg-yellow-500/20 text-yellow-500 rounded-full text-sm">
                <Lock size={14} />
                <span>Restricted Access</span>
              </div>
            )}
          </div>
          <button
            onClick={onClose}
            className="p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <X size={20} className="text-white" />
          </button>
        </div>

        {/* Body */}
        <form onSubmit={handleSubmit} className="p-6 space-y-6">
          {error && (
            <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg">
              {error}
            </div>
          )}

          {/* Restricted View */}
          {isEditMode && !canReadContent && (
            <div className="bg-white/5 border border-white/10 rounded-lg p-6 text-center">
              <Lock className="mx-auto mb-4 text-gray-400" size={48} />
              <h3 className="text-lg font-semibold text-white mb-2">Limited Access</h3>
              <p className="text-gray-400">
                You can see this task exists, but you don't have permission to view its details.
              </p>
              {task?.assignee_id && (
                <p className="text-gray-400 mt-4">
                  Assigned to: <span className="text-white">{task.assignee_id}</span>
                </p>
              )}
              {task?.due_date && (
                <p className="text-gray-400 mt-2">
                  Due: <span className="text-white">{new Date(task.due_date).toLocaleDateString()}</span>
                </p>
              )}
            </div>
          )}

          {/* Full Form */}
          {(!isEditMode || canReadContent) && (
            <>
              {/* Title */}
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Title *
                </label>
                <input
                  type="text"
                  value={formData.title}
                  onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                  disabled={!canEdit}
                  required
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  placeholder="Enter task title"
                />
              </div>

              {/* Description */}
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Description
                </label>
                <textarea
                  value={formData.description}
                  onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                  disabled={!canEdit}
                  rows={4}
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  placeholder="Enter task description"
                />
              </div>

              {/* Status and Due Date */}
              <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                <div>
                  <label className="block text-sm font-semibold text-white mb-2">
                    Status
                  </label>
                  <select
                    value={formData.status}
                    onChange={(e) => setFormData({ ...formData, status: e.target.value as TaskStatus })}
                    disabled={!canEdit}
                    className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  >
                    <option value="todo">To Do</option>
                    <option value="in_progress">In Progress</option>
                    <option value="review">Review Pending</option>
                    <option value="done">Done</option>
                  </select>
                </div>

                <div>
                  <label className="block text-sm font-semibold text-white mb-2">
                    Due Date
                  </label>
                  <input
                    type="date"
                    value={formData.due_date}
                    onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                    disabled={!canEdit}
                    className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  />
                </div>
              </div>

              {/* Assignee */}
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Assignee (Discord User ID)
                </label>
                <input
                  type="text"
                  value={formData.assignee_id}
                  onChange={(e) => setFormData({ ...formData, assignee_id: e.target.value })}
                  disabled={!canEdit}
                  className="w-full px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                  placeholder="Enter user ID"
                />
              </div>

              {/* Tags */}
              <div>
                <label className="block text-sm font-semibold text-white mb-2">
                  Tags
                </label>
                <div className="flex gap-2 mb-2">
                  <input
                    type="text"
                    value={tagInput}
                    onChange={(e) => setTagInput(e.target.value)}
                    onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
                    disabled={!canEdit}
                    className="flex-1 px-4 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50"
                    placeholder="Add a tag"
                  />
                  <button
                    type="button"
                    onClick={addTag}
                    disabled={!canEdit}
                    className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50"
                  >
                    Add
                  </button>
                </div>
                <div className="flex flex-wrap gap-2">
                  {formData.tags.map((tag: string, index: number) => (
                    <span
                      key={index}
                      className="px-3 py-1 bg-blue-500/20 text-blue-400 rounded-full text-sm font-semibold flex items-center gap-2"
                    >
                      {tag}
                      {canEdit && (
                        <button
                          type="button"
                          onClick={() => removeTag(tag)}
                          className="hover:text-red-400"
                        >
                          ×
                        </button>
                      )}
                    </span>
                  ))}
                </div>
              </div>
            </>
          )}

          {/* Footer */}
          <div className="flex items-center justify-between pt-6 border-t border-white/10">
            <div>
              {isEditMode && canDelete && (
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={loading}
                  className="flex items-center gap-2 px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50"
                >
                  <Trash2 size={16} />
                  Delete
                </button>
              )}
            </div>
            <div className="flex gap-2">
              <button
                type="button"
                onClick={onClose}
                className="px-4 py-2 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
              >
                Cancel
              </button>
              {canEdit && (
                <button
                  type="submit"
                  disabled={loading}
                  className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50"
                >
                  <Save size={16} />
                  {loading ? 'Saving...' : isEditMode ? 'Save Changes' : 'Create Task'}
                </button>
              )}
            </div>
          </div>
        </form>
      </div>
    </div>
  );
};

export default TaskModal;
