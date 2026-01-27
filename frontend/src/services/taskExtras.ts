import { api } from './api';
import type { TaskComment, TaskChecklistItem, CreateCommentRequest, CreateChecklistItemRequest, UpdateChecklistItemRequest } from '../types/tasks';

export const commentsService = {
  getAll: async (taskId: number): Promise<TaskComment[]> => {
    return await api.get(`/api/tasks/${taskId}/comments`);
  },

  create: async (taskId: number, data: CreateCommentRequest): Promise<TaskComment> => {
    return await api.post(`/api/tasks/${taskId}/comments`, data);
  },

  update: async (commentId: number, data: CreateCommentRequest): Promise<TaskComment> => {
    return await api.put(`/api/comments/${commentId}`, data);
  },

  delete: async (commentId: number): Promise<void> => {
    await api.delete(`/api/comments/${commentId}`);
  },
};

export const checklistService = {
  getAll: async (taskId: number): Promise<TaskChecklistItem[]> => {
    return await api.get(`/api/tasks/${taskId}/checklist`);
  },

  create: async (taskId: number, data: CreateChecklistItemRequest): Promise<TaskChecklistItem> => {
    return await api.post(`/api/tasks/${taskId}/checklist`, data);
  },

  update: async (itemId: number, data: UpdateChecklistItemRequest): Promise<TaskChecklistItem> => {
    return await api.put(`/api/checklist/${itemId}`, data);
  },

  toggle: async (itemId: number): Promise<TaskChecklistItem> => {
    return await api.post(`/api/checklist/${itemId}/toggle`, {});
  },

  delete: async (itemId: number): Promise<void> => {
    await api.delete(`/api/checklist/${itemId}`);
  },
};
