import React from 'react';
import { useSortable } from '@dnd-kit/sortable';
import { CSS } from '@dnd-kit/utilities';
import type { FilteredTask, PermissionLevel } from '../../types/tasks';
import { Lock, Calendar, User } from 'lucide-react';

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
  const isRestricted = !canReadContent;

  const getPermissionIcon = (permission: PermissionLevel) => {
    if (permission === 'existence' || permission === 'read_title') {
      return <Lock size={14} className="text-yellow-500" />;
    }
    return null;
  };

  if (!canView) {
    return null; // Don't render if user has no permission
  }

  return (
    <div
      ref={setNodeRef}
      style={style}
      {...attributes}
      {...listeners}
      onClick={onClick}
      className={`bg-white/5 border border-white/10 rounded-xl p-3 hover:bg-white/10 transition-all cursor-pointer group ${
        isDragging || isSortableDragging ? 'opacity-50' : ''
      }`}
    >
      {/* Task Header */}
      <div className="flex items-start justify-between mb-2">
        {canReadTitle ? (
          <h4 className="font-bold text-white text-sm group-hover:text-blue-400 transition-colors flex-1">
            {task.title}
          </h4>
        ) : (
          <h4 className="font-bold text-gray-500 text-sm italic flex-1">
            [Restricted]
          </h4>
        )}
        {isRestricted && getPermissionIcon(task.permission)}
      </div>

      {/* Description Preview (only if can read content) */}
      {canReadContent && task.description && (
        <p className="text-gray-400 text-xs line-clamp-2 mb-3">
          {task.description}
        </p>
      )}

      {/* Tags (only if can read content) */}
      {canReadContent && task.tags && Array.isArray(task.tags) && task.tags.length > 0 && (
        <div className="flex flex-wrap gap-1 mb-3">
          {task.tags.slice(0, 3).map((tag: string, index: number) => (
            <span
              key={index}
              className="px-2 py-1 bg-blue-500/20 text-blue-400 rounded-full text-xs font-semibold"
            >
              {tag}
            </span>
          ))}
        </div>
      )}

      {/* Meta Info (always visible based on permission level) */}
      <div className="flex items-center justify-between text-xs text-gray-500">
        <div className="flex items-center gap-2">
          {task.assignee_id && (
            <div className="flex items-center gap-1">
              <User size={12} />
              <span>Assigned</span>
            </div>
          )}
        </div>
        {task.due_date && (
          <div className="flex items-center gap-1">
            <Calendar size={12} />
            <span>{new Date(task.due_date).toLocaleDateString()}</span>
          </div>
        )}
      </div>

      {/* Permission Indicator */}
      {task.permission === 'existence' && (
        <div className="mt-2 pt-2 border-t border-white/10">
          <div className="text-xs text-gray-500 italic">Limited access</div>
        </div>
      )}
    </div>
  );
};

export default TaskCard;
