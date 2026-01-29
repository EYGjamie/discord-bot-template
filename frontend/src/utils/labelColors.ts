export interface LabelColor {
  value: string;
  bg: string;
  text: string;
  name: string;
}

export const LABEL_COLORS: LabelColor[] = [
  { value: 'green', bg: 'bg-green-600', text: 'text-white', name: 'Grün' },
  { value: 'yellow', bg: 'bg-yellow-500', text: 'text-gray-900', name: 'Gelb' },
  { value: 'orange', bg: 'bg-orange-600', text: 'text-white', name: 'Orange' },
  { value: 'red', bg: 'bg-red-600', text: 'text-white', name: 'Rot' },
  { value: 'purple', bg: 'bg-purple-600', text: 'text-white', name: 'Lila' },
  { value: 'blue', bg: 'bg-blue-600', text: 'text-white', name: 'Blau' },
  { value: 'pink', bg: 'bg-pink-600', text: 'text-white', name: 'Pink' },
  { value: 'indigo', bg: 'bg-indigo-600', text: 'text-white', name: 'Indigo' },
  { value: 'teal', bg: 'bg-teal-600', text: 'text-white', name: 'Türkis' },
  { value: 'lime', bg: 'bg-lime-600', text: 'text-white', name: 'Limette' },
  { value: 'cyan', bg: 'bg-cyan-600', text: 'text-white', name: 'Cyan' },
  { value: 'rose', bg: 'bg-rose-600', text: 'text-white', name: 'Rose' },
];

// Local cache for board labels
interface CachedBoardLabel {
  name: string;
  color: string;
}

const labelCache: Map<number, CachedBoardLabel[]> = new Map();

// Set labels in cache (call this when loading board data)
export const setBoardLabelsCache = (boardId: number, labels: CachedBoardLabel[]) => {
  labelCache.set(boardId, labels);
};

// Clear labels from cache
export const clearBoardLabelsCache = (boardId?: number) => {
  if (boardId !== undefined) {
    labelCache.delete(boardId);
  } else {
    labelCache.clear();
  }
};

// Get label color from cache (synchronous for rendering)
export const getLabelColorFromBoard = (boardId: number, labelName: string): LabelColor => {
  if (!boardId || !labelName) {
    return LABEL_COLORS[5]; // Default to blue
  }
  
  const cachedLabels = labelCache.get(boardId);
  if (cachedLabels) {
    const label = cachedLabels.find(l => l.name === labelName);
    if (label) {
      const color = LABEL_COLORS.find(c => c.value === label.color);
      if (color) return color;
    }
  }
  
  // Default to blue if not found
  return LABEL_COLORS[5];
};
