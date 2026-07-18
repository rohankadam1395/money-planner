import React from 'react';

interface CategoryBadgeProps {
  name: string;
  color?: string;
  icon?: string;
  confidence?: number;
  method?: string;
  llmProvider?: string;
  size?: 'sm' | 'md' | 'lg';
  highlighted?: boolean;
}

// CategoryBadge displays a transaction category with color and icon
export const CategoryBadge: React.FC<CategoryBadgeProps> = ({
  name,
  color = '#95A5A6',
  icon = '📌',
  confidence,
  method,
  llmProvider,
  size = 'md',
  highlighted = false,
}) => {
  const sizeClasses = {
    sm: 'px-2 py-1 text-xs',
    md: 'px-3 py-2 text-sm',
    lg: 'px-4 py-3 text-base',
  };

  const getMethodLabel = (m?: string, provider?: string) => {
    if (!m) return '';
    if (m === 'llm' && provider) {
      return ` (${provider})`;
    }
    return ` (${m === 'rule_based' ? 'known' : m === 'fuzzy' ? 'fuzzy' : m === 'llm' ? 'LLM' : 'manual'})`;
  };

  const isLowConfidence = confidence !== undefined && confidence < 0.75;
  const highlightClass = isLowConfidence && highlighted ? 'ring-2 ring-amber-400 ring-offset-1' : '';

  return (
    <div
      className={`inline-flex items-center gap-2 rounded-full font-medium transition-all ${sizeClasses[size]} ${highlightClass}`}
      style={{
        backgroundColor: isLowConfidence && highlighted ? '#FEF3C7' : `${color}20`,
        borderLeft: `3px solid ${isLowConfidence && highlighted ? '#F59E0B' : color}`,
        color: isLowConfidence && highlighted ? '#D97706' : color,
      }}
      title={confidence ? `Confidence: ${(confidence * 100).toFixed(0)}%${llmProvider ? ` (${llmProvider})` : ''}` : undefined}
    >
      <span>{icon}</span>
      <span>
        {name}
        {getMethodLabel(method, llmProvider)}
      </span>
      {confidence !== undefined && confidence < 1.0 && (
        <span className={`text-xs ${isLowConfidence && highlighted ? 'opacity-100 font-semibold' : 'opacity-75'}`}>
          {(confidence * 100).toFixed(0)}%
        </span>
      )}
      {isLowConfidence && highlighted && (
        <span className="text-xs font-semibold ml-1">⚠️</span>
      )}
    </div>
  );
};

export default CategoryBadge;
