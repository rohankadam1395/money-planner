# Data Model: Transaction Categorization

**Phase 1 Output** | Generated: 2026-07-12

Entity definitions and relationships for transaction categorization feature.

## Entities

### 1. Category

Represents a predefined transaction category (e.g., "Food", "Shopping").

```sql
CREATE TABLE categories (
  id UUID PRIMARY KEY,
  name VARCHAR(50) NOT NULL UNIQUE,          -- "Food", "Shopping", etc.
  description TEXT,                           -- "Restaurants, food delivery services"
  color VARCHAR(7),                          -- Hex color for UI display, e.g. "#FF6B6B"
  icon VARCHAR(50),                          -- Icon name for UI, e.g. "utensils"
  is_predefined BOOLEAN DEFAULT true,        -- true for system categories, false for future custom
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Validation Rules**:
- `name` is required, 3-50 characters
- `color` must be valid hex color (e.g., #FF6B6B)
- `is_predefined` immutable after creation

**Relationships**:
- One-to-many with MerchantDictionary
- One-to-many with TransactionCategory

---

### 2. MerchantDictionary

Maps merchant names to categories. Serves as the fast lookup table for rule-based categorization.

```sql
CREATE TABLE merchant_dictionary (
  id UUID PRIMARY KEY,
  merchant_name VARCHAR(255) NOT NULL,       -- "Swiggy", "Amazon", etc.
  merchant_pattern VARCHAR(255),              -- Regex or fuzzy match pattern (future)
  category_id UUID NOT NULL REFERENCES categories(id),
  source VARCHAR(50),                        -- "manual", "llm_learned", "user_correction"
  confidence INT DEFAULT 100,                -- 100 for exact, 85-99 for fuzzy
  match_type VARCHAR(50),                    -- "exact", "fuzzy", "pattern"
  frequency INT DEFAULT 0,                   -- How many times matched (for trending)
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(merchant_name, category_id)
);

CREATE INDEX idx_merchant_name ON merchant_dictionary(merchant_name);
CREATE INDEX idx_category_id ON merchant_dictionary(category_id);
```

**Validation Rules**:
- `merchant_name` required, 2-255 characters, case-insensitive lookup
- `category_id` must reference existing category
- `source` must be one of: "manual", "llm_learned", "user_correction"
- `match_type` must be one of: "exact", "fuzzy", "pattern"
- `confidence` must be 0-100

**Relationships**:
- Many-to-one with Category
- Logical relationship to TransactionCategory (used during categorization)

---

### 3. TransactionCategory

Assignment of a category to a transaction. Stores the categorization decision and its provenance, including which LLM provider was used (if any).

```sql
CREATE TABLE transaction_categories (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  transaction_id UUID NOT NULL REFERENCES transactions(id),
  category_id UUID NOT NULL REFERENCES categories(id),
  method VARCHAR(50) NOT NULL,               -- "rule_based", "fuzzy", "llm", "manual"
  llm_provider VARCHAR(50),                  -- "ollama", "claude", "openai", NULL if not LLM
  confidence FLOAT DEFAULT 1.0,              -- 0.0-1.0 for LLM; 1.0 for rule/manual
  llm_explanation TEXT,                      -- Why LLM chose this category
  assigned_by_user_id UUID REFERENCES users(id),  -- NULL if system-assigned
  assigned_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(transaction_id)
);

CREATE INDEX idx_user_id_assigned_at ON transaction_categories(user_id, assigned_at DESC);
CREATE INDEX idx_category_id ON transaction_categories(category_id);
CREATE INDEX idx_method ON transaction_categories(method);
CREATE INDEX idx_llm_provider ON transaction_categories(llm_provider);
```

**Validation Rules**:
- `user_id` and `transaction_id` required
- `category_id` must reference existing category
- `method` must be one of: "rule_based", "fuzzy", "llm", "manual"
- `llm_provider` must be one of: "ollama", "claude", "openai" (or NULL if method != "llm")
- `confidence` must be 0.0-1.0
- `assigned_at` immutable after creation
- Only one category per transaction (enforced by UNIQUE constraint)

**Relationships**:
- Many-to-one with User
- One-to-one with Transaction (via transaction_id)
- Many-to-one with Category
- Optional many-to-one with User (assigned_by_user_id for manual corrections)

---

### 4. CategoryStats

Pre-computed statistics for fast analytics queries. Updated incrementally as transactions are added/recategorized.

```sql
CREATE TABLE category_stats (
  id UUID PRIMARY KEY,
  user_id UUID NOT NULL REFERENCES users(id),
  category_id UUID NOT NULL REFERENCES categories(id),
  period VARCHAR(10) NOT NULL,               -- "2026-07" (YYYY-MM for monthly)
  total_spent DECIMAL(12, 2) DEFAULT 0,      -- Sum of transaction amounts
  transaction_count INT DEFAULT 0,           -- Count of transactions
  average_transaction DECIMAL(12, 2) DEFAULT 0,
  min_transaction DECIMAL(12, 2),
  max_transaction DECIMAL(12, 2),
  last_transaction_at TIMESTAMP,
  created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
  UNIQUE(user_id, category_id, period)
);

CREATE INDEX idx_user_period ON category_stats(user_id, period DESC);
CREATE INDEX idx_category_period ON category_stats(category_id, period DESC);
```

**Validation Rules**:
- `user_id`, `category_id`, `period` required
- `period` format: YYYY-MM (e.g., "2026-07")
- All amount fields must be >= 0
- `average_transaction` computed as total_spent / transaction_count

**Relationships**:
- Many-to-one with User
- Many-to-one with Category
- Derived from TransactionCategory (aggregate view)

---

## State Transitions

### TransactionCategory Lifecycle

```
UNCATEGORIZED (on import)
    ↓
RULE_BASED (if merchant found in dictionary)
    ↓ (user corrects)
MANUAL
    ↓ (if user confirms)
FINAL

OR

UNCATEGORIZED (on import)
    ↓
LLM_SUGGESTED (if merchant unknown, LLM categorized)
    ↓ (user reviews)
    ├→ ACCEPTED (user confirms)
    ├→ REJECTED (user corrects)
    └→ UNCATEGORIZED (user skips)
    ↓ (if user corrects)
MANUAL
    ↓
FINAL
```

---

## Relationships Diagram

```
Category (1) ←─ (M) MerchantDictionary
             ←─ (M) TransactionCategory
             ←─ (M) CategoryStats

Transaction (1) ←─ (1) TransactionCategory ─→ (M) Category

User (1) ←─ (M) TransactionCategory
         ←─ (M) CategoryStats

MerchantDictionary (assists categorization, not a direct FK)
  └─→ Used during categorization process to determine category_id
```

---

## Queries (High-Level)

### Get transactions by category
```sql
SELECT t.* FROM transactions t
JOIN transaction_categories tc ON t.id = tc.transaction_id
WHERE tc.user_id = ? AND tc.category_id = ?
ORDER BY t.date DESC;
```

### Get category totals for a month
```sql
SELECT category_id, SUM(amount) as total, COUNT(*) as count
FROM transactions t
JOIN transaction_categories tc ON t.id = tc.transaction_id
WHERE tc.user_id = ? AND DATE_TRUNC('month', t.date) = ?
GROUP BY category_id;
```

### Find low-confidence categorizations
```sql
SELECT * FROM transaction_categories
WHERE user_id = ? AND confidence < 0.75 AND method = 'llm'
ORDER BY assigned_at DESC;
```

### Get merchant dictionary for caching
```sql
SELECT merchant_name, category_id, confidence, match_type
FROM merchant_dictionary
ORDER BY frequency DESC
LIMIT 500;
```

---

## Migration Strategy

**Phase 2 Implementation**:
1. Create categories table (10 predefined rows)
2. Create merchant_dictionary table with initial 500 entries
3. Create transaction_categories table with index
4. Create category_stats table
5. Add `category_id` field to transactions table (optional, for denormalization if needed)

**No Breaking Changes**: Phase 1 transactions table unmodified; categorization is purely additive.
