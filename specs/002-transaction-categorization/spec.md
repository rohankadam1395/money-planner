# Feature Specification: Transaction Categorization

**Feature Branch**: `002-transaction-categorization`

**Created**: 2026-07-12

**Status**: Draft

**Input**: User description: "Phase 2 - Transaction Categorization"

## User Scenarios & Testing *(mandatory)*

### User Story 1 - Rule-Based Transaction Categorization (Priority: P1)

A user uploads bank statements and the system automatically categorizes each transaction into predefined categories based on merchant name matching. Known merchants (Swiggy→Food, Amazon→Shopping, Uber→Transport, etc.) are categorized instantly during import preview, allowing users to review categorization accuracy before confirming.

**Why this priority**: Core analytics feature; categorization enables spending analysis, budget tracking, and insights generation. Depends on Phase 1 (Statement Import).

**Independent Test**: User uploads statement with known merchants and all transactions display with correct category assignments without manual correction needed.

**Acceptance Scenarios**:

1. **Given** user uploads statement with merchants like "Swiggy", "Amazon", "Uber", **When** preview displays, **Then** each transaction shows appropriate category (Food, Shopping, Transport)
2. **Given** statement contains salary credit, **When** categorization runs, **Then** income transactions are identified and categorized as "Income"
3. **Given** user confirms import with categorized transactions, **When** data persists, **Then** categories are stored and retrievable

---

### User Story 2 - LLM-Based Categorization for Unknown Merchants (Priority: P2 - Phase 4+)

**Deferred to Phase 4+ (outside MVP)**. System will maintain a dictionary of known merchant→category mappings. For merchants not in the dictionary, system uses configurable LLM (default: Ollama running locally; supports future switching to other providers like Claude) to infer category based on merchant name and transaction amount. User reviews and approves LLM suggestions in preview before import.

**Why deferred**: Reduces MVP complexity; Phase 2 focuses on rule-based with merchant dictionary. Handles long-tail merchants post-MVP.

---

### User Story 3 - Category Management & Analytics (Priority: P3 - Phase 5+)

**Deferred to Phase 5+ (outside MVP)**. User will be able to view all transactions grouped by category, see category-level spending totals, and recategorize transactions post-import. System exposes category data for downstream features (budgets, insights, dashboard).

**Why deferred**: Reduces MVP scope; Phase 2 focuses on rule-based categorization in preview only. Analytics added post-MVP.

---

### Edge Cases (Phase 2 MVP)

- What if merchant name is empty or null? → Categorize as "Uncategorized"
- What if transaction amount is negative (credit/refund)? → If sender matches merchant dictionary, use that category; if not, check transaction type (credit→"Income", debit→normal flow)
- How are subscription services (Netflix, Spotify) categorized? → Entertainment (predefined, see `categories-reference.md`)
- What if rule-based matching has multiple possible categories? → Use highest-confidence match; if tie, default to first alphabetically

**Phase 4+ edge cases** (deferred):
- What if LLM categorization returns low confidence (<50%)—should it default to "Uncategorized"? → Yes (Phase 4+)
- Can users create custom categories? → No (Phase 2); custom categories deferred to Phase 3+

## Requirements *(mandatory)*

### Functional Requirements

#### Core Categorization (User Stories 1-3)

- **FR-101**: System MUST maintain a dictionary of merchant names → category mappings. Initial dictionary includes at least 500 major merchants (Swiggy, Amazon, Uber, Netflix, etc.) across 10 predefined categories
- **FR-102**: System MUST support 10 predefined categories: Food, Shopping, Transport, Housing, Utilities, Entertainment, Income, Healthcare, Education, Miscellaneous
- **FR-103**: System MUST categorize transactions using rule-based matching (exact and fuzzy merchant name matching) against the merchant dictionary
- **FR-104**: System MUST apply category during statement import preview, before user confirms; allows user to review and manually override category if needed
- **FR-105**: System MUST call configurable LLM (default: Ollama) for transactions with unknown merchants to infer category based on merchant name and amount. LLM provider is swappable via configuration (environment variables or config file) without code changes
- **FR-106**: System MUST only call LLM if rule-based matching fails (known merchant not found) to reduce API costs
- **FR-107**: System MUST handle LLM errors gracefully—if API unavailable or fails, categorize transaction as "Uncategorized" and allow manual categorization later
- **FR-108**: System MUST allow user to recategorize transactions post-import via UI and update category assignment in database
- **FR-109**: System MUST track categorization confidence score (rule-based=100%, LLM=score from API) for future filtering and learning
- **FR-110**: System MUST persist categories to database and expose category data via query API for downstream features (budget planning, dashboard, insights)
- **FR-111**: System MUST support future merchant dictionary updates via admin interface or automated learning from user corrections

#### Transaction Category History & Analytics

- **FR-112**: System MUST track category assignment timestamp and reason (rule-based, LLM, manual) for audit trail
- **FR-113**: System MUST support querying transactions by category with date range and amount range filters
- **FR-114**: System MUST calculate and expose category-level metrics: total spent, transaction count, average transaction, date range

### Key Entities

- **Category**: ID, Name (e.g., "Food"), Description, Color (for UI), Is Predefined (boolean)
- **MerchantDictionary**: ID, Merchant Name (or pattern), Category ID, Confidence (100 for rule-based), Created/Updated Timestamp, Source (rule-based, LLM, manual)
- **TransactionCategory**: Transaction ID, Category ID, Assignment Method (rule-based, LLM, manual), Confidence Score, Assigned Timestamp
- **CategoryStats**: User ID, Category ID, Period (month/year), Total Spent, Transaction Count, Average Transaction

## Success Criteria *(mandatory)*

### Measurable Outcomes

- **SC-101**: 90% of transactions from major Indian banks are automatically categorized without manual intervention
- **SC-102**: LLM-categorized transactions have ≥75% accuracy when spot-checked against user corrections
- **SC-103**: Categorization completes within preview latency budget (<10 seconds total import time including categorization)
- **SC-104**: User can recategorize a transaction and see updated category totals within 2 seconds
- **SC-105**: System supports at least 10 concurrent LLM API calls without degrading user import experience

## Assumptions *(mandatory)*

- **MVP Scope (Phase 2)**: Rule-based categorization only. LLM provider integration deferred to Phase 4+.
- Merchant dictionary can be maintained and updated via user corrections (learning); admin interface deferred to Phase 3+
- 10 predefined categories are fixed for Phase 2 (custom categories deferred to Phase 3+); see `categories-reference.md` for full list
- Transaction amount available and reliable (used by LLM in Phase 4+, not required for Phase 2)
- Initial merchant dictionary can be bootstrapped from public sources or manual curation (~500 Indian bank merchants)
- **Success Criteria Definitions**:
  - SC-101 "major Indian banks": Primary 5 banks (HDFC, ICICI, SBI, Axis, Kotak) representing ~70% of Indian banking market
  - SC-102 "≥75% accuracy": LLM accuracy validated against user corrections in preview (sample: ≥100 transactions post-import)
  - SC-104 "within 2 seconds": p99 latency from recategorize API call to UI update (includes network + server processing)
- LLM provider integration (Phase 4+): Provider available and accessible via HTTP API (Ollama local or Claude/OpenAI remote)

## Non-Functional Requirements

- LLM API calls must retry with exponential backoff on transient failures
- Categorization must not block statement import; if LLM slow, batch requests asynchronously
- Merchant dictionary lookup must use efficient string matching (trie or similar) for <10ms latency
- All categorization data must be queryable for downstream features (analytics, budget, insights)
