package config

// MerchantsConfig holds merchant dictionary configuration
type MerchantsConfig struct {
	TrieSettings         TrieSettings `mapstructure:"trie"`
	FuzzyMatchThreshold  float64      `mapstructure:"fuzzy_match_threshold"`
	CacheSize            int          `mapstructure:"cache_size"`
	MaxConcurrentLookups int          `mapstructure:"max_concurrent_lookups"`
}

// TrieSettings holds trie-specific configuration
type TrieSettings struct {
	Enabled       bool   `mapstructure:"enabled"`
	CaseSensitive bool   `mapstructure:"case_sensitive"`
	MaxPrefixLen  int    `mapstructure:"max_prefix_len"`
	Timeout       string `mapstructure:"timeout"`
}

// DefaultMerchantsConfig returns default merchant configuration
func DefaultMerchantsConfig() MerchantsConfig {
	return MerchantsConfig{
		TrieSettings: TrieSettings{
			Enabled:       true,
			CaseSensitive: false,
			MaxPrefixLen:  20,
			Timeout:       "5s",
		},
		FuzzyMatchThreshold:  0.85,
		CacheSize:           10000,
		MaxConcurrentLookups: 100,
	}
}
