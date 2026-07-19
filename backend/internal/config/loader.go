package config

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/viper"
)

// LoadMerchantsConfig loads merchant dictionary configuration from environment and config file
func LoadMerchantsConfig() (MerchantsConfig, error) {
	cfg := DefaultMerchantsConfig()

	// Load from environment variables with MERCHANTS_ prefix
	v := viper.New()
	v.SetEnvPrefix("MERCHANTS")
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Bind environment variables
	v.BindEnv("cache_size", "MERCHANTS_CACHE_SIZE")
	v.BindEnv("fuzzy_match_threshold", "MERCHANTS_FUZZY_THRESHOLD")

	// Override with environment values if set
	if cacheSize := v.GetInt("cache_size"); cacheSize > 0 {
		cfg.CacheSize = cacheSize
	}
	if threshold := v.GetFloat64("fuzzy_match_threshold"); threshold > 0 {
		cfg.FuzzyMatchThreshold = threshold
	}

	// Validate configuration
	if err := validateMerchantsConfig(cfg); err != nil {
		return cfg, fmt.Errorf("invalid merchants config: %w", err)
	}

	return cfg, nil
}

func validateMerchantsConfig(cfg MerchantsConfig) error {
	if cfg.FuzzyMatchThreshold < 0 || cfg.FuzzyMatchThreshold > 1 {
		return fmt.Errorf("fuzzy_match_threshold must be between 0 and 1, got %f", cfg.FuzzyMatchThreshold)
	}
	if cfg.CacheSize <= 0 {
		return fmt.Errorf("cache_size must be > 0, got %d", cfg.CacheSize)
	}
	if cfg.MaxConcurrentLookups <= 0 {
		return fmt.Errorf("max_concurrent_lookups must be > 0, got %d", cfg.MaxConcurrentLookups)
	}
	return nil
}

// LogConfig logs configuration values (for debugging, excluding secrets)
func LogConfig(cfg MerchantsConfig) {
	fmt.Fprintf(os.Stderr, "Merchants Config Loaded:\n")
	fmt.Fprintf(os.Stderr, "  Trie Enabled: %v\n", cfg.TrieSettings.Enabled)
	fmt.Fprintf(os.Stderr, "  Fuzzy Match Threshold: %.2f\n", cfg.FuzzyMatchThreshold)
	fmt.Fprintf(os.Stderr, "  Cache Size: %d\n", cfg.CacheSize)
	fmt.Fprintf(os.Stderr, "  Max Concurrent Lookups: %d\n", cfg.MaxConcurrentLookups)
}
