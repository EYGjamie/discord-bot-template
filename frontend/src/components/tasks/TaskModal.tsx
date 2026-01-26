import React, { useState } from 'react';
import { X, Lock, Trash2 } from 'lucide-react';
import type { FilteredTask, Task, TaskStatus, CreateTaskRequest, UpdateTaskRequest } from '../../types/tasks';
import { tasksService } from '../../services/tasks';
import AssignModal from './AssignModal';
import { useAuth } from '../../hooks/useAuth';

interface TaskModalProps {
  task?: FilteredTask | Task;
  boardId?: number;
  onClose: () => void;
  onUpdate: () => void;
}

const TaskModal: React.FC<TaskModalProps> = ({ task, boardId, onClose, onUpdate }) => {
  const { user } = useAuth();
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
    tags: (task && 'tags' in task && Array.isArray(task.tags)) ? task.tags : [],
    group_id: (task && 'group_id' in task) ? task.group_id : undefined,
  });
  const [tagInput, setTagInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [comments, setComments] = useState<Array<{id: number, userId: string, userName: string, userAvatar: string, text: string, date: string}>>([]);
  const [newComment, setNewComment] = useState('');

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
    const tags = formData.tags || [];
    if (tagInput.trim() && !tags.includes(tagInput.trim())) {
      setFormData({ ...formData, tags: [...tags, tagInput.trim()] });
      setTagInput('');
    }
  };

  const removeTag = (tag: string) => {
    const tags = formData.tags || [];
    setFormData({ ...formData, tags: tags.filter((t: string) => t !== tag) });
  };

  const handleAssign = (assignees: string[]) => {
    // For now, we only support single assignee in the backend
    // So we'll take the first one or empty string if none selected
    const assignee = assignees.length > 0 ? assignees[0] : '';
    setFormData({ ...formData, assignee_id: assignee });
  };

  const handleAddComment = () => {
    if (newComment.trim() && user) {
      const comment = {
        id: Date.now(),
        userId: user.discord_id,
        userName: user.display_name || user.username,
        userAvatar: user.avatar_url || '',
        text: newComment,
        date: new Date().toISOString()
      };
      setComments([...comments, comment]);
      setNewComment('');
      // TODO: Save to backend API
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4">
      <div className="bg-[#1a1d29] border border-white/10 rounded-2xl w-full max-w-4xl max-h-[90vh] overflow-y-auto">
        {/* Header */}
        <div className="flex items-center justify-between p-6 border-b border-white/10">
          <div className="flex flex-col gap-2">
            <input
              type="text"
              value={formData.title}
              onChange={(e) => setFormData({ ...formData, title: e.target.value })}
              disabled={!canEdit}
              required
              className="text-xl font-bold text-white bg-transparent border-none focus:outline-none focus:ring-0 disabled:opacity-50"
              placeholder={isEditMode ? 'Task Title' : 'Create New Task'}
            />
            {task?.group_id && (
              <span className="text-sm text-blue-400 px-3 py-1 bg-blue-500/20 rounded-md w-fit">
                Team • Valorant
              </span>
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
        <form onSubmit={handleSubmit} className="p-6">
          {error && (
            <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-4">
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

          {/* Two Column Layout */}
          {(!isEditMode || canReadContent) && (
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-6">
              {/* Left Column - Description & Tags (2 columns) */}
              <div className="lg:col-span-2 space-y-4">
                <div className="bg-[#0d0f15] border border-white/10 rounded-lg p-4">
                  <h3 className="text-xs font-bold text-gray-400 uppercase mb-3 tracking-wider">
                    Description
                  </h3>
                  <textarea
                    value={formData.description}
                    onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                    disabled={!canEdit}
                    rows={6}
                    className="w-full px-0 py-0 bg-transparent border-none text-gray-300 text-sm placeholder-gray-500 focus:outline-none focus:ring-0 disabled:opacity-50 resize-none"
                    placeholder="• Scrims anfragen (2–3 Gegner)\n• Timeslots bestätigen\n• Post im Discord planen\n\nZiel: Plan bis Mittwoch fix haben."
                  />
                </div>

                {/* Tags */}
                <div>
                  <label className="block text-xs font-bold text-gray-400 uppercase mb-2 tracking-wider">
                    Tags
                  </label>
                  <div className="flex gap-2 mb-2">
                    <input
                      type="text"
                      value={tagInput}
                      onChange={(e) => setTagInput(e.target.value)}
                      onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), addTag())}
                      disabled={!canEdit}
                      className="flex-1 px-3 py-1.5 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 text-xs"
                      placeholder="Add a tag"
                    />
                    <button
                      type="button"
                      onClick={addTag}
                      disabled={!canEdit}
                      className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 text-xs"
                    >
                      Add
                    </button>
                  </div>
                  <div className="flex flex-wrap gap-1.5">
                    {(formData.tags || []).map((tag: string, index: number) => (
                      <span
                        key={index}
                        className="px-2 py-0.5 bg-blue-500/20 text-blue-400 rounded-full text-xs font-semibold flex items-center gap-1"
                      >
                        {tag}
                        {canEdit && (
                          <button
                            type="button"
                            onClick={() => removeTag(tag)}
                            className="hover:text-red-400 text-xs"
                          >
                            ×
                          </button>
                        )}
                      </span>
                    ))}
                  </div>
                </div>
              </div>

              {/* Right Column - Details & Comments (1 column) */}
              <div className="lg:col-span-1 space-y-4">
                {/* Details Section */}
                <div className="bg-[#0d0f15] border border-white/10 rounded-lg p-4 space-y-4">
                  <h3 className="text-xs font-bold text-gray-400 uppercase tracking-wider">
                    Details
                  </h3>

                  {/* Assignee */}
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">
                      Assignee
                    </label>
                    <div className="text-right text-white text-sm">
                      {formData.assignee_id || 'Unassigned'}
                    </div>
                  </div>

                  {/* Due Date */}
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">
                      Due
                    </label>
                    <input
                      type="date"
                      value={formData.due_date}
                      onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                      disabled={!canEdit}
                      className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 text-sm text-right"
                    />
                  </div>

                  {/* Status */}
                  <div>
                    <label className="block text-sm text-gray-400 mb-2">
                      Status
                    </label>
                    <select
                      value={formData.status}
                      onChange={(e) => setFormData({ ...formData, status: e.target.value as TaskStatus })}
                      disabled={!canEdit}
                      className="w-full px-3 py-2 bg-[#1a1d29] border border-white/10 rounded-lg text-white focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 text-sm"
                    >
                      <option value="todo" className="bg-[#1a1d29] text-white">To Do</option>
                      <option value="in_progress" className="bg-[#1a1d29] text-white">In Progress</option>
                      <option value="review" className="bg-[#1a1d29] text-white">Review</option>
                      <option value="done" className="bg-[#1a1d29] text-white">Done</option>
                    </select>
                  </div>
                </div>

                {/* Comments Section */}
                <div className="bg-[#0d0f15] border border-white/10 rounded-lg p-4">
                  <h3 className="text-xs font-bold text-gray-400 uppercase mb-3 tracking-wider">
                    Comments
                  </h3>
                  
                  {/* Comments List */}
                  <div className="space-y-3 mb-4 max-h-64 overflow-y-auto">
                    {comments.length === 0 ? (
                      <p className="text-xs text-gray-500 italic">No comments yet</p>
                    ) : (
                      comments.map((comment) => (
                        <div key={comment.id} className="bg-white/5 rounded-lg p-3">
                          <div className="flex items-center gap-2 mb-2">
                            {comment.userAvatar ? (
                              <img
                                src={comment.userAvatar}
                                alt={comment.userName}
                                className="w-6 h-6 rounded-full"
                              />
                            ) : (
                              <div className="w-6 h-6 rounded-full bg-gray-700 flex items-center justify-center">
                                <span className="text-xs text-gray-400">{comment.userName.charAt(0)}</span>
                              </div>
                            )}
                            <div className="flex-1 flex items-center justify-between">
                              <span className="text-xs font-semibold text-white">{comment.userName}</span>
                              <span className="text-xs text-gray-500">
                                {new Date(comment.date).toLocaleString(undefined, { 
                                  month: 'short', 
                                  day: 'numeric', 
                                  hour: '2-digit', 
                                  minute: '2-digit' 
                                })}
                              </span>
                            </div>
                          </div>
                          <p className="text-xs text-gray-300 pl-8">{comment.text}</p>
                        </div>
                      ))
                    )}
                  </div>

                  {/* Add Comment */}
                  <div className="space-y-2">
                    <textarea
                      value={newComment}
                      onChange={(e) => setNewComment(e.target.value)}
                      placeholder="Add a comment..."
                      rows={2}
                      className="w-full px-3 py-2 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 text-xs resize-none"
                    />
                    <button
                      type="button"
                      onClick={handleAddComment}
                      className="w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-xs flex items-center justify-center gap-1"
                    >
                      💬 Comment
                    </button>
                  </div>
                </div>
              </div>
            </div>
          )}

          {/* Footer */}
          <div className="flex flex-col gap-4 pt-6 border-t border-white/10">
            {/* Action Buttons */}
            <div className="flex flex-wrap gap-2">
              {canEdit && (
                <>
                  <button
                    type="submit"
                    disabled={loading}
                    className="px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-lg transition-colors text-sm flex items-center gap-2 disabled:opacity-50"
                  >
                    <span>✏️</span>
                    {loading ? 'Saving...' : 'Save'}
                  </button>
                  
                  <button
                    type="button"
                    onClick={() => setShowAssignModal(true)}
                    className="px-4 py-2 bg-white/5 hover:bg-white/10 border border-white/10 text-white rounded-lg transition-colors text-sm flex items-center gap-2"
                  >
                    <span>👤</span>
                    Assign
                  </button>
                </>
              )}

              {isEditMode && canDelete && (
                <button
                  type="button"
                  onClick={handleDelete}
                  disabled={loading}
                  className="ml-auto px-4 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50 text-sm flex items-center gap-2"
                >
                  <Trash2 size={16} />
                  Delete
                </button>
              )}
            </div>
          </div>
        </form>
      </div>

      {/* Assign Modal */}
      {showAssignModal && (
        <AssignModal
          currentAssignees={formData.assignee_id ? [formData.assignee_id] : []}
          onClose={() => setShowAssignModal(false)}
          onAssign={handleAssign}
        />
      )}
    </div>
  );
};

export default TaskModal;
