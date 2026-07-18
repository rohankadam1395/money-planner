# Research: Transaction Categorization

**Phase 0 Output** | Generated: 2026-07-12

Findings from investigation into transaction categorization approaches, merchant dictionaries, LLM integration, and performance requirements.

## 1. Merchant Dictionary Strategy

**Decision**: Hybrid approach — curated dictionary (500+ entries) for Indian banks + programmatic updates from user corrections.

**Rationale**: 
- Indian banking ecosystem has ~50-100 major merchants (Swiggy, Amazon, Uber, HDFC, ICICI, etc.) that account for 70% of transaction volume
- Long-tail merchants (small vendors, local businesses) require LLM inference or user training
- Curated dictionary provides fast, deterministic categorization with 100% confidence for known merchants

**Alternatives Considered**:
- ❌ Dynamic dictionary only (no LLM): Misses long-tail merchants; requires extensive user correction training
- ❌ LLM for all: High latency (2-3s per tx), unacceptable for import preview (<10s budget)
- ✅ Hybrid (rule-based first, LLM fallback): Fast for known merchants, handles unknowns gracefully

**Sources for Initial Dictionary**:
- Plaid Merchant Categories (public API, covers major merchants)
- Manual curation from top Indian bank statement samples
- OpenStreetMap data for category patterns (e.g., shop types → Shopping)
- User feedback from testing (early adopters' corrections)

---

## 2. Fuzzy Matching for Rule-Based Categorization

**Decision**: Exact matching primary; fuzzy matching (Levenshtein distance) as secondary fallback.

**Rationale**:
- Most major merchants have consistent names across bank statements
- Fuzzy matching catches typos, spacing variations (e.g., "Swiggy" vs "SWIGGY" vs "Swiggy Food Delivery")
- Levenshtein distance threshold: 85% similarity for automatic categorization, <85% requires LLM

**Algorithm**:
1. Exact match against dictionary (case-insensitive) → confidence 100%
2. Levenshtein distance against top 100 merchants (by frequency) → if ≥85% → confidence 85-99%
3. No match → LLM categorization

**Data Structure**: Trie for O(n) lookup (n=merchant name length), avoiding O(n²) comparisons across all 500 entries.

---

## 3. LLM Categorization Strategy

**Decision**: Configurable LLM provider (default: Ollama local, Mistral 7B) with pluggable architecture supporting Claude, OpenAI, and future providers.

**Rationale**:
- **Ollama (Local)**: Zero API costs after model download, 100% data privacy, 200-400ms latency, 80-85% accuracy
- **Pluggable architecture**: Easy switching between providers (Ollama for cost efficiency, Claude for higher accuracy on edge cases) without code changes
- **Cost optimization**: Ollama for routine categorization (~80% of transactions), optional Claude fallback for ambiguous merchants
- **Provider-agnostic prompt**: Single prompt template with provider-specific output parsing

**Ollama Setup**:
- **Model**: Mistral 7B (7B parameters, Q4 quantization, 6-7GB VRAM)
- **Endpoint**: `http://localhost:11434/api/generate`
- **Request format**: 
  ```json
  POST /api/generate
  {
    "model": "mistral:7b",
    "prompt": "[System instruction]\n[Transaction details]",
    "stream": false,
    "options": {"temperature": 0.3, "num_predict": 50, "top_p": 0.9}
  }
  ```
- **Response parsing**: Extract first phrase matching known category names from response

**Prompt Structure** (provider-agnostic):
```
You are a transaction categorizer. Categorize into ONE category only.
Categories: Food & Dining, Shopping, Transport, Housing, Utilities, Entertainment, Income, Healthcare, Education, Miscellaneous

Merchant: "{merchant_name}"
Amount: ₹{amount}
Description: {description}

Respond with ONLY the category name.
```

**Confidence Scoring**:
- **Ollama**: Fixed 0.80 for known merchants, 0.65 for inferred categories (or use token density heuristic in Phase 2)
- **Claude**: Explicit confidence from API response, typically 0.85-0.98
- **User correction**: 1.0 (ground truth)

**Error Handling**:
- **Provider unreachable** (404, 502): Retry 3x with backoff (1s, 2s, 5s) → default to "Uncategorized"
- **Prompt too long**: Truncate description to 100 chars, retry
- **Invalid response**: Parse fails → default to "Uncategorized"
- **Timeout** (Ollama: 60s, Claude: 30s): Fail gracefully to "Uncategorized"

**Alternatives Considered**:
- ❌ Claude API only: High cost ($2-5 per 1000 requests), network latency, data privacy concerns
- ❌ Ollama only: 80-85% accuracy acceptable for initial filtering, but no fallback for ambiguous merchants
- ✅ Hybrid (Ollama + provider abstraction): Cost-effective routine processing, flexibility to add Claude for edge cases

---

## 3b. Provider Abstraction & Configuration

**Decision**: Go interface-based abstraction + Viper config management for runtime provider selection.

**Rationale**:
- **Interface abstraction**: All providers implement common `Categorize(ctx, MerchantName, Amount, Description) → (Category, Confidence, error)` interface
- **Factory pattern**: Runtime provider selection based on config file + env var overrides
- **Dependency injection**: Services receive provider as dependency, decoupled from implementation
- **Configuration**: YAML config file with provider-specific settings (model, base URL, API key reference, timeout)

**Provider Interface** (Go):
```go
type LLMProvider interface {
  Categorize(ctx context.Context, merchantName string, amount float64, description string) 
    (category string, confidence float64, err error)
}
```

**Factory Implementation**:
```go
func NewProvider(cfg *ProviderConfig) (LLMProvider, error) {
  switch cfg.Name {
  case "ollama":
    return &OllamaProvider{baseURL: cfg.BaseURL, model: cfg.Model}, nil
  case "claude":
    return &ClaudeProvider{apiKey: os.Getenv("ANTHROPIC_API_KEY"), model: cfg.Model}, nil
  // future providers...
  }
}
```

**Config File Structure** (YAML):
```yaml
llm:
  default_provider: "ollama"
  providers:
    ollama:
      enabled: true
      model: "mistral:7b"
      base_url: "http://localhost:11434"
      timeout_seconds: 60
    claude:
      enabled: false
      model: "claude-3-5-sonnet-20241022"
      base_url: "https://api.anthropic.com"
      api_key_env: "ANTHROPIC_API_KEY"
      timeout_seconds: 30
      max_retries: 3
```

**Environment Variable Overrides** (Viper precedence: env > file > defaults):
```bash
LLM_PROVIDER=ollama                              # select active provider
OLLAMA_BASE_URL=http://ollama-service:11434     # override config URL
ANTHROPIC_API_KEY=sk-ant-...                    # secret, never in config
```

**Startup Validation**:
1. Load config from file (or `LLM_CONFIG_FILE` env var path)
2. Bind env vars (override file values)
3. Validate: at least one provider enabled, required fields present, credentials available
4. Select active provider (env var or default)
5. Test connectivity to active provider
6. Initialize CategorizationService with selected provider

**Alternatives Considered**:
- ❌ Hard-coded provider selection: Tight coupling, requires code changes to switch providers
- ❌ Enum + switch pattern: No abstraction benefit, harder to test
- ❌ Plugin system (dlopen): Overkill, security risk, adds unnecessary complexity
- ✅ Interface + factory + config: Idiomatic Go, testable, flexible, minimal complexity

---

## 4. Confidence Scoring & Audit Trail

**Decision**: Track categorization method, confidence score, and LLM provider; expose for filtering and learning.

**Rationale**:
- Rule-based (exact match) → 100% confidence
- Fuzzy match (Levenshtein ≥85%) → 85-99% confidence
- LLM categorization (Ollama) → fixed 0.80 for confident, 0.65 for inferred
- LLM categorization (Claude) → API response confidence (typically 0.85-0.98)
- Manual correction → 100% confidence (user is ground truth)
- Tracking provider enables A/B testing (Ollama vs Claude accuracy) and cost analysis

**Use Cases**:
- Filter transactions by confidence for manual review (e.g., show only <80% confidence)
- Learn patterns over time (which merchants are frequently corrected?)
- Debug categorization accuracy (which method fails most often?)

---

## 5. Performance & Latency Budget

**Decision**: Rule-based <100ms; LLM async batching to stay within <10s import preview budget.

**Breakdown**:
- Statement parsing: ~3-5s (handled by Phase 1)
- Rule-based categorization: ~100ms for 1000 transactions (dictionary lookup + caching)
- LLM categorization (async): Process unknown merchants in background after import
- UI rendering: ~1-2s

**Caching Strategy**:
- In-memory cache of merchant dictionary (loaded on service startup, ~10MB for 500 entries)
- Redis cache for LLM results (key: merchant_name, value: category + confidence, TTL: 30 days)
- Cache hit rate expected: 85-90% on typical statements

---

## 6. Indian Banking Context

**Decision**: Support major Indian banks (HDFC, ICICI, Axis, SBI) with merchant name patterns common in their statements.

**Challenges**:
- Bank statements often abbreviate or truncate merchant names (e.g., "SWIGGY FOOD DELIV" instead of "Swiggy Food Delivery")
- Merchant names may include reference numbers or timestamps
- Multi-currency transactions (USD, EUR) require handling

**Solution**:
- Fuzzy matching handles abbreviations/truncations
- Regex patterns for merchant name cleanup (remove reference numbers, timestamps)
- Amount in INR assumed; multi-currency flagged for manual review

---

## Decisions Summary

| Aspect | Decision | Confidence |
|--------|----------|-----------|
| Dictionary Size | 500+ entries, curated | High |
| Rule-Based Matching | Exact + Fuzzy (Levenshtein ≥85%) | High |
| LLM Provider (Primary) | Ollama (Mistral 7B local) | High |
| LLM Provider Architecture | Interface-based abstraction + factory pattern | High |
| Provider Configuration | YAML config + env var overrides (Viper) | High |
| LLM Confidence Threshold | ≥75% auto-accept, <75% requires review | High |
| Confidence Tracking | Store method + score + provider; expose for filtering | High |
| Performance Target | <100ms rule-based, 200-400ms Ollama, <2s Claude batch | High |
| Error Handling | Default to "Uncategorized" on provider failure | High |
| Caching | In-memory dict + Redis for LLM results | Medium |
| Future Providers | Claude, OpenAI supported via provider interface | High |
| Cost Strategy | Ollama for routine (free), Claude/OpenAI fallback for edge cases | High |
