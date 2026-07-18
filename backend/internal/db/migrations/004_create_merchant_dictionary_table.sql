-- T007: Create merchant_dictionary table for rule-based categorization
CREATE TABLE IF NOT EXISTS merchant_dictionary (
  id UUID PRIMARY KEY,
  merchant_name VARCHAR(255) NOT NULL,
  merchant_pattern VARCHAR(255),
  category_id UUID NOT NULL REFERENCES categories(id),
  source VARCHAR(50),
  confidence INT DEFAULT 100,
  match_type VARCHAR(50),
  frequency INT DEFAULT 0,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(merchant_name, category_id)
);

CREATE INDEX IF NOT EXISTS idx_merchant_name ON merchant_dictionary(merchant_name);
CREATE INDEX IF NOT EXISTS idx_merchant_category_id ON merchant_dictionary(category_id);
