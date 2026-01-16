import React from 'react';
import { useDroppable } from '@dnd-kit/core';
import { SortableContext, verticalListSortingStrategy } from '@dnd-kit/sortable';
import type { KanbanColumn } from '../../types/tasks';
import TaskCard from './TaskCard';

interface KanbanColumnProps {
  column: KanbanColumn;
  onTaskClick: (task: any) => void;
}

const KanbanColumnComponent: React.FC<KanbanColumnProps> = ({ column, onTaskClick }) => {
  const { setNodeRef, isOver } = useDroppable({
    id: column.id,
  });

  return (
    <div
      ref={setNodeRef}
      className={`bg-white/5 border border-white/10 rounded-2xl p-4 min-h-[500px] transition-all ${
        isOver ? 'bg-white/10 border-blue-500' : ''
      }`}
    >
      {/* Column Header */}
      <div className="flex items-center justify-between mb-4 pb-3 border-b border-white/10">
        <h3 className="font-bold text-white uppercase text-sm tracking-wide">
          {column.title}
        </h3>
        <span className="px-2 py-1 bg-white/10 rounded-full text-xs text-gray-400">
          {column.tasks.length}
        </span>
      </div>

      {/* Tasks */}
      <SortableContext
        items={column.tasks.map((task) => task.id)}
        strategy={verticalListSortingStrategy}
      >
        <div className="space-y-3">
          {column.tasks.map((task) => (
            <TaskCard
              key={task.id}
              task={task}
              onClick={() => onTaskClick(task)}
            />
          ))}
        </div>
      </SortableContext>

      {column.tasks.length === 0 && (
        <div className="text-center text-gray-500 text-sm py-8">
          No tasks yet
        </div>
      )}
    </div>
  );
};

export default KanbanColumnComponent;
