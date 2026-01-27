import React, { useEffect, useState } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { DndContext, DragOverlay, PointerSensor, useSensor, useSensors } from '@dnd-kit/core';
import type { DragEndEvent, DragStartEvent } from '@dnd-kit/core';
import { boardsService, tasksService } from '../services/tasks';
import type { Board, FilteredTask, KanbanColumn, TaskStatus } from '../types/tasks';
import KanbanColumnComponent from '../components/tasks/KanbanColumn';
import TaskCard from '../components/tasks/TaskCard';
import TaskModal from '../components/tasks/TaskModal';
import { ArrowLeft, Plus, Settings } from 'lucide-react';
import { useAuth } from '../hooks/useAuth';

const KanbanBoardPage: React.FC = () => {
  const { boardId } = useParams<{ boardId: string }>();
  const navigate = useNavigate();
  const { user } = useAuth();
  const [board, setBoard] = useState<Board | null>(null);
  const [tasks, setTasks] = useState<FilteredTask[]>([]);
  const [columns, setColumns] = useState<KanbanColumn[]>([]);
  const [loading, setLoading] = useState(true);
  const [canEditBoard, setCanEditBoard] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [activeTask, setActiveTask] = useState<FilteredTask | null>(null);
  const [selectedTask, setSelectedTask] = useState<FilteredTask | null>(null);
  const [showTaskModal, setShowTaskModal] = useState(false);
  const [showCreateModal, setShowCreateModal] = useState(false);

  const sensors = useSensors(
    useSensor(PointerSensor, {
      activationConstraint: {
        distance: 8,
      },
    })
  );

  const columnDefinitions: Array<{ id: TaskStatus; title: string }> = [
    { id: 'todo', title: 'To Do' },
    { id: 'in_progress', title: 'In Progress' },
    { id: 'review', title: 'Review Pending' },
    { id: 'done', title: 'Done' },
  ];

  useEffect(() => {
    if (boardId) {
      loadBoard();
      loadTasks();
    }
  }, [boardId]);

  useEffect(() => {
    if (boardId && user?.discord_id) {
      checkBoardEditPermission();
    }
  }, [boardId, user]);

  useEffect(() => {
    organizeTasksIntoColumns();
  }, [tasks]);

  const loadBoard = async () => {
    try {
      const data = await boardsService.getById(Number(boardId));
      setBoard(data);
    } catch (err: any) {
      setError(err?.message || 'Failed to load board');
    }
  };

  const checkBoardEditPermission = async () => {
    if (!user?.discord_id) return;
    
    try {
      const permissions = await boardsService.getPermissions(Number(boardId));
      
      // Check if user has explicit can_edit_board permission
      const userPermission = permissions.find(
        (p) => p.user_id === user.discord_id && p.can_edit_board
      );
      
      if (userPermission) {
        setCanEditBoard(true);
        return;
      }
      
      // Check if any of user's roles have can_edit_board permission
      if (user.roles && user.roles.length > 0) {
        const roleIds = user.roles.map(r => r.id);
        const rolePermission = permissions.find(
          (p) => p.role_id && roleIds.includes(p.role_id) && p.can_edit_board
        );
        
        if (rolePermission) {
          setCanEditBoard(true);
          return;
        }
      }
      
      setCanEditBoard(false);
    } catch (err) {
      console.error('Failed to check board edit permission:', err);
      setCanEditBoard(false);
    }
  };

  const loadTasks = async () => {
    try {
      setLoading(true);
      const data = await tasksService.getByBoard(Number(boardId));
      setTasks(data || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load tasks');
      setTasks([]);
    } finally {
      setLoading(false);
    }
  };

  const organizeTasksIntoColumns = () => {
    const newColumns: KanbanColumn[] = columnDefinitions.map((col) => ({
      id: col.id,
      title: col.title,
      tasks: (tasks || [])
        .filter((task) => task.status === col.id)
        .sort((a, b) => a.position - b.position),
    }));
    setColumns(newColumns);
  };

  const handleDragStart = (event: DragStartEvent) => {
    const taskId = event.active.id as number;
    const task = tasks.find((t) => t.id === taskId);
    if (task) {
      setActiveTask(task);
    }
  };

  const handleDragEnd = async (event: DragEndEvent) => {
    const { active, over } = event;
    setActiveTask(null);

    if (!over) return;

    const taskId = active.id as number;
    const task = tasks.find((t) => t.id === taskId);
    if (!task) return;

    // Check if task has edit permission
    if (task.permission !== 'edit' && task.permission !== 'delete') {
      setError('You do not have permission to move this task');
      return;
    }

    const overId = over.id as string;
    let newStatus: TaskStatus;
    let newPosition = 0;

    // Check if dropped on a column or another task
    if (columnDefinitions.some((col) => col.id === overId)) {
      newStatus = overId as TaskStatus;
      const targetColumn = columns.find((col) => col.id === newStatus);
      newPosition = targetColumn ? targetColumn.tasks.length : 0;
    } else {
      const overTask = tasks.find((t) => t.id === Number(overId));
      if (overTask) {
        newStatus = overTask.status;
        newPosition = overTask.position;
      } else {
        return;
      }
    }

    // Only update if status or position changed
    if (task.status === newStatus && task.position === newPosition) {
      return;
    }

    try {
      await tasksService.move(taskId, { status: newStatus, position: newPosition });
      await loadTasks(); // Reload to get updated positions
    } catch (err: any) {
      setError(err.response?.data?.message || 'Failed to move task');
    }
  };

  const handleTaskClick = (task: FilteredTask) => {
    setSelectedTask(task);
    setShowTaskModal(true);
  };

  const handleCreateTask = () => {
    setSelectedTask(null);
    setShowCreateModal(true);
  };

  const handleTaskUpdated = () => {
    loadTasks();
    setShowTaskModal(false);
    setShowCreateModal(false);
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
    <div className="min-h-screen bg-gradient-to-br from-gray-900 via-blue-900/20 to-gray-900">
      <div className="container mx-auto px-4 py-6">
        {/* Header */}
        <div className="flex items-center justify-between mb-6">
          <div className="flex items-center gap-4">
            <button
              onClick={() => navigate('/tasks')}
              className="p-2 hover:bg-white/10 rounded-lg transition-colors"
            >
              <ArrowLeft size={20} className="text-white" />
            </button>
            <div>
              <h1 className="text-2xl font-bold text-white">{board.name}</h1>
              <p className="text-gray-400 text-sm">{board.description}</p>
            </div>
          </div>
          <div className="flex items-center gap-2">
            <button
              onClick={handleCreateTask}
              className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
            >
              <Plus size={20} />
              New Task
            </button>
            {(user?.is_admin || board?.created_by === user?.discord_id || canEditBoard) && (
              <button
                onClick={() => navigate(`/tasks/boards/${boardId}/settings`)}
                className="p-2 hover:bg-white/10 rounded-lg transition-colors"
              >
                <Settings size={20} className="text-white" />
              </button>
            )}
          </div>
        </div>

        {error && (
          <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
            {error}
          </div>
        )}

        {/* Kanban Board */}
        <DndContext
          sensors={sensors}
          onDragStart={handleDragStart}
          onDragEnd={handleDragEnd}
        >
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {columns.map((column) => (
              <KanbanColumnComponent
                key={column.id}
                column={column}
                onTaskClick={handleTaskClick}
              />
            ))}
          </div>

          <DragOverlay>
            {activeTask ? <TaskCard task={activeTask} isDragging /> : null}
          </DragOverlay>
        </DndContext>
      </div>

      {/* Task Detail Modal */}
      {showTaskModal && selectedTask && (
        <TaskModal
          task={selectedTask}
          onClose={() => setShowTaskModal(false)}
          onUpdate={handleTaskUpdated}
        />
      )}

      {/* Create Task Modal */}
      {showCreateModal && (
        <TaskModal
          boardId={Number(boardId)}
          onClose={() => setShowCreateModal(false)}
          onUpdate={handleTaskUpdated}
        />
      )}
    </div>
  );
};

export default KanbanBoardPage;
