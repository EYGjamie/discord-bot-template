import React, { useEffect, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { boardsService } from '../services/tasks';
import type { Board } from '../types/tasks';
import { Plus, Folder } from 'lucide-react';

const BoardsListPage: React.FC = () => {
  const navigate = useNavigate();
  const [boards, setBoards] = useState<Board[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    loadBoards();
  }, []);

  const loadBoards = async () => {
    try {
      setLoading(true);
      const data = await boardsService.getAll();
      setBoards(data || []);
    } catch (err: any) {
      setError(err?.message || 'Failed to load boards');
      setBoards([]);
    } finally {
      setLoading(false);
    }
  };

  if (loading) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-500"></div>
      </div>
    );
  }

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="flex justify-between items-center mb-8">
        <div>
          <h1 className="text-3xl font-bold text-white">Task Boards</h1>
          <p className="text-gray-400 mt-2">
            Manage your tasks across multiple Kanban boards
          </p>
        </div>
        <button
          onClick={() => navigate('/tasks/boards/new')}
          className="flex items-center gap-2 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
        >
          <Plus size={20} />
          New Board
        </button>
      </div>

      {error && (
        <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
          {error}
        </div>
      )}

      {boards.length === 0 ? (
        <div className="bg-white/5 border border-white/10 rounded-2xl p-12 text-center">
          <Folder className="mx-auto mb-4 text-gray-400" size={48} />
          <h3 className="text-xl font-semibold text-white mb-2">No boards yet</h3>
          <p className="text-gray-400 mb-6">
            Create your first board to start organizing tasks
          </p>
          <button
            onClick={() => navigate('/tasks/boards/new')}
            className="px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors"
          >
            Create First Board
          </button>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
          {boards.map((board) => (
            <div
              key={board.id}
              onClick={() => navigate(`/tasks/boards/${board.id}`)}
              className="bg-white/5 border border-white/10 rounded-2xl p-6 hover:bg-white/10 transition-all cursor-pointer group"
            >
              <div className="flex items-start justify-between mb-4">
                <div
                  className="w-12 h-12 rounded-xl flex items-center justify-center text-2xl"
                  style={{ backgroundColor: board.color + '20', color: board.color }}
                >
                  📋
                </div>
              </div>
              <h3 className="text-lg font-bold text-white mb-2 group-hover:text-blue-400 transition-colors">
                {board.name}
              </h3>
              <p className="text-gray-400 text-sm line-clamp-2 mb-4">
                {board.description || 'No description'}
              </p>
              <div className="flex items-center justify-between text-xs text-gray-500">
                <span>Created {new Date(board.created_at).toLocaleDateString()}</span>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default BoardsListPage;
