-- T008: Create transaction_categories table for categorization assignments
CREATE TABLE IF NOT EXISTS transaction_categories (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID NOT NULL REFERENCES users(user_id),
  transaction_id UUID NOT NULL REFERENCES transactions(transaction_id),
  category_id UUID NOT NULL REFERENCES categories(id),
  method VARCHAR(50) NOT NULL,
  llm_provider VARCHAR(50),
  confidence FLOAT DEFAULT 1.0,
  llm_explanation TEXT,
  assigned_by_user_id UUID REFERENCES users(user_id),
  assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(transaction_id)
);

CREATE INDEX IF NOT EXISTS idx_txn_cat_user_assigned ON transaction_categories(user_id, assigned_at DESC);
CREATE INDEX IF NOT EXISTS idx_txn_cat_category ON transaction_categories(category_id);
CREATE INDEX IF NOT EXISTS idx_txn_cat_method ON transaction_categories(method);
CREATE INDEX IF NOT EXISTS idx_txn_cat_llm_provider ON transaction_categories(llm_provider);
