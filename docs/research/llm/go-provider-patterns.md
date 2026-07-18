# Go LLM Provider Abstraction Patterns

**Source**: Agent research on Go design patterns for pluggable providers  
**Date**: 2026-07-18  
**Status**: Validated, used in Phase 002 implementation

## Pattern Overview

Go's implicit interface satisfaction and composition enable clean, pluggable provider abstraction. This pattern mirrors `database/sql` — proven at massive scale.

## Core Pattern: Segregated Interfaces

```go
// Minimal interface: all providers must implement
type Provider interface {
    Generate(ctx context.Context, req *GenerateRequest) (*GenerateResponse, error)
}

// Optional capabilities (not all providers implement)
type StreamProvider interface {
    GenerateStream(ctx context.Context, req *GenerateRequest) (StreamResponse, error)
}

type HealthChecker interface {
    Healthy(ctx context.Context) error
}

// Shared request/response types (reusable across providers)
type GenerateRequest struct {
    Prompt      string
    Model       string
    Temperature float32
    MaxTokens   int
    Extra       map[string]interface{} // Provider-specific overrides
}

type GenerateResponse struct {
    Text     string
    Tokens   int
    Provider string // Track which provider generated this
}
```

**Why segregated?**
- Forces focused contracts
- Implementers don't stub unused methods
- Extends without breaking existing code (proven in `io.Reader`, `io.Writer`)

## Factory Pattern with Functional Options

```go
// Configuration (provider-agnostic)
type ProviderConfig struct {
    APIKey     string
    BaseURL    string
    Model      string
    Timeout    time.Duration
    MaxRetries int
}

// Option function for composable config
type Option func(*ProviderConfig) error

// Factory with registry (simple, not over-engineered)
func NewProvider(name string, opts ...Option) (Provider, error) {
    cfg := &ProviderConfig{
        Timeout:    30 * time.Second,
        MaxRetries: 3,
    }
    
    for _, opt := range opts {
        if err := opt(cfg); err != nil {
            return nil, fmt.Errorf("config error: %w", err)
        }
    }
    
    switch name {
    case "ollama":
        return newOllamaProvider(cfg)
    case "claude":
        return newClaudeProvider(cfg)
    case "openai":
        return newOpenAIProvider(cfg)
    default:
        return nil, fmt.Errorf("unknown provider: %s", name)
    }
}

// Functional option functions
func WithAPIKey(key string) Option {
    return func(cfg *ProviderConfig) error {
        if key == "" {
            return errors.New("APIKey cannot be empty")
        }
        cfg.APIKey = key
        return nil
    }
}

func WithBaseURL(url string) Option {
    return func(cfg *ProviderConfig) error {
        cfg.BaseURL = url
        return nil
    }
}

func WithTimeout(d time.Duration) Option {
    return func(cfg *ProviderConfig) error {
        cfg.Timeout = d
        return nil
    }
}
```

**Why functional options?**
- Composable, extensible configuration
- Validates each setting independently
- No long unreadable parameter lists
- Proven pattern (Google Cloud SDK, AWS SDK)

## Environment-Driven Factory

```go
func FromEnv() (Provider, error) {
    providerName := os.Getenv("LLM_PROVIDER")
    if providerName == "" {
        providerName = "ollama" // Default
    }
    
    var opts []Option
    
    // Load provider-specific config from env
    switch providerName {
    case "openai":
        if key := os.Getenv("OPENAI_API_KEY"); key != "" {
            opts = append(opts, WithAPIKey(key))
        }
    case "claude":
        if key := os.Getenv("ANTHROPIC_API_KEY"); key != "" {
            opts = append(opts, WithAPIKey(key))
        }
    case "ollama":
        if url := os.Getenv("OLLAMA_BASE_URL"); url != "" {
            opts = append(opts, WithBaseURL(url))
        }
    }
    
    return NewProvider(providerName, opts...)
}
```

**Zero code changes to switch providers** — just set `LLM_PROVIDER=claude` and restart.

## Dependency Injection

```go
// Services depend on interface, NOT concrete type
type CategorizationService struct {
    provider llm.Provider  // Interface!
    db       *sql.DB
    logger   Logger
}

func NewCategorizationService(
    provider llm.Provider,
    db *sql.DB,
    logger Logger,
) *CategorizationService {
    return &CategorizationService{
        provider: provider,
        db:       db,
        logger:   logger,
    }
}

// Categorization logic is provider-agnostic
func (cs *CategorizationService) Categorize(
    ctx context.Context,
    merchant string,
    amount float64,
) (string, error) {
    // Service doesn't know if provider is Ollama, Claude, or OpenAI
    resp, err := cs.provider.Generate(ctx, &llm.GenerateRequest{
        Prompt: fmt.Sprintf("Categorize: %s (₹%.2f)", merchant, amount),
        Model:  "auto", // Provider uses its default model
    })
    if err != nil {
        cs.logger.Error("categorization failed", "error", err)
        return "Uncategorized", nil // Graceful degradation
    }
    return resp.Text, nil
}
```

## Wire-Up in main.go

```go
func main() {
    logger := logrus.New()
    
    // 1. Create provider once (injected everywhere)
    provider, err := llm.FromEnv()
    if err != nil {
        logger.Fatalf("failed to initialize LLM provider: %v", err)
    }
    
    db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
    if err != nil {
        logger.Fatalf("database error: %v", err)
    }
    
    // 2. Inject into services (zero dependencies on concrete types)
    categSvc := categorization.NewService(provider, db, logger)
    advisorSvc := advisor.NewService(provider, db, logger)
    
    // 3. Services transparently use any provider
    router := chi.NewRouter()
    router.POST("/api/v1/categorize", handlers.Categorize(categSvc))
    router.POST("/api/v1/advisor", handlers.Advisor(advisorSvc))
    
    logger.Infof("LLM provider: %s", os.Getenv("LLM_PROVIDER"))
    router.ListenAndServe(":8080")
}
```

## Testing with Mock Provider

```go
// Minimal mock (just implements interface)
type MockProvider struct {
    Response *GenerateResponse
    Err      error
}

func (mp *MockProvider) Generate(
    ctx context.Context,
    req *GenerateRequest,
) (*GenerateResponse, error) {
    return mp.Response, mp.Err
}

// Service tests: no network, fast, deterministic
func TestCategorize_Success(t *testing.T) {
    mock := &MockProvider{
        Response: &GenerateResponse{
            Text:     "Food",
            Provider: "mock",
        },
    }
    
    svc := categorization.NewService(mock, nil, logger)
    category, err := svc.Categorize(context.Background(), "Swiggy", 500)
    
    assert.NoError(t, err)
    assert.Equal(t, "Food", category)
}

func TestCategorize_ProviderFailure(t *testing.T) {
    mock := &MockProvider{
        Err: errors.New("provider unavailable"),
    }
    
    svc := categorization.NewService(mock, nil, logger)
    category, err := svc.Categorize(context.Background(), "Unknown", 100)
    
    // Graceful degradation: no error, defaults to "Uncategorized"
    assert.NoError(t, err)
    assert.Equal(t, "Uncategorized", category)
}
```

## Middleware for Cross-Cutting Concerns

```go
// Retry middleware
func WithRetry(p Provider, maxRetries int) Provider {
    return &retryMiddleware{provider: p, maxRetries: maxRetries}
}

// Logging middleware
func WithLogging(p Provider, logger Logger) Provider {
    return &loggingMiddleware{provider: p, logger: logger}
}

// Caching middleware (if needed)
func WithCache(p Provider, ttl time.Duration) Provider {
    return &cacheMiddleware{provider: p, ttl: ttl}
}

// Composition in main.go
provider := llm.NewProvider("ollama", ...)
provider = WithLogging(provider, logger)
provider = WithRetry(provider, 3)
provider = WithCache(provider, 24*time.Hour)

// Services use provider transparently; concerns are automatic
```

## Comparison to Alternatives

| Approach | Complexity | Testability | Flexibility | Recommended |
|----------|-----------|-------------|-------------|-------------|
| **Interface + Factory** | Low | Excellent | High | ✅ YES |
| Enum + Switch | Low | Poor | Low | ❌ Tight coupling |
| Plugin system | High | Medium | Very High | ❌ Overkill |
| Monolithic class | High | Poor | Low | ❌ Inflexible |

## Implementation Checklist

- [ ] Define `Provider` interface with `Generate` method
- [ ] Create `ProviderConfig` and `Option` types
- [ ] Implement `NewProvider` factory with registry
- [ ] Implement `FromEnv()` for zero-config switching
- [ ] Create `OllamaProvider` implementing `Provider`
- [ ] Create `ClaudeProvider` stub (fill in Phase 4)
- [ ] Create `MockProvider` for testing
- [ ] Wire up in `main.go` with dependency injection
- [ ] Add middleware for retry, logging, caching

## Timeline

- **2-3 hours**: Interface + factory + Ollama provider
- **1-2 hours**: Service integration + dependency injection
- **1-2 hours**: Testing with mock provider
- **2-3 hours**: Add Claude + OpenAI providers (later phases)

**Total for MVP**: 4-7 hours

---

**References**:
- [Go Code Review Comments - Interfaces](https://github.com/golang/go/wiki/CodeReviewComments#interfaces)
- [Functional Options Pattern](https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis)
- [Dependency Injection in Go](https://pkg.go.dev/google.golang.org/grpc)
