# Implementation Plan: Transaction Categorization

**Branch**: `002-transaction-categorization` | **Date**: 2026-07-12 | **Spec**: [spec.md](spec.md)

**Input**: Feature specification from `/specs/002-transaction-categorization/spec.md`

**Note**: This plan is filled in by the `/speckit-plan` command. See `.specify/templates/plan-template.md` for the execution workflow.

## Summary

Transaction Categorization is Phase 2 of MoneyPlan AI. After Statement Import (Phase 1) provides clean transaction data, this feature automatically categorizes transactions into 10 predefined categories using a three-tier strategy with pluggable LLM providers:

1. **Rule-Based (P1)**: Fast merchant dictionary lookup (500+ known merchants) during import preview (<100ms)
2. **LLM Fallback (P2)**: Configurable LLM provider (default: Ollama/Mistral 7B local; supports Claude, OpenAI, and future providers) for unknown merchants to infer category with confidence score (200-400ms)
3. **User Control & Analytics (P3)**: Post-import recategorization, category-level spending views, audit trails

**MVP Scope (Phases 1-3)**: Rule-based categorization with merchant dictionary. Phase 4+ adds LLM fallback (Ollama, Claude, OpenAI) and analytics. Minimal MVP unblocks downstream analytics, budgeting, and AI insights with zero external API cost.

## Technical Context

**Language/Version**: Go 1.25+ (backend); React/Next.js with TypeScript (frontend)

**Primary Dependencies**: 
- Backend: `spf13/viper` (configuration management), `go-cache` or Redis (merchant dictionary cache), `sqlc` (type-safe SQL)
- LLM Client: Provider-specific SDKs (optional: `anthropic-sdk-go` for Claude fallback, injected at runtime based on config)
- Frontend: existing API client, category UI components

**Storage**: PostgreSQL 14+ (extends existing schema); new tables: Category, MerchantDictionary, TransactionCategory, CategoryStats

**Testing**: Go `testing` package with `testify` (backend); `vitest` + React Testing Library (frontend); contract tests for categorization endpoints

**Target Platform**: Web service (backend REST API); browser client (React)

**Project Type**: Web service with backend + frontend (extends Statement Import architecture from Phase 1)

**Performance Goals**: 
- Rule-based categorization: <100ms per transaction (dictionary lookup + cache)
- LLM categorization (Ollama): 200-400ms per transaction (local inference on GPU/CPU)
- LLM categorization (Claude fallback): <2s per batch (with 10 concurrent API calls)
- Preview with categories: <10s total (import + rule-based + async Ollama)
- Recategorization update: <2s (UI + database)

**Constraints**: 
- Initial merchant dictionary: ≥500 entries for Indian banks
- LLM API cost: Zero for Ollama (local), minimize external API calls via rule-based-first strategy
- Confidence scoring: Rule=100%, Fuzzy=85-99%, Ollama LLM=0.65-0.80, Claude LLM=0.85-0.98
- Graceful degradation: If LLM unavailable, default to "Uncategorized" without blocking import
- Provider abstraction: Easy switching between Ollama, Claude, OpenAI via config (no code changes)

**Scale/Scope**: MVP with 10 predefined categories; extends existing transaction model (from Phase 1)

## Constitution Check

*GATE: Must pass before Phase 0 research. Re-check after Phase 1 design.*

### Alignment with MoneyPlan AI Constitution

| Principle | Status | Rationale |
|-----------|--------|-----------|
| **I. Data Privacy & Security** | ✅ PASS | No additional PII generated; categorization derived from transactions already imported; LLM API key managed via environment variables; user data never sent to external LLM beyond merchant name + amount |
| **II. Modular Service Architecture** | ✅ PASS | Categorization Service extends Transaction domain independently; clear REST API contract; no direct DB access from frontend; service owns categorization logic |
| **III. Data Quality First** | ✅ PASS | Merchant dictionary validated (≥500 entries); confidence scores track quality; user corrections feed back into dictionary; LLM includes confidence threshold (75% min) |
| **IV. API Contract Testing** | ✅ PASS | Contract tests for categorize endpoint; LLM API mocked in unit tests; merchant dictionary queries tested |
| **V. AI Grounding & Explainability** | ✅ PASS | LLM suggestions include explanation (merchant name + inferred category); confidence scores shown to user; rule-based matches explained (known merchant); user corrections visible with audit trail |

**Gate Status**: ✅ PASS — No constitution violations. Feature aligns with all applicable principles. AI Grounding principle is directly applicable.

## Project Structure

### Documentation (this feature)

```text
specs/002-transaction-categorization/
├── plan.md                      # This file
├── spec.md                      # Feature specification (user stories, requirements)
├── categories-reference.md      # 10 predefined categories (names, colors, icons, examples)
├── research.md                  # Phase 0 output (merchant dictionary sources, LLM accuracy, fuzzy matching)
├── data-model.md                # Phase 1 output (Category, MerchantDictionary, TransactionCategory, CategoryStats)
├── contracts/                   # Phase 1 output (categorize, recategorize, analytics API schemas)
│   ├── categorize-endpoint.md
│   ├── recategorize-endpoint.md
│   └── category-analytics-endpoint.md
├── quickstart.md                # Phase 1 output (validation: upload → categorize → preview → confirm → view)
└── tasks.md                     # Phase 2 output (/speckit-tasks command; Phases 1-6 implementation tasks)
```

### Source Code (repository root)

**Selected Structure**: Web application with backend and frontend with pluggable LLM provider architecture

```text
backend/
├── internal/
│   ├── categorization/           # NEW: Transaction Categorization Service
│   │   ├── models.go             # Category, MerchantDictionary, TransactionCategory, CategoryStats
│   │   ├── service.go            # Categorization business logic (provider-agnostic)
│   │   ├── merchant_dict.go      # Merchant dictionary lookup (trie for <10ms latency)
│   │   ├── provider.go           # LLMProvider interface (Categorize method)
│   │   ├── providers/            # NEW: Provider implementations
│   │   │   ├── ollama.go         # Ollama provider (local Mistral 7B)
│   │   │   ├── claude.go         # Claude provider (optional, fallback)
│   │   │   └── openai.go         # OpenAI provider (future)
│   │   └── confidence.go         # Confidence scoring logic
│   ├── config/                   # NEW: Configuration management
│   │   ├── llm_config.go         # LLM provider config struct + Viper loading
│   │   └── config.yaml           # LLM provider definitions
│   ├── api/
│   │   ├── categorize.go         # UPDATED: POST /api/transactions/categorize
│   │   ├── recategorize.go       # NEW: POST /api/transactions/{id}/recategorize
│   │   └── category_stats.go     # NEW: GET /api/categories/{id}/stats
│   └── db/
│       ├── migrations/           # SQL migrations for new tables
│       └── queries/              # sqlc-generated queries
└── tests/
    ├── contract/                 # API contract tests for categorization
    │   └── categorize_test.go
    ├── integration/              # Service integration tests
    │   ├── merchant_dict_test.go
    │   └── llm_categorization_test.go
    └── unit/
        ├── confidence_test.go
        └── merchant_matching_test.go

frontend/
├── src/
│   ├── pages/
│   │   ├── statements/
│   │   │   └── PreviewModal.tsx        # UPDATED: Show categories during preview
│   │   └── categories/
│   │       ├── CategoryDashboard.tsx   # NEW: Category spending view
│   │       ├── CategoryDetail.tsx      # NEW: Drill-down transactions
│   │       └── RecategorizeModal.tsx   # NEW: Recategorize UI
│   ├── components/
│   │   ├── CategoryBadge.tsx           # NEW: Category with color
│   │   └── TransactionPreview.tsx      # UPDATED: Include category column
│   └── services/
│       └── categorizationApi.ts        # NEW: API client
└── tests/
    └── features/
        ├── categorize.test.tsx         # NEW: Categorization preview
        └── category-analytics.test.tsx # NEW: Category dashboard
```

**Structure Decision**: Categorization Service added to backend `internal/categorization/`. Frontend extends existing pages with new category features. Service uses existing Transaction infrastructure but owns categorization logic independently.

## Complexity Tracking

> **No constitution violations; no complexity exceptions needed.**

---

## Next Steps

**Phase 0 (Research)** — ✅ DONE: Research materials in `research.md`

**Phase 1 (Design)** — ✅ DONE: 
- Data models defined in `data-model.md`
- API contracts in `contracts/`
- Merchant dictionary schema drafted in `categories-reference.md`
- Quickstart scenarios in `quickstart.md`

**Phase 2-6 (Implementation)** — IN PROGRESS: 
- MVP (Phases 1-3): Rule-based categorization with merchant dictionary
- Phases 4+: LLM integration, analytics, admin interface (deferred)
- See `tasks.md` for complete task breakdown and dependencies

**MVP Focus**: User uploads statement → automatically categorized by ≥500-merchant dictionary → categories shown in preview → persisted to DB → ready for downstream analytics/budgets.
