import React, { useState, useEffect } from 'react';
import { X, Lock, Trash2, CheckSquare } from 'lucide-react';
import type { FilteredTask, Task, TaskStatus, CreateTaskRequest, UpdateTaskRequest, TaskComment, TaskChecklistItem } from '../../types/tasks';
import { tasksService } from '../../services/tasks';
import { commentsService, checklistService } from '../../services/taskExtras';
import { api } from '../../services/api';
import AssignModal from './AssignModal';
import LabelModal from './LabelModal';
import { getLabelColorFromBoard } from '../../utils/labelColors';

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
    start_date: task?.start_date ? task.start_date.split('T')[0] : '',
    due_date: task?.due_date ? task.due_date.split('T')[0] : '',
    tags: (task && 'tags' in task && Array.isArray(task.tags)) ? task.tags : [],
    group_id: (task && 'group_id' in task) ? task.group_id : undefined,
  });
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showAssignModal, setShowAssignModal] = useState(false);
  const [showLabelModal, setShowLabelModal] = useState(false);
  const [comments, setComments] = useState<TaskComment[]>([]);
  const [newComment, setNewComment] = useState('');
  const [commentLoading, setCommentLoading] = useState(false);
  const [checklist, setChecklist] = useState<TaskChecklistItem[]>([]);
  const [newChecklistItem, setNewChecklistItem] = useState('');
  const [assigneeInfo, setAssigneeInfo] = useState<{ display_name: string; username: string; avatar: string | null } | null>(null);

  // Load comments, checklist, and assignee info when task is loaded
  useEffect(() => {
    if (task?.id) {
      loadComments();
      loadChecklist();
    }
  }, [task?.id]);

  // Load assignee info when assignee_id changes
  useEffect(() => {
    if (formData.assignee_id) {
      loadAssigneeInfo(formData.assignee_id);
    } else {
      setAssigneeInfo(null);
    }
  }, [formData.assignee_id]);

  const loadAssigneeInfo = async (userId: string) => {
    try {
      const data = await api.get(`/api/discord/members/search?q=${userId}`);
      if (Array.isArray(data) && data.length > 0) {
        const user = data[0];
        setAssigneeInfo({
          display_name: user.display_name,
          username: user.username,
          avatar: user.avatar,
        });
      }
    } catch (err) {
      console.error('Failed to load assignee info:', err);
      setAssigneeInfo(null);
    }
  };

  const loadComments = async () => {
    if (!task?.id) {
      console.log('loadComments: No task ID');
      return;
    }
    console.log('Loading comments for task:', task.id);
    try {
      const data = await commentsService.getAll(task.id);
      console.log('Comments loaded:', data);
      setComments(data || []);
    } catch (err: any) {
      console.error('Failed to load comments:', err);
      console.error('Error response:', err.response?.data);
      setComments([]);
    }
  };

  const loadChecklist = async () => {
    if (!task?.id) return;
    try {
      const data = await checklistService.getAll(task.id);
      setChecklist(data || []);
    } catch (err) {
      console.error('Failed to load checklist:', err);
      setChecklist([]);
    }
  };

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
          start_date: formData.start_date || undefined,
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
          start_date: formData.start_date || undefined,
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

  const removeTag = (tag: string) => {
    const tags = formData.tags || [];
    setFormData({ ...formData, tags: tags.filter((t: string) => t !== tag) });
  };

  const handleAssign = (assignees: string[]) => {
    // For now, we only support single assignee in the backend
    // So we'll take the first one or empty string if none selected
    const assignee = (assignees && assignees.length > 0) ? assignees[0] : '';
    setFormData({ ...formData, assignee_id: assignee });
  };

  const handleLabelsUpdate = (labels: string[]) => {
    setFormData({ ...formData, tags: labels });
  };

  const handleAddComment = async () => {
    if (!newComment.trim() || !task?.id) {
      console.log('Cannot add comment: empty text or no task ID');
      return;
    }
    
    console.log('Adding comment - Task ID:', task.id);
    console.log('Adding comment - Text:', newComment);
    console.log('Adding comment - Request body:', { text: newComment });
    setCommentLoading(true);
    setError(null);
    
    try {
      const comment = await commentsService.create(task.id, { text: newComment });
      console.log('Comment created:', comment);
      setNewComment('');
      // Reload comments to get complete data
      await loadComments();
    } catch (err: any) {
      console.error('Failed to add comment:', err);
      console.error('Error response:', err.response?.data);
      setError(err.response?.data?.message || 'Fehler beim Hinzufügen des Kommentars');
    } finally {
      setCommentLoading(false);
    }
  };

  const handleAddChecklistItem = async () => {
    if (!newChecklistItem.trim() || !task?.id) return;
    
    try {
      const item = await checklistService.create(task.id, { text: newChecklistItem });
      setChecklist([...checklist, item]);
      setNewChecklistItem('');
    } catch (err) {
      console.error('Failed to add checklist item:', err);
      setError('Failed to add checklist item');
    }
  };

  const handleToggleChecklistItem = async (itemId: number) => {
    try {
      const updated = await checklistService.toggle(itemId);
      setChecklist(checklist.map(item => item.id === itemId ? updated : item));
    } catch (err) {
      console.error('Failed to toggle checklist item:', err);
    }
  };

  const handleDeleteChecklistItem = async (itemId: number) => {
    try {
      await checklistService.delete(itemId);
      setChecklist(checklist.filter(item => item.id !== itemId));
    } catch (err) {
      console.error('Failed to delete checklist item:', err);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4">
      <div className="bg-[#1e2228] rounded-lg w-full max-w-6xl max-h-[90vh] overflow-y-auto shadow-2xl border border-white/10">
        {/* Main Content */}
        <div className="p-6 relative">
          {/* Close Button */}
          <button
            onClick={onClose}
            className="absolute top-2 right-2 p-2 hover:bg-white/10 rounded-lg transition-colors"
          >
            <X size={20} className="text-white" />
          </button>

          <form onSubmit={handleSubmit}>
            {error && (
              <div className="bg-red-500/10 border border-red-500 text-red-400 px-4 py-3 rounded-lg mb-4">
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
                    Assigned to: <span className="text-white font-medium">{task.assignee_id}</span>
                  </p>
                )}
                {task?.due_date && (
                  <p className="text-gray-400 mt-2">
                    Due: <span className="text-white font-medium">{new Date(task.due_date).toLocaleDateString()}</span>
                  </p>
                )}
              </div>
            )}

            {/* Main Layout: Left Content + Right Sidebars */}
            {(!isEditMode || canReadContent) && (
              <div className="grid grid-cols-1 lg:grid-cols-12 gap-6">
                {/* Left Main Content (6 columns) */}
                <div className="lg:col-span-6 space-y-6">
                  {/* Title */}
                  <div className="flex items-start gap-3">
                    <div className="text-white mt-1">📋</div>
                    <input
                      type="text"
                      value={formData.title}
                      onChange={(e) => setFormData({ ...formData, title: e.target.value })}
                      disabled={!canEdit}
                      required
                      className="flex-1 text-2xl font-bold text-white bg-transparent border-none focus:outline-none focus:ring-0 disabled:opacity-50 -mt-1"
                      placeholder={isEditMode ? 'Task Title' : 'Create New Task'}
                    />
                  </div>

                  {/* Labels Row */}
                  <div className="flex flex-wrap gap-2">
                    {(formData.tags || []).map((tag: string, index: number) => {
                      const currentBoardId = boardId || task?.board_id || 0;
                      const color = getLabelColorFromBoard(currentBoardId, tag);
                      return (
                        <div key={index} className="relative group">
                          <span className={`px-3 py-1 rounded ${color.bg} ${color.text} text-sm font-medium`}>
                            {tag}
                          </span>
                          {canEdit && (
                            <button
                              type="button"
                              onClick={() => removeTag(tag)}
                              className="absolute -top-2 -right-2 w-5 h-5 bg-red-500 text-white rounded-full text-xs opacity-0 group-hover:opacity-100 transition-opacity"
                            >
                              ×
                            </button>
                          )}
                        </div>
                      );
                    })}
                  </div>

                  {/* Action Buttons Row */}
                  <div className="flex flex-wrap gap-2">
                    {canEdit && (
                      <>
                        <button
                          type="button"
                          onClick={() => setShowAssignModal(true)}
                          className="px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded text-sm font-medium flex items-center gap-2 border border-white/10"
                        >
                          👤 Mitglieder
                        </button>
                        <button
                          type="button"
                          onClick={() => setShowLabelModal(true)}
                          className="px-3 py-1.5 bg-white/10 hover:bg-white/20 text-white rounded text-sm font-medium flex items-center gap-2 border border-white/10"
                        >
                          🏷️ Labels
                        </button>
                      </>
                    )}
                  </div>

                  {/* Members & Due Date Display */}
                  <div className="flex items-center gap-4 text-sm">
                    {/* Mitglieder */}
                    <div>
                      <div className="text-gray-400 font-semibold mb-2">Mitglieder</div>
                      <div className="flex items-center gap-2">
                        {formData.assignee_id ? (
                          <div className="flex items-center gap-2">
                            {assigneeInfo ? (
                              <>
                                {assigneeInfo.avatar ? (
                                  <img
                                    src={`https://cdn.discordapp.com/avatars/${formData.assignee_id}/${assigneeInfo.avatar}.png?size=64`}
                                    alt={assigneeInfo.display_name}
                                    className="w-8 h-8 rounded-full"
                                  />
                                ) : (
                                  <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white text-xs font-bold">
                                    {assigneeInfo.display_name.substring(0, 2).toUpperCase()}
                                  </div>
                                )}
                                <span className="text-gray-300">{assigneeInfo.display_name}</span>
                              </>
                            ) : (
                              <>
                                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white text-xs font-bold">
                                  {formData.assignee_id.substring(0, 2).toUpperCase()}
                                </div>
                                <span className="text-gray-300">{formData.assignee_id}</span>
                              </>
                            )}
                          </div>
                        ) : (
                          <span className="text-gray-500">Nicht zugewiesen</span>
                        )}
                        {canEdit && (
                          <button
                            type="button"
                            onClick={() => setShowAssignModal(true)}
                            className="w-8 h-8 rounded-full bg-white/10 hover:bg-white/20 flex items-center justify-center text-gray-400 border border-white/10"
                          >
                            +
                          </button>
                        )}
                      </div>
                    </div>

                    {/* Frist */}
                    {formData.due_date && (
                      <div>
                        <div className="text-gray-400 font-semibold mb-2">Frist</div>
                        <div className="flex items-center gap-2">
                          {(() => {
                            const isDueOverdue = new Date(formData.due_date) < new Date();
                            const isDueSoon = !isDueOverdue && 
                              (new Date(formData.due_date).getTime() - new Date().getTime()) < (24 * 60 * 60 * 1000);
                            return (
                              <span className={`px-2 py-1 rounded text-sm font-medium ${
                                isDueOverdue ? 'bg-red-600 text-white' :
                                isDueSoon ? 'bg-yellow-400 text-gray-900' :
                                'bg-white/10 text-gray-300 border border-white/10'
                              }`}>
                                {new Date(formData.due_date).toLocaleDateString('de-DE', { day: 'numeric', month: 'short' })}
                                {isDueSoon && ' ⚠️ Bald fällig'}
                                {isDueOverdue && ' ⚠️ Überfällig'}
                              </span>
                            );
                          })()}
                        </div>
                      </div>
                    )}
                  </div>

                  {/* Description */}
                  <div>
                    <div className="flex items-center gap-2 mb-3">
                      <span className="text-white">📝</span>
                      <h3 className="text-lg font-semibold text-white">Beschreibung</h3>
                    </div>
                    <textarea
                      value={formData.description}
                      onChange={(e) => setFormData({ ...formData, description: e.target.value })}
                      disabled={!canEdit}
                      rows={4}
                      className="w-full px-4 py-3 bg-[#0d0f15] border border-white/10 rounded-lg text-gray-300 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 resize-none"
                      placeholder="Detaillierte Beschreibung hinzufügen..."
                    />
                  </div>

                  {/* Checklist */}
                  {isEditMode && (
                    <div>
                      <div className="flex items-center justify-between mb-3">
                        <div className="flex items-center gap-2">
                          <span className="text-white">☑️</span>
                          <h3 className="text-lg font-semibold text-white">Checkliste</h3>
                        </div>
                        {checklist && checklist.length > 0 && (
                          <div className="flex items-center gap-2">
                            <span className="text-sm text-gray-400">
                              {Math.round((checklist.filter(item => item.is_completed).length / checklist.length) * 100)}%
                            </span>
                            <div className="w-20 h-2 bg-white/10 rounded-full overflow-hidden">
                              <div 
                                className="h-full bg-green-500 transition-all"
                                style={{ width: `${checklist.length > 0 ? (checklist.filter(item => item.is_completed).length / checklist.length) * 100 : 0}%` }}
                              />
                            </div>
                          </div>
                        )}
                      </div>

                      <div className="space-y-2 mb-3 max-h-64 overflow-y-auto pr-2">
                        {(checklist || []).map((item) => (
                          <div key={item.id} className="flex items-center gap-2 bg-[#0d0f15] border border-white/10 rounded-lg p-3 hover:bg-white/5">
                            <button
                              type="button"
                              onClick={() => handleToggleChecklistItem(item.id)}
                              className="flex-shrink-0"
                            >
                              <CheckSquare 
                                size={18} 
                                className={item.is_completed ? 'text-green-500' : 'text-gray-500'}
                              />
                            </button>
                            <span className={`text-sm flex-1 ${item.is_completed ? 'line-through text-gray-500' : 'text-gray-300'}`}>
                              {item.text}
                            </span>
                            <button
                              type="button"
                              onClick={() => handleDeleteChecklistItem(item.id)}
                              className="text-gray-500 hover:text-red-500 text-sm"
                            >
                              ×
                            </button>
                          </div>
                        ))}
                      </div>

                      <div className="flex gap-2">
                        <input
                          type="text"
                          value={newChecklistItem}
                          onChange={(e) => setNewChecklistItem(e.target.value)}
                          onKeyPress={(e) => e.key === 'Enter' && (e.preventDefault(), handleAddChecklistItem())}
                          placeholder="Element hinzufügen"
                          className="flex-1 px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-gray-300 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
                        />
                        <button
                          type="button"
                          onClick={handleAddChecklistItem}
                          className="px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm font-medium"
                        >
                          Hinzufügen
                        </button>
                      </div>
                    </div>
                  )}

                  {/* Action Buttons */}
                  <div className="flex gap-3 pt-4 border-t border-white/10">
                    {canEdit && (
                      <button
                        type="submit"
                        disabled={loading}
                        className="px-6 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm font-medium disabled:opacity-50"
                      >
                        {loading ? 'Speichern...' : 'Speichern'}
                      </button>
                    )}
                    {isEditMode && canDelete && (
                      <button
                        type="button"
                        onClick={handleDelete}
                        disabled={loading}
                        className="px-6 py-2 bg-red-600 hover:bg-red-700 text-white rounded-lg transition-colors disabled:opacity-50 text-sm font-medium flex items-center gap-2"
                      >
                        <Trash2 size={16} />
                        Löschen
                      </button>
                    )}
                  </div>
                </div>

                {/* Right Sidebars (6 columns total) */}
                <div className="lg:col-span-6 grid grid-cols-7 gap-4">
                  {/* Details Card (Left of right side - 3 columns) */}
                  <div className="bg-[#0d0f15] border border-white/10 rounded-lg p-4 h-fit col-span-7 lg:col-span-3">
                    <h3 className="text-sm font-semibold text-white mb-3">Details</h3>
                    
                    <div className="space-y-3 text-sm">
                      {/* Start Date */}
                      <div>
                        <label className="block text-gray-400 mb-1">Startdatum</label>
                        <input
                          type="date"
                          value={formData.start_date}
                          onChange={(e) => setFormData({ ...formData, start_date: e.target.value })}
                          disabled={!canEdit}
                          className="w-full px-2 py-1 bg-[#1e2228] border border-white/10 rounded text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 text-xs"
                        />
                      </div>

                      {/* Due Date */}
                      <div>
                        <label className="block text-gray-400 mb-1">Frist</label>
                        <input
                          type="date"
                          value={formData.due_date}
                          onChange={(e) => setFormData({ ...formData, due_date: e.target.value })}
                          disabled={!canEdit}
                          className="w-full px-2 py-1 bg-[#1e2228] border border-white/10 rounded text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 text-xs"
                        />
                      </div>

                      {/* Status */}
                      <div>
                        <label className="block text-gray-400 mb-1">Status</label>
                        <select
                          value={formData.status}
                          onChange={(e) => setFormData({ ...formData, status: e.target.value as TaskStatus })}
                          disabled={!canEdit}
                          className="w-full px-2 py-1 bg-[#1e2228] border border-white/10 rounded text-gray-300 focus:outline-none focus:ring-2 focus:ring-blue-500 disabled:opacity-50 text-xs"
                        >
                          <option value="todo" className="bg-[#1e2228]">To Do</option>
                          <option value="in_progress" className="bg-[#1e2228]">In Progress</option>
                          <option value="review" className="bg-[#1e2228]">Review</option>
                          <option value="done" className="bg-[#1e2228]">Done</option>
                        </select>
                      </div>
                    </div>
                  </div>

                  {/* Comments Section (Right of right side - 4 columns) */}
                  {isEditMode && (
                    <div className="bg-[#0d0f15] border border-white/10 rounded-lg p-4 col-span-7 lg:col-span-4">
                      <h3 className="text-sm font-semibold text-white mb-3">💬 Kommentare und Aktivität</h3>
                      
                      {/* Add Comment */}
                      <div className="mb-4">
                        <textarea
                          value={newComment}
                          onChange={(e) => setNewComment(e.target.value)}
                          placeholder="Schreiben Sie einen Kommentar..."
                          rows={3}
                          className="w-full px-3 py-2 bg-[#1e2228] border border-white/10 rounded-lg text-gray-300 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 text-xs resize-none"
                        />
                        <button
                          type="button"
                          onClick={handleAddComment}
                          disabled={commentLoading || !newComment.trim()}
                          className="mt-2 w-full px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-xs font-medium disabled:opacity-50 disabled:cursor-not-allowed"
                        >
                          {commentLoading ? 'Wird gespeichert...' : 'Speichern'}
                        </button>
                      </div>

                      {/* Comments List */}
                      <div className="space-y-4 max-h-96 overflow-y-auto pr-2">
                        {!comments || comments.length === 0 ? (
                          <p className="text-xs text-gray-500 italic">Noch keine Kommentare</p>
                        ) : (
                          (comments || []).filter(comment => comment).map((comment) => (
                            <div key={comment.id} className="flex gap-3 pb-4 border-b border-white/10 last:border-b-0">
                              {/* Avatar */}
                              {comment.user_avatar ? (
                                <img
                                  src={comment.user_avatar}
                                  alt={comment.user_name}
                                  className="w-8 h-8 rounded-full flex-shrink-0"
                                />
                              ) : (
                                <div className="w-8 h-8 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center flex-shrink-0">
                                  <span className="text-sm text-white font-semibold">
                                    {comment.user_name?.charAt(0)?.toUpperCase() || '?'}
                                  </span>
                                </div>
                              )}
                              
                              {/* Comment Content */}
                              <div className="flex-1 min-w-0">
                                <div className="flex items-baseline gap-2 mb-1 flex-wrap">
                                  <span className="text-sm font-semibold text-white">
                                    {comment.user_name || 'Unbekannt'}
                                  </span>
                                  <span className="text-xs text-gray-500">
                                    {new Date(comment.created_at).toLocaleString('de-DE', { 
                                      day: '2-digit',
                                      month: '2-digit',
                                      year: 'numeric',
                                      hour: '2-digit', 
                                      minute: '2-digit' 
                                    })}
                                  </span>
                                </div>
                                <p className="text-sm text-gray-300 break-words whitespace-pre-wrap">
                                  {comment.text}
                                </p>
                              </div>
                            </div>
                          ))
                        )}
                      </div>
                    </div>
                  )}
                </div>
              </div>
            )}
          </form>
        </div>
      </div>

      {/* Assign Modal */}
      {showAssignModal && (
        <AssignModal
          currentAssignees={formData.assignee_id ? [formData.assignee_id] : []}
          onClose={() => setShowAssignModal(false)}
          onAssign={handleAssign}
        />
      )}

      {/* Label Modal */}
      {showLabelModal && (
        <LabelModal
          boardId={boardId || task?.board_id || 0}
          currentLabels={formData.tags || []}
          onClose={() => setShowLabelModal(false)}
          onSave={handleLabelsUpdate}
        />
      )}
    </div>
  );
};

export default TaskModal;
