# LLM Configuration Strategy

**Source**: Agent research on configuration patterns for LLM provider selection  
**Date**: 2026-07-18  
**Status**: Validated, implemented in Phase 002

## Overview

YAML config file + environment variable overrides enables runtime provider switching without code changes. Inspired by Kubernetes, OpenStack, and cloud SDKs.

## Config File Structure

**File**: `backend/config/llm-config.yaml`

```yaml
llm:
  default_provider: "ollama"  # Fallback if LLM_PROVIDER not set
  
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
      api_key_env: "ANTHROPIC_API_KEY"  # Reference to secret source
      timeout_seconds: 30
      max_retries: 3
      
    openai:
      enabled: false
      model: "gpt-4-mini"
      base_url: "https://api.openai.com/v1"
      api_key_env: "OPENAI_API_KEY"
      timeout_seconds: 30
      max_retries: 3
```

**Why YAML + env references?**
- Symmetric provider config (all providers look the same)
- API keys never in config files (referenced via env var)
- Easy to enable/disable providers
- Clear which provider is active

## Config Struct in Go

```go
// backend/internal/config/llm_config.go

type LLMConfig struct {
    DefaultProvider string                    `mapstructure:"default_provider"`
    Providers       map[string]ProviderConfig `mapstructure:"providers"`
}

type ProviderConfig struct {
    Enabled      bool   `mapstructure:"enabled"`
    Model        string `mapstructure:"model"`
    BaseURL      string `mapstructure:"base_url"`
    APIKeyEnv    string `mapstructure:"api_key_env"`      // Env var name for secret
    TimeoutSecs  int    `mapstructure:"timeout_seconds"`
    MaxRetries   int    `mapstructure:"max_retries"`
}
```

## Loading with Viper

```go
// backend/internal/config/loader.go

func LoadLLMConfig(configPath string) (*LLMConfig, error) {
    v := viper.New()
    v.SetConfigType("yaml")
    
    // Explicit path or search in defaults
    if configPath != "" {
        v.SetConfigFile(configPath)
    } else {
        v.AddConfigPath("/etc/app")
        v.AddConfigPath("./config")
        v.AddConfigPath(".")
    }
    
    // Bind env vars (env > file > defaults)
    v.BindEnv("llm.default_provider", "LLM_PROVIDER")
    v.BindEnv("llm.providers.ollama.base_url", "OLLAMA_BASE_URL")
    v.BindEnv("llm.providers.claude.timeout_seconds", "CLAUDE_TIMEOUT_SECONDS")
    
    if err := v.ReadInConfig(); err != nil {
        return nil, fmt.Errorf("failed to read LLM config: %w", err)
    }
    
    var cfg LLMConfig
    if err := v.UnmarshalKey("llm", &cfg); err != nil {
        return nil, fmt.Errorf("failed to parse LLM config: %w", err)
    }
    
    return &cfg, nil
}
```

## Environment Variables

**Core Selection**:
```bash
LLM_CONFIG_FILE=/etc/app/llm-config.yaml  # Optional, searches defaults if unset
LLM_PROVIDER=ollama                        # Overrides default_provider in config
```

**Provider-Specific Secrets** (NEVER in config files):
```bash
ANTHROPIC_API_KEY=sk-ant-...              # For Claude provider
OPENAI_API_KEY=sk-...                     # For OpenAI provider
```

**Provider-Specific Overrides** (runtime tuning):
```bash
OLLAMA_BASE_URL=http://ollama-service:11434
CLAUDE_TIMEOUT_SECONDS=45
```

**Logging**:
```bash
LLM_LOG_LEVEL=debug
```

## Precedence (Viper order)

1. **Environment variables** (highest priority)
2. **Config file** values
3. **Defaults** (hardcoded in code)

Example:
- `LLM_PROVIDER=claude` env var overrides `default_provider: ollama` in config file
- `OLLAMA_BASE_URL=http://...` overrides config file's `base_url`

## Startup Validation

```go
// backend/cmd/statement-import-api/main.go

func main() {
    logger := logrus.New()
    
    // 1. Load config
    cfg, err := config.LoadLLMConfig("")
    if err != nil {
        logger.Fatalf("failed to load LLM config: %v", err)
    }
    
    // 2. Validate config
    if err := validateLLMConfig(cfg); err != nil {
        logger.Fatalf("invalid LLM config: %v", err)
    }
    
    // 3. Select active provider
    activeProvider := os.Getenv("LLM_PROVIDER")
    if activeProvider == "" {
        activeProvider = cfg.DefaultProvider
    }
    
    providerCfg, ok := cfg.Providers[activeProvider]
    if !ok {
        logger.Fatalf("provider %q not found in config", activeProvider)
    }
    
    if !providerCfg.Enabled {
        logger.Fatalf("provider %q is disabled in config", activeProvider)
    }
    
    // 4. Test connectivity (optional but recommended)
    if err := testProviderConnectivity(activeProvider, providerCfg); err != nil {
        logger.Warnf("provider %s unreachable: %v (will fail at runtime)", activeProvider, err)
    }
    
    // 5. Initialize provider
    provider, err := llm.NewProvider(activeProvider, opts...)
    if err != nil {
        logger.Fatalf("failed to initialize provider: %v", err)
    }
    
    logger.Infof("LLM provider initialized: %s", activeProvider)
    
    // ... rest of startup
}
```

## Validation Function

```go
func validateLLMConfig(cfg *LLMConfig) error {
    if cfg.DefaultProvider == "" {
        return fmt.Errorf("default_provider must be set")
    }
    
    // Default provider must exist
    if _, ok := cfg.Providers[cfg.DefaultProvider]; !ok {
        return fmt.Errorf("default_provider %q not found in providers", cfg.DefaultProvider)
    }
    
    // At least one provider must be enabled
    hasEnabled := false
    for name, p := range cfg.Providers {
        if !p.Enabled {
            continue
        }
        hasEnabled = true
        
        // Enabled provider must have required fields
        if p.Model == "" {
            return fmt.Errorf("provider %q has empty model", name)
        }
        
        // If provider requires API key, verify env var is set
        if p.APIKeyEnv != "" && os.Getenv(p.APIKeyEnv) == "" {
            return fmt.Errorf("provider %q requires %s env var to be set", name, p.APIKeyEnv)
        }
    }
    
    if !hasEnabled {
        return fmt.Errorf("at least one provider must be enabled in config")
    }
    
    return nil
}
```

## Development Setup

**.env.local** (gitignored):
```bash
LLM_PROVIDER=ollama
OLLAMA_BASE_URL=http://localhost:11434
```

## Production Setup

**With AWS Secrets Manager**:
```bash
# Load secrets before startup
ANTHROPIC_API_KEY=$(aws secretsmanager get-secret-value --secret-id anthropic-api-key | jq -r '.SecretString')
export ANTHROPIC_API_KEY

# Config file specifies providers, env var selects active one
export LLM_PROVIDER=claude

# Start app
./app
```

**With HashiCorp Vault**:
```bash
# Fetch secrets from Vault
ANTHROPIC_API_KEY=$(vault kv get -field=key secret/anthropic)
export ANTHROPIC_API_KEY

export LLM_PROVIDER=claude
./app
```

## Switching Providers at Runtime

**From Ollama to Claude**:
```bash
# 1. Set env var
export LLM_PROVIDER=claude
export ANTHROPIC_API_KEY=sk-ant-...

# 2. Restart backend
# (Config file already has Claude provider defined)

# 3. Verify
curl http://localhost:8080/api/v1/health
# Should show provider: claude
```

**Back to Ollama**:
```bash
export LLM_PROVIDER=ollama
# Restart backend
```

## Hot-Reload (Optional)

```go
// Watch config file for changes (Viper feature)
v.WatchConfig()
v.OnConfigChange(func(e fsnotify.Event) {
    var cfg LLMConfig
    if err := v.UnmarshalKey("llm", &cfg); err != nil {
        logger.Warnf("failed to reload config: %v", err)
        return
    }
    
    if err := validateLLMConfig(&cfg); err != nil {
        logger.Warnf("invalid reloaded config: %v", err)
        return
    }
    
    // Update global config (don't switch providers dynamically)
    updateLLMConfig(&cfg)
    logger.Info("LLM config reloaded (timeouts, retries)")
})
```

**Caveat**: Can reload operational settings (timeout, retry count), but don't dynamically switch providers — requires careful state management and potential consistency issues. Prefer restart for provider changes.

## Dependencies

```go
// go.mod
require github.com/spf13/viper v1.18.0  // Configuration loading + env binding
```

Optional for unified LLM interface:
```go
require github.com/teilomillet/gollm v0.6.0  // Pre-built provider abstraction (alternative to custom)
```

## Deployment Checklist

- [ ] Create `backend/config/llm-config.yaml` with all providers (one enabled)
- [ ] Add `.env.example` with `LLM_PROVIDER`, `ANTHROPIC_API_KEY`, `OPENAI_API_KEY`
- [ ] Add `.env.local` to `.gitignore`
- [ ] Load config in `main.go` before initializing services
- [ ] Validate config at startup (fail fast)
- [ ] Test connectivity to active provider (warn if unreachable)
- [ ] Log active provider on startup
- [ ] Handle provider switching via LLM_PROVIDER env var + restart

---

**References**:
- [Viper Configuration Library](https://github.com/spf13/viper)
- [Kubernetes Configuration Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [OpenStack Cloud Configuration](https://docs.openstack.org/os-client-config/latest/)
