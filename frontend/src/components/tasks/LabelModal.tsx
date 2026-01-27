import React, { useState, useEffect } from 'react';
import { X, Plus, Tag, Palette } from 'lucide-react';
import { LABEL_COLORS, type BoardLabel } from '../../utils/labelColors';

interface LabelModalProps {
  boardId: number;
  currentLabels: string[];
  onClose: () => void;
  onSave: (labels: string[]) => void;
}

const LabelModal: React.FC<LabelModalProps> = ({ boardId, currentLabels, onClose, onSave }) => {
  const [selectedLabels, setSelectedLabels] = useState<string[]>(currentLabels);
  const [newLabelName, setNewLabelName] = useState('');
  const [newLabelColor, setNewLabelColor] = useState('blue');
  const [boardLabels, setBoardLabels] = useState<BoardLabel[]>([]);
  const [editingLabel, setEditingLabel] = useState<string | null>(null);

  // Load board labels from localStorage
  useEffect(() => {
    const storedLabels = localStorage.getItem(`board_${boardId}_labels`);
    if (storedLabels) {
      try {
        const parsed = JSON.parse(storedLabels);
        // Handle old format (array of strings) and new format (array of objects)
        if (Array.isArray(parsed) && parsed.length > 0) {
          if (typeof parsed[0] === 'string') {
            // Old format - convert to new format
            const converted = parsed.map((name: string) => ({ name, color: 'blue' }));
            setBoardLabels(converted);
            saveBoardLabels(converted);
          } else {
            setBoardLabels(parsed);
          }
        }
      } catch (e) {
        console.error('Failed to parse labels:', e);
      }
    }
  }, [boardId]);

  // Save board labels to localStorage
  const saveBoardLabels = (labels: BoardLabel[]) => {
    localStorage.setItem(`board_${boardId}_labels`, JSON.stringify(labels));
    setBoardLabels(labels);
  };

  const toggleLabel = (labelName: string) => {
    if (selectedLabels.includes(labelName)) {
      setSelectedLabels(selectedLabels.filter(l => l !== labelName));
    } else {
      setSelectedLabels([...selectedLabels, labelName]);
    }
  };

  const addNewLabel = () => {
    if (newLabelName.trim() && !boardLabels.find(l => l.name === newLabelName.trim())) {
      const newLabel: BoardLabel = { name: newLabelName.trim(), color: newLabelColor };
      const updatedLabels = [...boardLabels, newLabel];
      saveBoardLabels(updatedLabels);
      setSelectedLabels([...selectedLabels, newLabelName.trim()]);
      setNewLabelName('');
      setNewLabelColor('blue');
    }
  };

  const updateLabelColor = (labelName: string, newColor: string) => {
    const updatedLabels = boardLabels.map(l => 
      l.name === labelName ? { ...l, color: newColor } : l
    );
    saveBoardLabels(updatedLabels);
    setEditingLabel(null);
  };

  const deleteLabel = (labelName: string) => {
    if (confirm(`Label "${labelName}" vom Board entfernen?`)) {
      const updatedLabels = boardLabels.filter(l => l.name !== labelName);
      saveBoardLabels(updatedLabels);
      setSelectedLabels(selectedLabels.filter(l => l !== labelName));
    }
  };

  const handleSave = () => {
    onSave(selectedLabels);
    onClose();
  };

  const getLabelColor = (labelName: string) => {
    const label = boardLabels.find(l => l.name === labelName);
    const colorValue = label?.color || 'blue';
    return LABEL_COLORS.find(c => c.value === colorValue) || LABEL_COLORS[5];
  };

  return (
    <div className="fixed inset-0 bg-black/60 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-[#1e2228] rounded-lg w-full max-w-md border border-white/10 shadow-2xl" onClick={(e) => e.stopPropagation()}>
        {/* Header */}
        <div className="flex items-center justify-between p-4 border-b border-white/10">
          <div className="flex items-center gap-2">
            <Tag size={20} className="text-blue-400" />
            <h2 className="text-lg font-semibold text-white">Labels verwalten</h2>
          </div>
          <button
            onClick={onClose}
            className="p-1 hover:bg-white/10 rounded transition-colors"
          >
            <X size={20} className="text-gray-400" />
          </button>
        </div>

        {/* Content */}
        <div className="p-4 space-y-4">
          {/* Add New Label */}
          <div>
            <label className="text-sm font-medium text-gray-400 mb-2 block">Neues Label erstellen</label>
            <div className="space-y-2">
              <input
                type="text"
                value={newLabelName}
                onChange={(e) => setNewLabelName(e.target.value)}
                onKeyPress={(e) => e.key === 'Enter' && addNewLabel()}
                placeholder="Label Name..."
                className="w-full px-3 py-2 bg-[#0d0f15] border border-white/10 rounded-lg text-gray-300 placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-blue-500 text-sm"
              />
              <div className="flex gap-2 items-center">
                <label className="text-xs text-gray-400">Farbe:</label>
                <div className="grid grid-cols-6 gap-1 flex-1">
                  {LABEL_COLORS.map((color) => (
                    <button
                      key={color.value}
                      type="button"
                      onClick={() => setNewLabelColor(color.value)}
                      className={`w-8 h-8 rounded ${color.bg} border-2 ${
                        newLabelColor === color.value ? 'border-white' : 'border-transparent'
                      } hover:opacity-80 transition-all`}
                      title={color.name}
                    />
                  ))}
                </div>
                <button
                  onClick={addNewLabel}
                  disabled={!newLabelName.trim()}
                  className="px-3 py-1.5 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-xs font-medium disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-1"
                >
                  <Plus size={14} />
                  Hinzufügen
                </button>
              </div>
            </div>
          </div>

          {/* Available Labels */}
          <div>
            <label className="text-sm font-medium text-gray-400 mb-2 block">
              Verfügbare Labels ({boardLabels.length})
            </label>
            <div className="space-y-2 max-h-64 overflow-y-auto pr-2">
              {boardLabels.length === 0 ? (
                <p className="text-sm text-gray-500 italic text-center py-4">
                  Noch keine Labels vorhanden. Erstelle dein erstes Label oben!
                </p>
              ) : (
                boardLabels.map((label) => {
                  const color = getLabelColor(label.name);
                  const isSelected = selectedLabels.includes(label.name);
                  const isEditing = editingLabel === label.name;
                  return (
                    <div
                      key={label.name}
                      className="flex items-center gap-2 bg-[#0d0f15] border border-white/10 rounded-lg p-2 hover:bg-white/5 transition-colors"
                    >
                      <input
                        type="checkbox"
                        checked={isSelected}
                        onChange={() => toggleLabel(label.name)}
                        className="w-4 h-4 rounded border-white/20 bg-[#0d0f15] text-blue-600 focus:ring-2 focus:ring-blue-500 focus:ring-offset-0 cursor-pointer"
                      />
                      <span className={`flex-1 px-3 py-1 rounded ${color.bg} ${color.text} text-sm font-medium`}>
                        {label.name}
                      </span>
                      {isEditing ? (
                        <div className="grid grid-cols-6 gap-1">
                          {LABEL_COLORS.map((c) => (
                            <button
                              key={c.value}
                              type="button"
                              onClick={() => updateLabelColor(label.name, c.value)}
                              className={`w-6 h-6 rounded ${c.bg} border ${
                                label.color === c.value ? 'border-white' : 'border-transparent'
                              } hover:opacity-80`}
                              title={c.name}
                            />
                          ))}
                        </div>
                      ) : (
                        <>
                          <button
                            onClick={() => setEditingLabel(label.name)}
                            className="text-gray-400 hover:text-blue-400 p-1"
                            title="Farbe ändern"
                          >
                            <Palette size={16} />
                          </button>
                          <button
                            onClick={() => deleteLabel(label.name)}
                            className="text-gray-500 hover:text-red-500 text-lg px-2"
                            title="Label löschen"
                          >
                            ×
                          </button>
                        </>
                      )}
                    </div>
                  );
                })
              )}
            </div>
          </div>

          {/* Selected Labels Preview */}
          {selectedLabels.length > 0 && (
            <div>
              <label className="text-sm font-medium text-gray-400 mb-2 block">
                Ausgewählte Labels ({selectedLabels.length})
              </label>
              <div className="flex flex-wrap gap-2 p-3 bg-[#0d0f15] border border-white/10 rounded-lg">
                {selectedLabels.map((label) => {
                  const color = getLabelColor(label);
                  return (
                    <span key={label} className={`px-3 py-1 rounded ${color.bg} ${color.text} text-sm font-medium`}>
                      {label}
                    </span>
                  );
                })}
              </div>
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="flex gap-3 p-4 border-t border-white/10">
          <button
            onClick={onClose}
            className="flex-1 px-4 py-2 bg-white/10 hover:bg-white/20 text-white rounded-lg transition-colors text-sm font-medium"
          >
            Abbrechen
          </button>
          <button
            onClick={handleSave}
            className="flex-1 px-4 py-2 bg-blue-600 hover:bg-blue-700 text-white rounded-lg transition-colors text-sm font-medium"
          >
            Speichern
          </button>
        </div>
      </div>
    </div>
  );
};

export default LabelModal;
