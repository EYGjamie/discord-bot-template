import React, { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { boardsService } from '../services/tasks';
import { ArrowLeft } from 'lucide-react';

const CreateBoardPage: React.FC = () => {
  const navigate = useNavigate();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [formData, setFormData] = useState({
    name: '',
    description: '',
    color: '#3B82F6',
  });

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    
    if (!formData.name.trim()) {
      setError('Board name is required');
      return;
    }

    try {
      setLoading(true);
      setError(null);
      const board = await boardsService.create({
        name: formData.name,
        description: formData.description,
        color: formData.color,
      });
      navigate(`/tasks/boards/${board.id}`);
    } catch (err: any) {
      setError(err?.message || 'Failed to create board');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="container mx-auto px-4 py-8 max-w-2xl">
      <button
        onClick={() => navigate('/tasks')}
        className="flex items-center gap-2 text-gray-400 hover:text-white mb-6 transition-colors"
      >
        <ArrowLeft size={20} />
        Back to Boards
      </button>

      <div className="bg-white/5 border border-white/10 rounded-2xl p-8">
        <h1 className="text-3xl font-bold text-white mb-2">Create New Board</h1>
        <p className="text-gray-400 mb-8">
          Set up a new Kanban board to organize your tasks
        </p>

        {error && (
          <div className="bg-red-500/10 border border-red-500 text-red-500 px-4 py-3 rounded-lg mb-6">
            {error}
          </div>
        )}

        <form onSubmit={handleSubmit} className="space-y-6">
          <div>
            <label htmlFor="name" className="block text-sm font-medium text-gray-300 mb-2">
              Board Name *
            </label>
            <input
              type="text"
              id="name"
              value={formData.name}
              onChange={(e) => setFormData({ ...formData, name: e.target.value })}
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
              placeholder="e.g., Sprint Planning, Feature Development"
              required
            />
          </div>

          <div>
            <label htmlFor="description" className="block text-sm font-medium text-gray-300 mb-2">
              Description
            </label>
            <textarea
              id="description"
              value={formData.description}
              onChange={(e) => setFormData({ ...formData, description: e.target.value })}
              rows={4}
              className="w-full px-4 py-3 bg-white/5 border border-white/10 rounded-lg text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent resize-none"
              placeholder="Describe the purpose of this board..."
            />
          </div>

          <div>
            <label htmlFor="color" className="block text-sm font-medium text-gray-300 mb-2">
              Board Color
            </label>
            <div className="flex items-center gap-4">
              <input
                type="color"
                id="color"
                value={formData.color}
                onChange={(e) => setFormData({ ...formData, color: e.target.value })}
                className="h-12 w-20 cursor-pointer rounded-lg"
              />
              <span className="text-gray-400 text-sm">{formData.color}</span>
            </div>
            <p className="text-gray-500 text-sm mt-2">
              Choose a color to help identify this board
            </p>
          </div>

          <div className="flex gap-4 pt-4">
            <button
              type="button"
              onClick={() => navigate('/tasks')}
              className="flex-1 px-6 py-3 bg-white/5 hover:bg-white/10 text-white rounded-lg transition-colors"
              disabled={loading}
            >
              Cancel
            </button>
            <button
              type="submit"
              className="flex-1 px-6 py-3 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
              disabled={loading}
            >
              {loading ? 'Creating...' : 'Create Board'}
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};

export default CreateBoardPage;
