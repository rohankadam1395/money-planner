-- T009: Create category_stats table for analytics
CREATE TABLE IF NOT EXISTS category_stats (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  category_id UUID NOT NULL REFERENCES categories(id),
  period VARCHAR(10) NOT NULL,
  total_spent DECIMAL(12, 2) DEFAULT 0,
  transaction_count INT DEFAULT 0,
  average_transaction DECIMAL(12, 2) DEFAULT 0,
  min_transaction DECIMAL(12, 2),
  max_transaction DECIMAL(12, 2),
  last_transaction_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, category_id, period)
);

CREATE INDEX IF NOT EXISTS idx_cat_stats_user_period ON category_stats(user_id, period DESC);
CREATE INDEX IF NOT EXISTS idx_cat_stats_category_period ON category_stats(category_id, period DESC);
