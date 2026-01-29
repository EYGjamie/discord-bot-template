import type {
  Board,
  BoardPermission,
  BoardLabel,
  Task,
  TaskGroup,
  TaskGroupPermission,
  FilteredTask,
  CreateBoardRequest,
  UpdateBoardRequest,
  CreateTaskRequest,
  UpdateTaskRequest,
  MoveTaskRequest,
  CreateTaskGroupRequest,
  UpdateTaskGroupRequest,
  SetBoardPermissionRequest,
  SetTaskGroupPermissionRequest,
} from '../types/tasks';
import { api } from './api';

// Boards
export const boardsService = {
  getAll: async (): Promise<Board[]> => {
    return await api.get('/api/boards');
  },

  getById: async (id: number): Promise<Board> => {
    return await api.get(`/api/boards/${id}`);
  },

  create: async (data: CreateBoardRequest): Promise<Board> => {
    return await api.post('/api/boards', data);
  },

  update: async (id: number, data: UpdateBoardRequest): Promise<Board> => {
    return await api.put(`/api/boards/${id}`, data);
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/boards/${id}`);
  },

  // Board permissions
  getPermissions: async (boardId: number): Promise<BoardPermission[]> => {
    return await api.get(`/api/boards/${boardId}/permissions`);
  },

  setPermission: async (
    boardId: number,
    data: SetBoardPermissionRequest
  ): Promise<BoardPermission> => {
    return await api.post(`/api/boards/${boardId}/permissions`, data);
  },

  updatePermission: async (
    boardId: number,
    permissionId: number,
    data: Partial<SetBoardPermissionRequest>
  ): Promise<void> => {
    await api.put(`/api/boards/${boardId}/permissions/${permissionId}`, data);
  },

  deletePermission: async (boardId: number, permissionId: number): Promise<void> => {
    await api.delete(`/api/boards/${boardId}/permissions/${permissionId}`);
  },

  // Board labels
  getLabels: async (boardId: number): Promise<BoardLabel[]> => {
    return await api.get(`/api/boards/${boardId}/labels`);
  },

  createLabel: async (boardId: number, data: { name: string; color: string }): Promise<BoardLabel> => {
    return await api.post(`/api/boards/${boardId}/labels`, data);
  },

  updateLabel: async (boardId: number, labelId: number, data: { name: string; color: string }): Promise<BoardLabel> => {
    return await api.put(`/api/boards/${boardId}/labels/${labelId}`, data);
  },

  deleteLabel: async (boardId: number, labelId: number): Promise<void> => {
    await api.delete(`/api/boards/${boardId}/labels/${labelId}`);
  },
};

// Tasks
export const tasksService = {
  getByBoard: async (boardId: number): Promise<FilteredTask[]> => {
    return await api.get(`/api/boards/${boardId}/tasks`);
  },

  getById: async (id: number): Promise<Task> => {
    return await api.get(`/api/tasks/${id}`);
  },

  create: async (data: CreateTaskRequest): Promise<Task> => {
    return await api.post('/api/tasks', data);
  },

  update: async (id: number, data: UpdateTaskRequest): Promise<Task> => {
    return await api.put(`/api/tasks/${id}`, data);
  },

  move: async (id: number, data: MoveTaskRequest): Promise<Task> => {
    return await api.put(`/api/tasks/${id}/move`, data);
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/tasks/${id}`);
  },
};

// Task Groups
export const taskGroupsService = {
  getAll: async (): Promise<TaskGroup[]> => {
    return await api.get('/api/task-groups');
  },

  getById: async (id: number): Promise<TaskGroup> => {
    return await api.get(`/api/task-groups/${id}`);
  },

  create: async (data: CreateTaskGroupRequest): Promise<TaskGroup> => {
    return await api.post('/api/task-groups', data);
  },

  update: async (id: number, data: UpdateTaskGroupRequest): Promise<TaskGroup> => {
    return await api.put(`/api/task-groups/${id}`, data);
  },

  delete: async (id: number): Promise<void> => {
    await api.delete(`/api/task-groups/${id}`);
  },

  // Task group permissions
  getPermissions: async (groupId: number): Promise<TaskGroupPermission[]> => {
    return await api.get(`/api/task-groups/${groupId}/permissions`);
  },

  setPermission: async (
    groupId: number,
    data: SetTaskGroupPermissionRequest
  ): Promise<TaskGroupPermission> => {
    return await api.post(`/api/task-groups/${groupId}/permissions`, data);
  },

  deletePermission: async (groupId: number, permissionId: number): Promise<void> => {
    await api.delete(`/api/task-groups/${groupId}/permissions/${permissionId}`);
  },
};
