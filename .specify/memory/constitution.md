# MoneyPlan AI Constitution

## Core Principles

### I. Data Privacy & Security
Financial data is sensitive and personally identifying. MUST: Encrypt PII at rest and in transit; Minimize data retention; Implement role-based access control; Document all data handling paths; Validate inputs against injection attacks. Rationale: A breach exposes the user's complete financial history and personal behaviors.

### II. Modular Service Architecture
The backend comprises distinct services (Transaction, Budget, AI Insight, Forecast). MUST: Each service owns its domain; Define clear API contracts between services; No direct database access across service boundaries; Version APIs explicitly. Rationale: Services deployed independently; teams can work in parallel; isolates failure domains.

### III. Data Quality First
Accurate financial advice depends entirely on clean transaction data. MUST: Merchant normalization rules documented; Duplicate detection tested against real bank data; Category mappings validated; Handle edge cases (unusual formats, multi-currency). Rationale: Downstream features (budgets, insights, forecasts) inherit garbage-in-garbage-out risk.

### IV. API Contract Testing
Services communicate via REST/gRPC APIs; frontend consumes backend APIs. MUST: Integration tests validate API contracts; Mock external APIs (OpenAI, bank statement parsers) in unit tests; Schema changes trigger contract verification; Backwards compatibility justified explicitly. Rationale: Prevents silent contract drift; catches breaking changes before deployment.

### V. AI Grounding & Explainability
LLM recommendations must reference the user's actual data. MUST: Every insight includes supporting data (transaction examples, date ranges, amounts); LLM outputs are summarized, not passed directly to users; Deterministic rules complement AI (e.g., fraud detection). Rationale: User trust depends on seeing why the system made a recommendation; prevents hallucinated financial advice.

## Technology Standards

**Backend**: Go 1.25+; services expose REST APIs; structured logging with correlation IDs; PostgreSQL for ACID transactions; OpenSearch for analytics queries.

**Frontend**: React/Next.js; types validated with TypeScript; API client auto-generated from schema.

**Data Pipelines**: CSV/PDF parsing tested against real bank formats; duplicate detection over 12-month windows; rule-based categorization with LLM fallback.

**Testing Discipline**: Unit tests for business logic (categorization, forecasting); integration tests for service-to-service APIs; end-to-end tests for critical flows (upload → categorize → display).

## Development Workflow

**Feature Development**:
1. Spec: What transactions/categories/amounts must the feature handle?
2. Contract: What API shape? What data model?
3. Tests: Write failing tests (TDD); mock external dependencies.
4. Implement: Code to pass tests.
5. Review: Contract compliance, data privacy checklist, test coverage ≥80%.

**Deployment**:
- Each service versioned independently (semver).
- Database migrations must be backwards-compatible or include a rollback plan.
- Sensitive config (API keys, encryption keys) via environment, never in code.

## Governance

This constitution supersedes all other practices. Amendments require:
- Written justification (why existing principle is insufficient).
- Migration plan for affected code.
- Documentation update to dependent templates.

**Compliance checks**: All PRs must verify this constitution in code review. Exceptions are rare and require explicit sign-off and documentation.

**Version**: 1.0.0 | **Ratified**: 2026-07-05 | **Last Amended**: 2026-07-05

---

## Sync Impact Report

**Version Change**: Initial → 1.0.0 (PATCH level for future amendments)

**Principles Defined**:
- Data Privacy & Security
- Modular Service Architecture
- Data Quality First
- API Contract Testing
- AI Grounding & Explainability

**Sections Added**:
- Technology Standards
- Development Workflow
- Governance

**Template Files Requiring Review**:
- ✅ `.specify/templates/plan-template.md` — Verify alignment with Modular Service Architecture and Development Workflow
- ✅ `.specify/templates/spec-template.md` — Ensure specs capture Data Quality requirements
- ✅ `.specify/templates/tasks-template.md` — Task categorization reflects Testing Discipline and API Contract Testing
- ⚠ Runtime guidance (e.g., backend service templates) — May benefit from Privacy & Security checklist

