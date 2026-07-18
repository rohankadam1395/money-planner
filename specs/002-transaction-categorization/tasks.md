# Tasks: Transaction Categorization with Pluggable LLM Providers

**Feature**: 002-Transaction-Categorization | **Branch**: `002-transaction-categorization` | **Dates**: 2026-07-18

**Input**: Design documents from `/specs/002-transaction-categorization/`

**Prerequisites**: plan.md, spec.md, research.md, data-model.md, contracts/, quickstart.md

**Organization**: Tasks grouped by user story (P1, P2, P3) for independent implementation and delivery

---

## Implementation Strategy

**MVP Scope (Phases 1-3)**: User Story 1 only - Rule-based categorization with merchant dictionary  
**Phase 4+**: User Story 2 - LLM categorization (Ollama, Claude, OpenAI)  
**Phase 5+**: User Story 3 - Analytics & recategorization  
**Full Feature**: Complete all three user stories + pluggable LLM provider abstraction (Phases 1-6)

**Parallel Opportunities**:
- Phase 3 (US1) can run parallel with Phase 2 foundation setup once database schema is complete
- US1 and US2 implementations can run in parallel once LLMProvider interface is defined
- Frontend components for all stories can be developed in parallel

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization, directory structure, and basic tooling

- [X] T001 Create categorization service directories: `backend/internal/categorization/`, `backend/internal/config/`, `frontend/src/pages/categories/`
- [X] T002 [P] Add dependencies to `backend/go.mod`: `github.com/spf13/viper` (config), `github.com/stretchr/testify` (testing)
- [X] T003 [P] Update `frontend/package.json`: Ensure `vitest`, React Testing Library installed
- [X] T004 [P] Configure linting: Update `.golangci.yml` for Go backend, `.eslintrc.json` for frontend
- [X] T005 Initialize git branch: `git checkout -b 002-transaction-categorization`

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Database schema and configuration management - MUST COMPLETE before user story implementation

**⚠️ CRITICAL**: No user story implementation can begin until this phase finishes

**MVP Note**: LLM provider abstraction deferred to Phase 4+. Phase 2 focuses on database, merchant dictionary, and rule-based categorization foundation.

### Database & Entities

- [X] T006 Create database migration: `backend/db/migrations/004_create_categories_table.sql` with categories table (ID, name, description, color, icon, is_predefined)
- [X] T007 Create database migration: `backend/db/migrations/005_create_merchant_dictionary_table.sql` with merchant_dictionary table (ID, merchant_name, category_id, source, confidence, match_type, frequency)
- [X] T008 Create database migration: `backend/db/migrations/006_create_transaction_categories_table.sql` with transaction_categories table (ID, transaction_id, category_id, method, llm_provider, confidence, assigned_at)
- [X] T009 Create database migration: `backend/db/migrations/007_create_category_stats_table.sql` with category_stats table (ID, user_id, category_id, period, total_spent, transaction_count, avg_transaction)
- [X] T010 [P] Generate sqlc stubs: Run `sqlc generate` for new tables in `backend/db/queries/`

### LLM Provider Abstraction (Phase 4+, Deferred)

**Skipped in Phase 2 MVP**. These tasks deferred to Phase 4+ for LLM categorization:
- T011-T015 (LLMProvider interface, Ollama, Claude, mock providers)

### Configuration Management (Phase 2 MVP)

- [X] T016 [P] Create merchant dictionary config in `backend/internal/config/merchants_config.go`: Trie settings, cache size, fuzzy match threshold (0.85)
- [X] T017 Create config loading in `backend/internal/config/loader.go`: Load merchant dictionary on startup, validate structure
- [X] T018 [P] Add environment variable documentation: Update `backend/.env.example` with MERCHANT_DICT_SIZE, FUZZY_MATCH_THRESHOLD

### Core Categorization Logic

- [X] T021 Create categorization service in `backend/internal/categorization/service.go`: CategorizationService struct, constructor with provider + db injection, Categorize method stub
- [X] T022 Create merchant dictionary cache in `backend/internal/categorization/merchant_dict.go`: Trie-based lookup, cache loading on startup, exact/fuzzy matching logic (Levenshtein distance ≥85%)
- [X] T023 Create confidence scoring in `backend/internal/categorization/confidence.go`: Score mapping - exact match (1.0), fuzzy (0.85-0.99 by Levenshtein distance), uncategorized (0.0). **Acceptance**: Pass contract tests verifying scoring logic for known/fuzzy/unknown merchants

**Checkpoint**: Foundation complete - LLM provider abstraction is pluggable, config is flexible, categorization service skeleton ready. User story work can now begin in parallel.

---

## Phase 3: User Story 1 - Rule-Based Transaction Categorization (Priority: P1) 🎯 MVP

**Goal**: Transactions with known merchants are automatically categorized during import preview with 100% confidence

**Independent Test**: Upload statement with Swiggy/Amazon/Uber transactions → preview shows correct categories (Food/Shopping/Transport) → confirm import → categories persisted to database

### Implementation for User Story 1

- [X] T024 [P] [US1] Create Category model in `backend/internal/categorization/models.go`: Category struct with ID, name, description, color, icon
- [X] T025 [P] [US1] Create MerchantDictionary model in `backend/internal/categorization/models.go`: MerchantDictionary struct with ID, merchant_name, category_id, confidence, match_type
- [X] T026 [P] [US1] Create TransactionCategory model in `backend/internal/categorization/models.go`: TransactionCategory struct with ID, transaction_id, category_id, method, llm_provider, confidence
- [X] T026-CATEGORIES [P] [US1] Create/review `specs/002-transaction-categorization/categories-reference.md` with all 10 predefined categories (names, descriptions, colors, icons, examples). Link from schema, frontend components.
- [X] T027 [US1] Seed merchant dictionary in `backend/db/seeds/merchant_dictionary_seed.sql`: Insert ≥500 entries (Swiggy→Food, Amazon→Shopping, Uber→Transport, etc.) for Indian banks. Reference `categories-reference.md` for category IDs.
- [X] T028 [US1] Implement rule-based categorization in `backend/internal/categorization/service.go`: CategorizeRule method using merchant dictionary exact + fuzzy matching (Levenshtein ≥85%), returns (category, confidence, explanation)
- [X] T029 [US1] Implement /api/v1/transactions/categorize endpoint in `backend/internal/api/categorize.go`: POST handler accepting transactions array, returns categorizations with method/confidence/explanation. Include stats (rule_based_count, fuzzy_count, uncategorized_count)
- [X] T031 [US1] Integrate categorization into statement import flow in `backend/internal/api/statements.go`: Call categorization service during preview, return categories with transactions

### Contract Tests for US1

- [X] T032 [P] [US1] Create contract test `backend/tests/contract/categorize_rule_based_test.go`: Known merchant (Swiggy) → correct category (Food) with confidence 1.0, method "rule_based"
- [X] T033 [P] [US1] Create contract test: Fuzzy match (SWIGGY FD) → Food with confidence 0.85-0.99, method "fuzzy"
- [X] T034 [P] [US1] Create contract test: Unknown merchant → "Uncategorized", method "none"
- [X] T035 [US1] Create contract test: Batch categorization (10 txns) → partial results, correct stats (rule_based count, total)

### Frontend for US1

- [X] T036 [P] [US1] Create CategoryBadge component in `frontend/src/components/CategoryBadge.tsx`: Display category name with color icon
- [X] T037 [P] [US1] Update TransactionPreview component in `frontend/src/pages/statements/PreviewModal.tsx`: Add category column showing badge + confidence + method
- [X] T038 [US1] Update StatementUpload flow in `frontend/src/pages/statements/UploadPage.tsx`: Call /api/categorize before preview, display categories to user

**Checkpoint**: User Story 1 complete and independently testable. Transactions automatically categorized by merchant dictionary. MVP feature ready for deployment.

---

## Phase 4: User Story 2 - LLM-Based Categorization for Unknown Merchants (Priority: P2, Deferred)

**Goal**: Unknown merchants are categorized by Ollama (default) or configured provider with confidence score and explanation

**Independent Test**: Upload statement with unknown merchant "Aashish Restaurant Pvt Ltd" → Ollama infers "Food" with confidence 0.80 → user can review and correct → LLM provider logged in response

### Implementation for User Story 2

- [ ] T039 [US2] Implement Ollama provider in `backend/internal/llm/ollama.go`: POST /api/generate with categorization prompt, parse response, handle errors gracefully
- [ ] T040 [US2] Implement LLM categorization in `backend/internal/categorization/service.go`: CategorizeLLM method calling provider, confidence mapping (Ollama: 0.65-0.80, Claude: 0.85-0.98)
- [ ] T041 [US2] Update categorization logic in `backend/internal/categorization/service.go`: Categorize method routes to CategorizeRule → if no match, try CategorizeLLM → if LLM fails, default to "Uncategorized"
- [ ] T042 [US2] Implement graceful degradation: If LLM provider unavailable, log error and return "Uncategorized" without blocking import (see FR-107)
- [ ] T043 [US2] Add llm_provider field to categorization response in `/api/v1/transactions/categorize`: Return "ollama", "claude", or null for method and llm_provider
- [ ] T044 [US2] Update stats in /api/categorize response: Include llm_providers map with provider counts (e.g., {"ollama": 5, "claude": 0})
- [ ] T045 [US2] Implement Claude provider in `backend/internal/llm/claude.go`: Use anthropic-sdk-go (if ANTHROPIC_API_KEY set), request/response mapping, error handling
- [ ] T046 [P] [US2] Create provider switching logic: Update main.go to select provider based on LLM_PROVIDER env var, log active provider on startup

### Contract Tests for US2

- [ ] T047 [P] [US2] Create contract test `backend/tests/contract/categorize_llm_test.go`: Unknown merchant → LLM-categorized with method "llm" and llm_provider "ollama"
- [ ] T048 [P] [US2] Create contract test: Mock Ollama timeout → transaction defaults to "Uncategorized" without blocking (206 Partial Content)
- [ ] T049 [P] [US2] Create contract test: Invalid LLM response → fallback to "Uncategorized" with error logged
- [ ] T050 [P] [US2] Create contract test: LLM_PROVIDER env var override → provider switches without code changes
- [ ] T051 [US2] Create integration test `backend/tests/integration/llm_categorization_test.go`: End-to-end with Ollama (or mock), verify confidence scores, verify provider tracking

### Frontend for US2

- [ ] T052 [P] [US2] Update TransactionPreview component: Show LLM-suggested category with llm_provider indicator (badge showing "Ollama" or "Claude")
- [ ] T053 [US2] Add LLM suggestion override modal in `frontend/src/pages/statements/PreviewModal.tsx`: Allow user to correct LLM category before confirming
- [ ] T054 [US2] Add confidence-based highlighting: Highlight low-confidence (< 0.75) LLM suggestions for user review
- [ ] T055 [US2] Update import confirmation: Show breakdown by categorization method (rule_based, llm, uncategorized counts)

**Checkpoint**: User Stories 1 AND 2 complete. Known merchants use fast rule-based, unknown merchants use Ollama/configurable LLM. Provider abstraction enables easy switching (env var only).

---

## Phase 5: User Story 3 - Category Management & Analytics (Priority: P3, Deferred)

**Goal**: Users view spending by category, drill down to transactions, and recategorize post-import

**Independent Test**: Import transactions → view category dashboard showing Food: ₹4500 (24 txns), Shopping: ₹8000 (15 txns) → click category → see all Food transactions with sort/filter → recategorize one → totals update

### Implementation for User Story 3

- [ ] T056 [P] [US3] Create CategoryStats model in `backend/internal/categorization/models.go`: CategoryStats struct with user_id, category_id, period, total_spent, transaction_count, avg_transaction
- [ ] T057 [US3] Implement category stats aggregation in `backend/internal/categorization/service.go`: UpdateCategoryStats method calculating totals per category per month (trigger on transaction confirm or recategorize)
- [ ] T058 [US3] Create /api/v1/categories endpoint in `backend/internal/api/categories.go`: GET handler returning all categories with stats (period parameter, include_stats query param)
- [ ] T059 [US3] Create /api/v1/categories/{id}/transactions endpoint in `backend/internal/api/categories.go`: GET handler returning transactions in category with sort/filter (date_start, date_end, limit, sort_by)
- [ ] T060 [US3] Implement /api/v1/transactions/{id}/recategorize endpoint in `backend/internal/api/recategorize.go`: POST handler accepting new category_id, updates transaction_categories, updates category_stats, returns old/new category info
- [ ] T061 [US3] Implement merchant dictionary learning in `backend/internal/categorization/service.go`: If learn_correction flag set, insert user's correction into merchant_dictionary with source "user_correction"
- [ ] T062 [US3] Add category queries in `backend/db/queries/categories.sql`: Queries for GetCategoryStats, GetTransactionsByCategory, UpdateCategoryStats

### Contract Tests for US3

- [ ] T063 [P] [US3] Create contract test `backend/tests/contract/categories_api_test.go`: GET /api/v1/categories → returns all 10 categories with stats (total_spent, transaction_count)
- [ ] T064 [P] [US3] Create contract test: GET /api/v1/categories/cat_food/transactions → returns Food transactions sorted by date
- [ ] T065 [P] [US3] Create contract test: POST /api/v1/transactions/{id}/recategorize → old/new category updated, stats recalculated
- [ ] T066 [P] [US3] Create contract test: Recategorization with learn_correction=true → new entry in merchant_dictionary
- [ ] T067 [US3] Create integration test `backend/tests/integration/category_analytics_test.go`: Full journey - import, categorize, view dashboard, recategorize, verify stats update

### Frontend for US3

- [ ] T068 [P] [US3] Create CategoryDashboard page in `frontend/src/pages/categories/CategoryDashboard.tsx`: Table/cards showing all categories with spending totals, transaction counts, month/year selector
- [ ] T069 [P] [US3] Create CategoryDetail page in `frontend/src/pages/categories/CategoryDetail.tsx`: List transactions in category, sort/filter by date/amount, show categorization method + confidence
- [ ] T070 [P] [US3] Create RecategorizeModal in `frontend/src/pages/categories/RecategorizeModal.tsx`: Dropdown to select new category, toggle "learn_correction" checkbox, submit
- [ ] T071 [US3] Add category navigation in `frontend/src/components/Navbar.tsx`: Link to categories dashboard from main nav
- [ ] T072 [P] [US3] Add category colors to UI: Use category.color field to display colored badges/bars in dashboard
- [ ] T073 [US3] Implement category stats refresh: After recategorization, update dashboard totals in real-time (refetch /api/categories endpoint)

**Checkpoint**: All three user stories complete. Full categorization feature with rule-based, LLM fallback, analytics, and recategorization working independently.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Testing, documentation, performance optimization, monitoring

### Testing & Validation

- [ ] T074 Run all contract tests: `cd backend && go test ./tests/contract/...`
- [ ] T075 Run all integration tests: `cd backend && go test ./tests/integration/...`
- [ ] T076 Create unit tests for merchant dictionary trie lookup in `backend/tests/unit/merchant_dict_test.go`
- [ ] T077 Create unit tests for confidence scoring in `backend/tests/unit/confidence_test.go`
- [ ] T078 [P] Create frontend unit tests: CategoryBadge, RecategorizeModal components in `frontend/src/__tests__/`
- [ ] T079 Create end-to-end test: Full import → categorize → view → recategorize flow in `backend/tests/e2e/categorization_e2e_test.go`
- [ ] T080 Run full test suite: `cd backend && go test ./... && cd ../frontend && npm run test`

### Documentation & Quickstart

- [ ] T081 Update quickstart.md: Verify all scenarios pass (Scenario 1 rule-based, Scenario 2 Ollama, Scenario 3 analytics)
- [ ] T082 Create deployment guide in `docs/OLLAMA_SETUP.md`: Docker setup, model pulling, environment variables
- [ ] T083 Create provider switching guide in `docs/PROVIDER_SWITCHING.md`: Instructions for switching from Ollama to Claude/OpenAI
- [ ] T084 Update README: Add transaction categorization feature overview, MVP scope, future roadmap

### Performance & Monitoring

- [ ] T085 Load test categorization: `apache-bench` or `wrk` with 1000 transactions, measure latency p50/p99
- [ ] T086 Optimize merchant dictionary cache: Measure trie lookup performance, add Redis layer if needed (p50 > 100ms)
- [ ] T087 Add categorization metrics: Log categorization method counts, LLM provider usage, error rates
- [ ] T088 Add Prometheus metrics in `backend/internal/metrics/metrics.go`: Counter for categorizations by method, gauge for category stats, histogram for latency
- [ ] T089 Create health check endpoint in `/api/v1/health`: Verify LLM provider connectivity, database connectivity

### Deployment & Rollout

- [ ] T090 Update CI/CD: Add database migration step before deployment
- [ ] T091 Create feature flag for categorization in config (default: enabled for MVP)
- [ ] T092 Document rollback plan: Steps to disable categorization or downgrade LLM provider if issues occur
- [ ] T093 Create runbook: Troubleshooting guide for common issues (LLM timeout, merchant dict mismatch, category stats inconsistency)

**Checkpoint**: Feature complete, tested, documented, and production-ready. All user stories delivered with pluggable LLM provider abstraction.

---

## Task Summary

| Phase | Title | Tasks | Dependencies | MVP |
|-------|-------|-------|--------------|-----|
| **1** | Setup | T001-T005 | None | ✅ |
| **2** | Foundational (DB + Config) | T006-T023, T026-CATEGORIES | Phase 1 | ✅ |
| **3** | US1: Rule-Based (P1 MVP) | T024-T029, T031-T038 | Phase 2 | ✅ |
| **4** | US2: LLM Categorization (P2, Deferred) | T011-T015, T039-T055 | Phase 2, 3 | ⏸️ |
| **5** | US3: Analytics & Recategorize (P3, Deferred) | T056-T073 | Phase 2, 3 | ⏸️ |
| **6** | Polish & Cross-Cutting | T074-T093 | Phase 3, 4, 5 | ⏸️ |

**Total Tasks**: 93  
**MVP Tasks** (Phases 1-3): ~38 (excludes T011-T015 LLM provider tasks deferred to Phase 4)  
**Full Feature**: 93 (Phases 1-6)

---

## Parallel Execution Plan

### Week 1 (Foundation + Setup)
- **Parallel**: T001-T005 (Setup) + T006-T023 (Foundational)
- Time: ~2-3 days
- Deliverable: LLM provider abstraction, database schema, config management, merchant dictionary cache

### Week 2 (User Story 1 - MVP)
- **Parallel**: T024-T031 (Models + endpoint + seeding) with T032-T035 (Contract tests)
- Time: ~2-3 days
- Deliverable: Rule-based categorization working end-to-end

### Week 3 (Frontend US1 + User Story 2)
- **Parallel**: T036-T038 (Frontend) with T039-T046 (LLM providers)
- Time: ~3 days
- Deliverable: UI showing categories, Ollama integration complete

### Week 4 (User Story 3 + Testing)
- **Parallel**: T056-T073 (Analytics) with T047-T051 (US2 tests)
- Time: ~3 days
- Deliverable: Full categorization feature

### Week 5 (Polish + Deployment)
- Sequential: T074-T093 (Tests, docs, monitoring, deployment)
- Time: ~2-3 days
- Deliverable: Production-ready feature

**Total Timeline**: ~4-5 weeks for full feature with testing and documentation

---

## MVP Scope (Phases 1-3, ~2 weeks)

To ship transaction categorization MVP in 1-2 weeks, implement:

1. **Phase 1**: Setup (T001-T005)
2. **Phase 2**: Foundational (T006-T023, T026-CATEGORIES; exclude T011-T015 LLM provider tasks)
3. **Phase 3**: User Story 1 only (T024-T029, T031-T038, rule-based categorization with merchant dictionary)

**What's NOT in MVP**:
- LLM categorization (Phase 4): Deferred, will be added in subsequent release
- Analytics & recategorization (Phase 5): Deferred
- Comprehensive testing (Phase 6): MVP includes contract tests only; integration/E2E deferred

**MVP Deliverable**: Users upload statements → transactions automatically categorized by merchant dictionary (≥500 merchants) → categories shown in preview with confidence scores → categories persisted → 10 predefined categories + contract tests passing

---

## Dependencies & Assumptions

**Assumptions**:
- Phase 1 (Statement Import) complete and working
- PostgreSQL 14+ available with existing transactions table
- Go 1.25+, React/Next.js with TypeScript available
- Merchant dictionary seeded with ≥500 Indian bank merchants (see `categories-reference.md`)

**Blocking Dependencies (MVP)**:
- Phase 2 (Foundational) MUST complete before Phase 3 (User Story 1)
- Database migrations (T006-T010) MUST run before any categorization logic
- Category definitions (T026-CATEGORIES) MUST be finalized before backend/frontend implementation starts

**Non-Blocking (MVP)**:
- Frontend (T036-T038) can develop in parallel with backend once API contracts defined
- Tests (T032-T035) can run against API stubs once endpoints defined

**Deferred Dependencies (Phase 4+)**:
- Ollama or other LLM provider installation (Phase 4)
- LLM provider abstraction (T011-T015, Phase 4)
- Admin interface for merchant updates (Phase 3+)

---

## Notes

- **LLM Provider Switching**: After Phase 2 foundation, switching from Ollama to Claude/OpenAI requires only environment variable change (`LLM_PROVIDER=claude`). No code changes.
- **Merchant Dictionary**: Seeded with 500+ Indian bank merchants (Swiggy, Amazon, etc.). User corrections feed back via T061 learning mechanism.
- **Confidence Scores**: Crucial for user trust. Always track and display categorization method and confidence so users can review/correct.
- **Error Handling**: Graceful degradation is critical (FR-107). If LLM fails, default to "Uncategorized" rather than blocking import.
- **Tests**: TDD recommended. Write contract tests (T032-T067) FIRST, then implement endpoints to pass them.

