# Implementation Plan: Statement Import

**Branch**: `001-statement-import` | **Date**: 2026-07-05 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `specs/001-statement-import/spec.md`

**Note**: This plan is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Statement Import is Phase 1 of MoneyPlan AI. Users upload bank statements (PDF/CSV/Excel) from Indian banks. The system extracts transaction data (date, merchant, amount, debit/credit, balance, description), validates it, displays a preview, and persists to database. This feature unblocks all downstream features (categorization, budgeting, analytics, AI insights) by providing clean transaction data. 

**MVP Scope (Phase 1)**: Single-bank uploads, transaction extraction, validation, and preview confirmation (User Story 1: Upload & Parse). 
**Future Phases**: Multi-bank normalization (US2, P2) and duplicate detection (US3, P3) follow MVP completion.

## Technical Context

**Language/Version**: Go 1.25+ (backend); React/Next.js with TypeScript (frontend)

**Primary Dependencies**: 
- Backend: `pdfplumber` or `go-pdf` (PDF parsing), `gocsv` (CSV parsing), `xlsx` (Excel parsing), `sqlc` (type-safe SQL)
- Frontend: `react-dropzone` (file upload), `react-table` (preview grid), `zod` or `yup` (validation)

**Storage**: PostgreSQL 14+ (ACID transactions, JSON columns for flexible statement metadata)

**Testing**: Go `testing` package with `testify/require` (backend); `vitest` + `React Testing Library` (frontend); contract tests validate upload API schema

**Target Platform**: Web service (backend REST API on Linux); Browser client (React web app)

**Project Type**: Web service with frontend + backend

**Performance Goals**: 
- Upload-to-preview: <10 seconds for 50MB statement file
- Extraction accuracy: ≥95% (transaction count matches manual verification)
- Concurrent uploads: Handle 10+ simultaneous statement imports without degradation

**Constraints**: 
- File size: ≤50MB per statement (typical bank exports: 1-5MB for 12 months)
- **MVP Phase 1 (US1)**: No duplicate detection (deferred to US3, Phase 5)
- **Future Phase (US3)**: Duplicate detection must be deterministic over 12-month windows, <1s query time
- Security: Encrypt transactions at rest (PostgreSQL SSL), validate all file inputs (no zip bombs, script injection)

**Scale/Scope**: MVP; single user session (no distributed processing yet); ~1000 transactions per typical statement

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Alignment with MoneyPlan AI Constitution

| Principle | Status | Rationale |
|-----------|--------|-----------|
| **I. Data Privacy & Security** | ✅ PASS | Transactions are PII; plan includes encrypted storage (PostgreSQL SSL), input validation against injection, audit logging of all imports |
| **II. Modular Service Architecture** | ✅ PASS | Statement Import Service is first domain service with clear REST API contract; no direct database access from frontend; service owned by Transaction domain |
| **III. Data Quality First** | ✅ PASS | Core to feature: validation before persist (date format, numeric amounts), extraction accuracy tested against real bank formats, duplicate detection for data integrity |
| **IV. API Contract Testing** | ✅ PASS | Upload/Extract/Preview endpoints will have contract tests; external dependencies (PDF parser) mocked in unit tests |
| **V. AI Grounding & Explainability** | ⓘ N/A | Not applicable for Phase 1 (import is deterministic); applies to later phases (categorization, insights) |

**Gate Status**: ✅ PASS — No constitution violations. Feature aligns with all applicable principles.

## Project Structure

### Documentation (this feature)

```text
specs/001-statement-import/
├── plan.md              # This file
├── research.md          # Phase 0 output (PDF parsing approaches, bank format analysis)
├── data-model.md        # Phase 1 output (Transaction, Statement, ImportJob entities)
├── contracts/           # Phase 1 output (upload, extract, preview API schemas)
│   ├── upload-endpoint.md
│   ├── extract-endpoint.md
│   └── preview-endpoint.md
├── quickstart.md        # Phase 1 output (validation guide for feature)
└── tasks.md             # Phase 2 output (/speckit-tasks command)
```

### Source Code (repository root)

**Selected Structure**: Web application with separate backend and frontend

```text
backend/
├── cmd/
│   └── statement-import-api/
│       └── main.go          # HTTP server entry point
├── internal/
│   ├── statement/           # Statement Import Service
│   │   ├── parser.go        # PDF/CSV/Excel parsing
│   │   ├── validator.go     # Transaction validation
│   │   ├── duplicate.go     # Duplicate detection
│   │   ├── models.go        # Domain models (Transaction, Statement, ImportJob)
│   │   └── service.go       # Business logic
│   ├── api/
│   │   ├── upload.go        # POST /api/statements/upload
│   │   ├── extract.go       # POST /api/statements/extract
│   │   └── preview.go       # GET /api/statements/{id}/preview
│   └── db/
│       ├── migrations/      # SQL migrations for Transaction, Statement, ImportJob tables
│       └── queries/         # sqlc-generated type-safe queries
├── tests/
│   ├── contract/           # API contract tests
│   │   └── upload_test.go
│   ├── integration/        # Service integration tests
│   │   ├── parser_test.go
│   │   └── duplicate_test.go
│   └── unit/              # Business logic unit tests
│       ├── validator_test.go
│       └── models_test.go
└── go.mod

frontend/
├── src/
│   ├── pages/
│   │   └── statements/
│   │       ├── UploadPage.tsx
│   │       ├── PreviewModal.tsx
│   │       └── HistoryPage.tsx
│   ├── components/
│   │   ├── FileDropZone.tsx
│   │   ├── TransactionPreview.tsx
│   │   └── UploadHistory.tsx
│   ├── services/
│   │   └── statementApi.ts  # API client for Statement Import endpoints
│   └── hooks/
│       └── useStatementUpload.ts
├── tests/
│   └── features/
│       ├── upload.test.tsx
│       └── preview.test.tsx
└── package.json
```

**Structure Decision**: Web application with backend Go REST API and React frontend. Statement Import Service is the first backend microservice, isolated in `internal/statement/`. Modular structure allows future services (Categorization, Budget, AI Insight) to be added alongside. Frontend components colocate with feature pages (statements/) for clarity.

## Complexity Tracking

> **No constitution violations; no complexity exceptions needed.**

---

## Next Steps

**Phase 0 (Research)**: Investigate PDF parsing libraries, Indian bank statement formats, duplicate detection algorithms.

**Phase 1 (Design)**: Define data models, API contracts, validation rules, quickstart guide.

**Phase 2 (Tasks)**: Generate task breakdown for TDD-first implementation across backend/frontend, contract tests, integration tests.
