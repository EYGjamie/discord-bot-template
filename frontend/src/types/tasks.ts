export type TaskStatus = 'todo' | 'in_progress' | 'review' | 'done';

export type PermissionLevel = 
  | 'none'
  | 'existence'    // See assignee and due date only
  | 'read_title'   // + task title
  | 'read_content' // + task details
  | 'edit'         // + edit details and change status
  | 'delete';      // + delete tasks

export interface Board {
  id: number;
  guild_id: string;
  name: string;
  description: string;
  color: string;
  position: number;
  created_by: string;
  created_at: string;
  updated_at: string;
}

export interface BoardPermission {
  id: number;
  board_id: number;
  role_id?: string;
  user_id?: string;
  role_name?: string;
  user_name?: string;
  user_display_name?: string;
  can_view_board: boolean;
  can_view_task_list: boolean;
  can_view_tasks: boolean;
  can_edit_tasks: boolean;
  can_edit_board: boolean;
  created_at: string;
}

export interface Task {
  id: number;
  board_id: number;
  group_id?: number;
  title: string;
  description: string;
  status: TaskStatus;
  position: number;
  assignee_id?: string;
  start_date?: string;
  due_date?: string;
  tags: string[]; // Array of tags
  created_by: string;
  created_at: string;
  updated_at: string;
  permission?: PermissionLevel; // Added by permission checker
}

export interface TaskWithTags extends Omit<Task, 'tags'> {
  tags: string[];
}

export interface FilteredTask {
  id: number;
  board_id: number;
  group_id?: number;
  title?: string; // Only if permission >= read_title
  description?: string; // Only if permission >= read_content
  status: TaskStatus;
  position: number;
  assignee_id?: string;
  start_date?: string;
  due_date?: string;
  tags?: string[]; // Only if permission >= read_content
  created_by?: string; // Only if permission >= read_content
  created_at: string;
  updated_at?: string; // Only if permission >= read_content
  permission: PermissionLevel;
}

export interface TaskGroup {
  id: number;
  guild_id: string;
  name: string;
  description: string;
  color: string;
  created_at: string;
  updated_at: string;
}

export interface TaskGroupPermission {
  id: number;
  group_id: number;
  role_id?: string;
  user_id?: string;
  permission: PermissionLevel;
  created_at: string;
}

export interface CreateBoardRequest {
  name: string;
  description: string;
  color?: string;
}

export interface UpdateBoardRequest {
  name: string;
  description: string;
  color: string;
  position?: number;
}

export interface CreateTaskRequest {
  board_id: number;
  group_id?: number;
  title: string;
  description: string;
  status?: TaskStatus;
  assignee_id?: string;
  start_date?: string;
  due_date?: string;
  tags?: string[];
}

export interface UpdateTaskRequest {
  title: string;
  description: string;
  status: TaskStatus;
  position?: number;
  assignee_id?: string;
  start_date?: string;
  due_date?: string;
  tags?: string[];
}

export interface MoveTaskRequest {
  status: TaskStatus;
  position: number;
}

export interface CreateTaskGroupRequest {
  name: string;
  description: string;
  color?: string;
}

export interface UpdateTaskGroupRequest {
  name: string;
  description: string;
  color: string;
}

export interface SetBoardPermissionRequest {
  role_id?: string;
  user_id?: string;
  can_view_board: boolean;
  can_view_task_list: boolean;
  can_view_tasks: boolean;
  can_edit_tasks: boolean;
  can_edit_board: boolean;
}

export interface SetTaskGroupPermissionRequest {
  role_id?: string;
  user_id?: string;
  permission: PermissionLevel;
}

// Kanban column definition
export interface KanbanColumn {
  id: TaskStatus;
  title: string;
  tasks: FilteredTask[];
}

// Comment and Checklist types
export interface TaskComment {
  id: number;
  task_id: number;
  user_id: string;
  user_name: string;
  user_avatar: string;
  text: string;
  created_at: string;
  updated_at: string;
}

export interface TaskChecklistItem {
  id: number;
  task_id: number;
  text: string;
  is_completed: boolean;
  position: number;
  created_at: string;
  updated_at: string;
}

export interface CreateCommentRequest {
  text: string;
}

export interface CreateChecklistItemRequest {
  text: string;
}

export interface UpdateChecklistItemRequest {
  text?: string;
  is_completed?: boolean;
  position?: number;
}

