---
description: "Task list for statement import feature implementation"
---

# Tasks: Statement Import

**Input**: Design documents from `specs/001-statement-import/`

**Prerequisites**: plan.md ✅, spec.md ✅, research.md ✅, data-model.md ✅, contracts/ ✅

**Organization**: Tasks grouped by user story (US1, US2, US3) to enable independent implementation and delivery of each story.

## Format: `[ID] [P?] [Story] Description`

- **[P]**: Can run in parallel (different files, no dependencies)
- **[Story]**: Which user story this task belongs to (US1, US2, US3, or none for setup/foundation)
- Include exact file paths in descriptions

## Path Conventions

- **Backend**: `backend/` at repository root
- **Frontend**: `frontend/` at repository root
- Paths shown assume web app structure (from plan.md)

---

## Phase 1: Setup (Shared Infrastructure)

**Purpose**: Project initialization and basic structure

- [x] T001 Initialize Go backend module and dependencies (`backend/go.mod`)
- [x] T002 Initialize React/Next.js frontend project (`frontend/package.json`)
- [x] T003 [P] Set up backend project structure: `backend/cmd/`, `backend/internal/`, `backend/tests/`
- [x] T004 [P] Set up frontend project structure: `frontend/src/pages/statements/`, `frontend/src/components/`, `frontend/src/services/`
- [x] T005 [P] Configure linting and formatting (Go: `gofmt`, `golangci-lint`; Frontend: `eslint`, `prettier`)
- [x] T006 [P] Set up CI/CD pipeline (GitHub Actions for tests, linting, type-checking)

---

## Phase 2: Foundational (Blocking Prerequisites)

**Purpose**: Core infrastructure that MUST be complete before ANY user story can be implemented

**⚠️ CRITICAL**: No user story work can begin until this phase is complete

- [x] T007 Create PostgreSQL database schema with `statements`, `transactions`, `import_jobs` tables (from data-model.md)
- [x] T008 Create database migration framework and migration runner (`backend/internal/db/migrations/`)
- [x] T009 [P] Implement JWT authentication middleware in `backend/internal/api/middleware/auth.go`
- [x] T010 [P] Implement error handling and logging middleware in `backend/internal/api/middleware/errors.go`
- [x] T011 [P] Set up HTTP router and base API structure in `backend/cmd/statement-import-api/main.go`
- [ ] T012 Create database query builder/ORM config using `sqlc` in `backend/internal/db/queries/`
- [x] T013 [P] Implement transaction validator helper functions in `backend/internal/statement/validator.go` (reusable across stories)
- [x] T014 [P] Set up React API client base service in `frontend/src/services/api.ts`
- [x] T015 [P] Create React auth context and JWT token management in `frontend/src/contexts/AuthContext.tsx`

**Checkpoint**: Foundation ready - user story implementation can now begin in parallel

---

## Phase 3: User Story 1 - Upload & Parse Bank Statement (Priority: P1) 🎯 MVP

**Goal**: Users can upload bank statements (PDF/CSV/Excel), system extracts transactions, displays preview for confirmation.

**Independent Test**: User can upload valid statement file, view extracted transactions in preview without needing other stories.

**Value**: Unblocks all downstream features by providing transaction data.

### Tests for User Story 1 (Contract Tests - TDD)

> **NOTE: Write these tests FIRST, ensure they FAIL before implementation**

- [ ] T016 [P] [US1] Contract test for upload endpoint (202 Accepted, PENDING status) in `backend/tests/contract/upload_test.go`
- [ ] T017 [P] [US1] Contract test for preview endpoint (transactions array, validation_summary) in `backend/tests/contract/preview_test.go`
- [ ] T018 [P] [US1] Contract test for confirm endpoint (persist transactions, SUCCESS status) in `backend/tests/contract/confirm_test.go`
- [ ] T019 [P] [US1] Integration test for PDF parsing (extract HDFC statement format) in `backend/tests/integration/pdf_parser_test.go`
- [ ] T020 [P] [US1] Integration test for CSV parsing (extract standard bank CSV) in `backend/tests/integration/csv_parser_test.go`
- [ ] T021 [P] [US1] Unit tests for transaction validator (date format, amount, merchant) in `backend/tests/unit/validator_test.go`

### Implementation for User Story 1

#### Models & Data Layer

- [ ] T022 [P] [US1] Create Transaction model and repository in `backend/internal/statement/models.go`
- [ ] T023 [P] [US1] Create Statement model and repository in `backend/internal/statement/models.go`
- [ ] T024 [P] [US1] Create ImportJob model and repository in `backend/internal/statement/models.go`
- [ ] T025 [US1] Implement sqlc queries for transactions (insert, select by statement) in `backend/internal/db/queries/transactions.sql`
- [ ] T026 [US1] Implement sqlc queries for statements (insert, select by user) in `backend/internal/db/queries/statements.sql`

#### Parsing & Validation Layer

- [ ] T027 [P] [US1] Implement PDF parser using `pdfplumber` library in `backend/internal/statement/pdf_parser.go` (extract table structure)
- [ ] T028 [P] [US1] Implement CSV parser using `gocsv` in `backend/internal/statement/csv_parser.go`
- [ ] T029 [P] [US1] Implement Excel parser using `excelize` in `backend/internal/statement/excel_parser.go`
- [ ] T030 [P] [US1] Implement transaction validator in `backend/internal/statement/validator.go` (date, amount, merchant, type validation)
- [ ] T031 [US1] Implement HDFC format configuration in `backend/internal/statement/formats/hdfc.go` (column mapping)
- [ ] T032 [US1] Implement ICICI format configuration in `backend/internal/statement/formats/icici.go` (column mapping)
- [ ] T033 [P] [US1] Implement statement metadata extractor (period_start, period_end) in `backend/internal/statement/metadata.go`

#### Service Layer

- [ ] T034 [US1] Implement StatementService.Upload() (validate file, create Statement record) in `backend/internal/statement/service.go`
- [ ] T035 [US1] Implement StatementService.ExtractTransactions() (parse file, extract data) in `backend/internal/statement/service.go`
- [ ] T036 [US1] Implement StatementService.PreviewTransactions() (return extracted data with validation summary) in `backend/internal/statement/service.go`
- [ ] T037 [US1] Implement StatementService.ConfirmImport() (validate, persist transactions) in `backend/internal/statement/service.go`
- [ ] T038 [US1] Implement file hash computation (SHA-256) for duplicate detection in `backend/internal/statement/service.go`
- [ ] T039 [P] [US1] Implement async job queue for statement processing in `backend/internal/jobs/statement_queue.go` (background processing)

#### API Layer

- [ ] T040 [US1] Implement POST /api/statements/upload endpoint in `backend/internal/api/upload.go` (file upload, validation, queue job)
- [ ] T041 [US1] Implement GET /api/statements/{id}/preview endpoint in `backend/internal/api/preview.go` (return extracted transactions)
- [ ] T042 [US1] Implement POST /api/statements/{id}/confirm endpoint in `backend/internal/api/confirm.go` (persist to DB)
- [ ] T043 [P] [US1] Implement error response handling for file validation errors in `backend/internal/api/errors.go`

#### Frontend

- [ ] T044 [P] [US1] Create file upload component (drag-and-drop) in `frontend/src/components/FileDropZone.tsx`
- [ ] T045 [P] [US1] Create bank code selector component in `frontend/src/components/BankSelector.tsx`
- [ ] T046 [US1] Create upload form page in `frontend/src/pages/statements/UploadPage.tsx` (combines file drop, bank selector, submit)
- [ ] T047 [US1] Implement file upload API call in `frontend/src/services/statementApi.ts` (POST /api/statements/upload)
- [ ] T048 [P] [US1] Create transaction preview table component in `frontend/src/components/TransactionPreview.tsx` (paginated grid)
- [ ] T049 [P] [US1] Create validation summary component in `frontend/src/components/ValidationSummary.tsx` (error count, error details)
- [ ] T050 [US1] Create preview page in `frontend/src/pages/statements/PreviewPage.tsx` (shows extracted transactions, confirm/cancel buttons)
- [ ] T051 [US1] Implement preview data fetch and polling in `frontend/src/hooks/useStatementPreview.ts` (GET /api/statements/{id}/preview)
- [ ] T052 [US1] Implement confirm import handler in `frontend/src/pages/statements/PreviewPage.tsx` (POST /api/statements/{id}/confirm)
- [ ] T053 [P] [US1] Add upload progress indicator component in `frontend/src/components/UploadProgress.tsx`
- [ ] T054 [US1] Integrate upload flow: UploadPage → Preview → Confirmation in `frontend/src/pages/statements/index.tsx`

#### Integration & Testing

- [ ] T055 [US1] Test end-to-end upload flow with sample HDFC PDF in quickstart scenario 1
- [ ] T056 [US1] Test end-to-end upload flow with sample ICICI CSV in quickstart scenario 4
- [ ] T057 [US1] Verify extracted transactions match manual count (95% accuracy) in quickstart scenario 7
- [ ] T058 [US1] Verify upload-to-preview latency <10 seconds in quickstart scenario 8

**Checkpoint**: At this point, User Story 1 should be fully functional and testable independently.

---

## Phase 4: User Story 2 - Multi-Bank Support (Priority: P2)

**Goal**: Support multiple banks with different statement formats; merge transactions from different banks into unified view.

**Independent Test**: User can upload HDFC statement, then ICICI statement, and query all transactions together from both banks.

**Value**: Enables users with multiple accounts to see complete financial picture.

### Tests for User Story 2 (Contract Tests)

- [ ] T059 [P] [US2] Contract test for multi-bank statement queries (list transactions across banks) in `backend/tests/contract/multi_bank_test.go`
- [ ] T060 [P] [US2] Integration test for merging HDFC + ICICI statements in `backend/tests/integration/multi_bank_test.go`
- [ ] T061 [P] [US2] Integration test for overlapping date ranges (US2 acceptance scenario) in `backend/tests/integration/overlap_test.go`

### Implementation for User Story 2

#### Models & Data Layer

- [ ] T062 [US2] Add bank-aware indexing to transactions table (speed up cross-bank queries) in `backend/internal/db/queries/transactions.sql`
- [ ] T063 [P] [US2] Implement transaction query service (list across multiple banks, filter by date range) in `backend/internal/statement/query_service.go`

#### Format Support

- [ ] T064 [US2] Implement Axis format configuration in `backend/internal/statement/formats/axis.go` (column mapping)
- [ ] T065 [US2] Implement SBI format configuration in `backend/internal/statement/formats/sbi.go` (column mapping)
- [ ] T066 [P] [US2] Create bank format registry and auto-detection logic in `backend/internal/statement/formats/registry.go`

#### Service Layer

- [ ] T067 [US2] Extend StatementService to detect and apply bank-specific format in `backend/internal/statement/service.go`
- [ ] T068 [US2] Implement unified transaction query service in `backend/internal/statement/service.go` (no bank boundaries)

#### API Layer

- [ ] T069 [US2] Implement GET /api/transactions (list transactions across all user's banks) in `backend/internal/api/transactions.go`
- [ ] T070 [P] [US2] Add query parameters (bank_code, date_range filters) to transactions endpoint in `backend/internal/api/transactions.go`

#### Frontend

- [ ] T071 [P] [US2] Create bank filter component in `frontend/src/components/BankFilter.tsx` (select multiple banks)
- [ ] T072 [P] [US2] Create date range filter component in `frontend/src/components/DateRangeFilter.tsx`
- [ ] T073 [US2] Create transactions list page in `frontend/src/pages/statements/TransactionsPage.tsx` (show all transactions, filtered by bank/date)
- [ ] T074 [US2] Implement unified transactions API call in `frontend/src/services/statementApi.ts` (GET /api/transactions with filters)
- [ ] T075 [US2] Add multi-bank support to routing in `frontend/src/pages/statements/index.tsx` (transactions list as main view)

#### Integration & Testing

- [ ] T076 [US2] Test uploading multiple bank statements and viewing unified list in quickstart scenario 4
- [ ] T077 [US2] Test overlapping date ranges from different banks in quickstart scenario 4

**Checkpoint**: At this point, User Stories 1 AND 2 should both work independently.

---

## Phase 5: User Story 3 - Upload History & Duplicate Detection (Priority: P3)

**Goal**: Show user's upload history and prevent duplicate statement imports.

**Independent Test**: Upload same statement twice; system warns and prevents second import.

**Value**: Prevents accidental re-imports, provides audit trail of uploads.

### Tests for User Story 3 (Contract Tests)

- [ ] T078 [P] [US3] Contract test for duplicate file detection (409 Conflict) in `backend/tests/contract/duplicate_test.go`
- [ ] T079 [P] [US3] Integration test for duplicate detection logic in `backend/tests/integration/duplicate_logic_test.go`
- [ ] T080 [P] [US3] Contract test for upload history endpoint in `backend/tests/contract/history_test.go`

### Implementation for User Story 3

#### Models & Data Layer

- [ ] T081 [US3] Add file_hash and duplicate detection query to Statement model in `backend/internal/statement/models.go`
- [ ] T082 [US3] Create index on (bank_code, account_number_hash, statement_period_start) for overlap detection in `backend/internal/db/queries/statements.sql`

#### Service Layer

- [ ] T083 [US3] Implement duplicate detection logic in StatementService.CheckForDuplicates() in `backend/internal/statement/service.go`
- [ ] T084 [US3] Implement overlapping period detection in StatementService.DetectOverlappingStatements() in `backend/internal/statement/service.go`
- [ ] T085 [US3] Implement upload history query in StatementService.ListStatements() in `backend/internal/statement/service.go`

#### API Layer

- [ ] T086 [US3] Update POST /api/statements/upload to check for duplicates and return 409 if found in `backend/internal/api/upload.go`
- [ ] T087 [US3] Implement GET /api/statements (user's upload history) in `backend/internal/api/statements.go`
- [ ] T088 [P] [US3] Add query parameters (date_range, bank_code filters) to statements endpoint in `backend/internal/api/statements.go`

#### Frontend

- [ ] T089 [P] [US3] Create statement history table component in `frontend/src/components/StatementHistory.tsx` (list of uploaded statements with dates, transaction counts)
- [ ] T090 [US3] Create history page in `frontend/src/pages/statements/HistoryPage.tsx` (shows all user's statement uploads)
- [ ] T091 [US3] Implement statement history API call in `frontend/src/services/statementApi.ts` (GET /api/statements)
- [ ] T092 [P] [US3] Add duplicate warning modal component in `frontend/src/components/DuplicateWarning.tsx` (shows existing statement details)
- [ ] T093 [US3] Integrate upload flow to show duplicate warning on conflict (409 response) in `frontend/src/pages/statements/UploadPage.tsx`
- [ ] T094 [US3] Add history view to statement navigation in `frontend/src/pages/statements/index.tsx`

#### Integration & Testing

- [ ] T095 [US3] Test duplicate detection in quickstart scenario 3
- [ ] T096 [US3] Test uploading multiple statements and viewing upload history

**Checkpoint**: All user stories should now be independently functional.

---

## Phase 6: Polish & Cross-Cutting Concerns

**Purpose**: Improvements that affect multiple user stories; final validation and documentation.

- [ ] T097 [P] Run full end-to-end test suite (all quickstart scenarios 1-8)
- [ ] T098 [P] Verify extraction accuracy ≥95% on test statements (success criterion SC-002)
- [ ] T099 [P] Verify upload-to-preview latency <10 seconds (success criterion SC-001)
- [ ] T100 [P] Load testing: concurrent uploads (stress test 10+ simultaneous uploads)
- [ ] T101 [P] Add debug logging for statement processing pipeline (aids troubleshooting)
- [ ] T102 [P] Documentation: API documentation (README for `/api/statements` endpoints)
- [ ] T103 [P] Documentation: Database schema documentation in `docs/database.md`
- [ ] T104 [P] Code review checklist: Constitution compliance (Data Privacy, Modular Architecture, Data Quality, API Contracts)
- [ ] T105 Run full quickstart validation guide (`quickstart.md` scenarios 1-8)
- [ ] T106 Deploy to staging environment and verify against production data volume

---

## Dependencies & Execution Order

### Phase Dependencies

- **Setup (Phase 1)**: No dependencies - can start immediately ✓
- **Foundational (Phase 2)**: Depends on Setup completion - BLOCKS all user stories ✓
- **User Stories (Phase 3+)**: All depend on Foundational phase completion
  - User stories can proceed in parallel (if staffed) OR sequentially in priority order (P1 → P2 → P3)
  - US1 and US2 are **independently testable** (don't depend on each other)
  - US3 builds on US1/US2 but is independently testable
- **Polish (Final Phase)**: Depends on all desired user stories being complete

### User Story Dependencies

```
Foundational Phase (T007-T015)
    ↓
US1: Upload & Parse (T016-T058) ✓ Independent MVP
    ├─ Can ship alone and deliver value
    └─ Unblocks all downstream features
    ↓
US2: Multi-Bank Support (T059-T077) ✓ Independent enhancement
    ├─ Builds on US1 data layer (not code dependency)
    └─ Independently queryable (doesn't require US3)
    ↓
US3: Duplicate Detection & History (T078-T096) ✓ Independent UX improvement
    └─ Independent from US2; can be done in parallel or after
    ↓
Polish Phase (T097-T106)
```

### Within Each User Story

1. **Tests MUST be written and FAIL before implementation**
   - Contract tests (T016-T021 for US1, etc.)
   - Write failing test code first
2. **Models before Services** (T022-T026, then T034-T039)
3. **Services before Endpoints** (service layer first, then API)
4. **Endpoints before Frontend** (backend must be testable via curl)
5. **Core implementation before integration** (individual pieces tested before combined)
6. **Story complete and independently testable before moving to next**

### Parallel Opportunities

#### Within Phase 1 (Setup)
- Tasks T003-T006 can run in parallel (different projects/configs)

#### Within Phase 2 (Foundational)
- Tasks T009-T015 can run in parallel (middleware, configs, setup)
- Task T007-T008 must complete before parallel tasks

#### Within Phase 3 (US1)
- **Tests (T016-T021)**: All [P] tests can run in parallel
- **Models (T022-T026)**: Models T022-T024 can run in parallel, T025-T026 depend on models
- **Parsers (T027-T033)**: All [P] parsers can run in parallel
- **Service Layer (T034-T039)**: Parsers must complete first, then services in parallel
- **API (T040-T043)**: Can run in parallel once service layer complete
- **Frontend (T044-T054)**: Can run in parallel with backend implementation

**Parallel Example: US1 Parser Tasks**
```bash
# All 3 parsers run concurrently (different files):
T027: PDF parser (backend/internal/statement/pdf_parser.go)
T028: CSV parser (backend/internal/statement/csv_parser.go)
T029: Excel parser (backend/internal/statement/excel_parser.go)

# Once parsers complete, format configs run in parallel:
T031: HDFC format (backend/internal/statement/formats/hdfc.go)
T032: ICICI format (backend/internal/statement/formats/icici.go)
```

#### Within Phase 4 (US2)
- Tasks T062-T066 can mostly run in parallel (new format configs)

#### Within Phase 6 (Polish)
- All [P] tasks (T097-T104) can run in parallel

---

## Implementation Strategy

### MVP First (User Story 1 Only) - Recommended Start

1. **Complete Phase 1: Setup** (6 tasks) — Estimated: 4-6 hours
   - Initialize projects, dependencies, structure

2. **Complete Phase 2: Foundational** (9 tasks) — Estimated: 8-12 hours
   - Database, auth, routing, error handling

3. **Complete Phase 3: US1 (P1)** (42 tasks) — Estimated: 40-60 hours
   - Full upload → parse → preview → confirm flow
   - PDF, CSV, Excel parsing
   - Transaction validation
   - Backend API + React frontend

   **MVP SHIP POINT**: Users can upload statements, see extracted transactions, confirm import. All downstream features (categorization, budgets, insights) now possible.

   **Total MVP time**: 52-78 hours (1-2 weeks with 1 developer)

### Incremental Delivery

1. **Week 1**: Phases 1-2 + US1 (Setup + Foundation + Upload/Parse/Preview) → **MVP ship**
2. **Week 2**: US2 (Multi-Bank Support) → Users with multiple accounts can consolidate
3. **Week 3**: US3 (Duplicate Detection) + Polish → Robust feature-complete

### Parallel Team Strategy

With 3 developers:
1. **Developer A**: Phases 1-2 (infrastructure) while others wait
2. Once infrastructure ready:
   - **Developer A**: US1 Backend (parsers, service, API)
   - **Developer B**: US1 Frontend (upload, preview, confirm pages)
   - **Developer C**: US2 (start after US1 service layer done)
3. Once US1 + US2 done:
   - **Developer A**: US3
   - **Developers B+C**: Polish & testing

---

## Notes

- [P] tasks = different files, no dependencies between them
- [Story] label = maps task to specific user story (US1, US2, US3)
- Each user story should be independently completable and testable
- Verify tests FAIL before implementing (TDD discipline per constitution)
- Commit after each task or logical group
- Stop at any checkpoint to validate story independently
- Constitution compliance checks in Phase 6 (Data Privacy, Modular Services, Data Quality First, API Contracts)

