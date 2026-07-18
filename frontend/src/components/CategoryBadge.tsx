import React from 'react';

interface CategoryBadgeProps {
  name: string;
  color?: string;
  icon?: string;
  confidence?: number;
  method?: string;
  size?: 'sm' | 'md' | 'lg';
}

// CategoryBadge displays a transaction category with color and icon
export const CategoryBadge: React.FC<CategoryBadgeProps> = ({
  name,
  color = '#95A5A6',
  icon = '📌',
  confidence,
  method,
  size = 'md',
}) => {
  const sizeClasses = {
    sm: 'px-2 py-1 text-xs',
    md: 'px-3 py-2 text-sm',
    lg: 'px-4 py-3 text-base',
  };

  const getMethodLabel = (m?: string) => {
    if (!m) return '';
    return ` (${m === 'rule_based' ? 'known' : m === 'fuzzy' ? 'fuzzy' : 'manual'})`;
  };

  return (
    <div
      className={`inline-flex items-center gap-2 rounded-full font-medium transition-all ${sizeClasses[size]}`}
      style={{
        backgroundColor: `${color}20`,
        borderLeft: `3px solid ${color}`,
        color: color,
      }}
      title={confidence ? `Confidence: ${(confidence * 100).toFixed(0)}%` : undefined}
    >
      <span>{icon}</span>
      <span>
        {name}
        {getMethodLabel(method)}
      </span>
      {confidence !== undefined && confidence < 1.0 && (
        <span className="text-xs opacity-75">
          {(confidence * 100).toFixed(0)}%
        </span>
      )}
    </div>
  );
};

export default CategoryBadge;
