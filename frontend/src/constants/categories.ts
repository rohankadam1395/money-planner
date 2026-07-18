export const CATEGORY_MAP: Record<string, { color: string; icon: string }> = {
  'Food & Dining': { color: '#FF6B6B', icon: '🍔' },
  'Shopping': { color: '#4ECDC4', icon: '🛍️' },
  'Transport': { color: '#45B7D1', icon: '🚗' },
  'Housing': { color: '#F7B731', icon: '🏠' },
  'Utilities': { color: '#5F27CD', icon: '💡' },
  'Entertainment': { color: '#EE5A6F', icon: '🎬' },
  'Income': { color: '#2ECC71', icon: '💰' },
  'Healthcare': { color: '#FF4757', icon: '🏥' },
  'Education': { color: '#1E90FF', icon: '📚' },
  'Miscellaneous': { color: '#95A5A6', icon: '📌' },
  'Uncategorized': { color: '#CCCCCC', icon: '❓' },
};

export const getCategoryStyle = (
  categoryName: string
): { color: string; icon: string } => {
  return CATEGORY_MAP[categoryName] || CATEGORY_MAP['Uncategorized'];
};
