import React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { FilteredTask } from '../../types/tasks';
import { Clock, CheckSquare, MessageSquare } from 'lucide-react';

interface TaskCardProps {
  task: FilteredTask;
  isDragging?: boolean;
  onClick?: () => void;
}

const TaskCard: React.FC<TaskCardProps> = ({ task, isDragging = false, onClick }) => {
  const {
    attributes,
    listeners,
    setNodeRef,
    transform,
    transition,
    isDragging: isSortableDragging,
  } = useSortable({ id: task.id });

  const style = {
    transform: CSS.Transform.toString(transform),
    transition,
  };

  const canView = task.permission && task.permission !== 'none';
  const canReadTitle = task.permission && ['read_title', 'read_content', 'edit', 'delete'].includes(task.permission);
  const canReadContent = task.permission && ['read_content', 'edit', 'delete'].includes(task.permission);

  if (!canView) {
    return null;
  }

  // Helper to check if due date is overdue
  const isDueOverdue = task.due_date && new Date(task.due_date) < new Date();
  const isDueSoon = task.due_date && !isDueOverdue && 
    (new Date(task.due_date).getTime() - new Date().getTime()) < (24 * 60 * 60 * 1000); // Less than 24 hours

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={onClick}
      className={`bg-[#282c34] rounded-lg shadow-sm border border-white/10 p-3 hover:bg-[#2d3139] hover:shadow-lg transition-all cursor-pointer ${
        isDragging || isSortableDragging ? 'opacity-50 rotate-2' : ''
      }`}
    >
      {/* Color Labels (if tags exist) */}
      {canReadContent && task.tags && Array.isArray(task.tags) && task.tags.length > 0 && (
        <div className="flex gap-1 mb-2">
          {task.tags.slice(0, 5).map((tag: string, index: number) => {
            // Generate consistent colors based on tag name
            const colors = [
              'bg-green-500',
              'bg-yellow-500', 
              'bg-orange-500',
              'bg-red-500',
              'bg-purple-500',
              'bg-blue-500',
              'bg-pink-500',
              'bg-indigo-500',
            ];
            const colorIndex = tag.split('').reduce((acc, char) => acc + char.charCodeAt(0), 0) % colors.length;
            return (
              <div
                key={index}
                className={`h-2 w-10 rounded-full ${colors[colorIndex]}`}
                title={tag}
              />
            );
          })}
        </div>
      )}

      {/* Task Title */}
      <div className="mb-2">
        {canReadTitle ? (
          <h4 className="text-white text-sm font-medium leading-snug hover:text-blue-400 transition-colors">
            {task.title}
          </h4>
        ) : (
          <h4 className="text-gray-500 text-sm italic">
            [Restricted]
          </h4>
        )}
      </div>

      {/* Badges & Meta Info */}
      {canReadContent && (
        <div className="flex flex-wrap items-center gap-2 text-xs">
          {/* Due Date Badge */}
          {task.due_date && (
            <div className={`flex items-center gap-1 px-2 py-1 rounded ${
              isDueOverdue ? 'bg-red-500/20 text-red-400 border border-red-500/30' :
              isDueSoon ? 'bg-yellow-500/20 text-yellow-400 border border-yellow-500/30' :
              'bg-white/5 text-gray-400 border border-white/10'
            }`}>
              <Clock size={12} />
              <span>{new Date(task.due_date).toLocaleDateString('de-DE', { month: 'short', day: 'numeric' })}</span>
            </div>
          )}

          {/* Checklist Progress (placeholder - will be dynamic) */}
          {task.description && task.description.includes('checklist') && (
            <div className="flex items-center gap-1 text-gray-400">
              <CheckSquare size={12} />
              <span>0/0</span>
            </div>
          )}

          {/* Comment Count (placeholder - will be dynamic) */}
          {task.description && task.description.includes('comment') && (
            <div className="flex items-center gap-1 text-gray-400">
              <MessageSquare size={12} />
              <span>0</span>
            </div>
          )}

          {/* Assignee Avatar (if exists) */}
          {task.assignee_id && (
            <div className="ml-auto">
              <div className="w-6 h-6 rounded-full bg-gradient-to-br from-blue-400 to-purple-500 flex items-center justify-center text-white text-xs font-bold">
                {task.assignee_id.substring(0, 2).toUpperCase()}
              </div>
            </div>
          )}
        </div>
      )}
    </div>
  );
};

export default TaskCard;
