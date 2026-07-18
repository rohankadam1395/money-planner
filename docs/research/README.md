# Research & Reference Materials

This folder contains research findings, design patterns, and reference materials for Money Planner AI features.

**Organization**:
- `llm/` — LLM integration research (Ollama, Claude, provider abstraction, prompt engineering)
- `architecture/` — System design patterns and decisions
- `performance/` — Benchmarks, optimization strategies, latency analysis

## LLM Research (Phase 002: Transaction Categorization)

See `llm/` subfolder for detailed findings on:
- **Ollama API Integration**: Mistral 7B specifications, performance, latency
- **Go Provider Abstraction**: Interface patterns, factory design, dependency injection
- **Configuration Strategy**: YAML config, Viper setup, environment variable handling
- **Prompt Engineering**: Unified templates, model-specific handling, confidence scoring
- **Embedding Models**: When semantic similarity is useful (Phase 5+)

### Key Decisions

1. **Default LLM**: Ollama Mistral 7B (free, local, 200-400ms latency, 80-85% accuracy)
2. **Provider Abstraction**: Interface-based, pluggable (switch via env var, no code changes)
3. **Not in MVP**: Embeddings, semantic search, anomaly detection (Phase 5+)

## How to Use This Research

**When implementing Phase 002** (Transaction Categorization):
- Read `llm/ollama-api-integration.md` for Ollama setup and API details
- Read `llm/go-provider-patterns.md` for implementation patterns
- Read `llm/prompt-engineering.md` for categorization prompt design

**When planning Phase 005+** (Anomaly Detection, Semantic Search):
- Refer to embedding models findings in `llm/embedding-models.md`
- Review LLM latency baselines in `llm/ollama-api-integration.md` and `llm/prompt-engineering.md`

**When adding new LLM-powered features**:
- Use provider abstraction pattern from Phase 002
- Follow configuration strategy in `llm/config-strategy.md`
- Copy functional options pattern from `llm/go-provider-patterns.md`

---

**Last Updated**: 2026-07-18 (Phase 002 planning)
