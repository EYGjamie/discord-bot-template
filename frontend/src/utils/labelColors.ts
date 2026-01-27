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

export interface BoardLabel {
  name: string;
  color: string;
}

export const getLabelColorFromBoard = (boardId: number, labelName: string): LabelColor => {
  if (!boardId || !labelName) {
    return LABEL_COLORS[5]; // Default to blue
  }
  
  try {
    const storedLabels = localStorage.getItem(`board_${boardId}_labels`);
    if (storedLabels) {
      const parsed: BoardLabel[] = JSON.parse(storedLabels);
      const label = parsed.find((l: BoardLabel) => l.name === labelName);
      if (label) {
        const color = LABEL_COLORS.find(c => c.value === label.color);
        if (color) return color;
      }
    }
  } catch (e) {
    console.error('Failed to parse labels:', e);
  }
  
  // Default to blue if not found
  return LABEL_COLORS[5];
};
