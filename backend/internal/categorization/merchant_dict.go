package categorization

import (
	"strings"
)

// MerchantDictionary provides fast merchant-to-category lookups using a Trie and cache
type MerchantDictionary struct {
	trie  *Trie
	cache map[string]*CategorizationResult
}

// Trie implements a prefix tree for merchant name matching
type Trie struct {
	root *TrieNode
}

// TrieNode represents a node in the trie
type TrieNode struct {
	children map[rune]*TrieNode
	isEnd    bool
	category string
	merchant string
}

// NewMerchantDictionary creates a new merchant dictionary
func NewMerchantDictionary() *MerchantDictionary {
	return &MerchantDictionary{
		trie: &Trie{
			root: &TrieNode{
				children: make(map[rune]*TrieNode),
			},
		},
		cache: make(map[string]*CategorizationResult),
	}
}

// Insert adds a merchant-category pair to the dictionary
func (md *MerchantDictionary) Insert(merchant string, category string) {
	lower := strings.ToLower(merchant)
	node := md.trie.root

	for _, char := range lower {
		if _, exists := node.children[char]; !exists {
			node.children[char] = &TrieNode{
				children: make(map[rune]*TrieNode),
			}
		}
		node = node.children[char]
	}

	node.isEnd = true
	node.category = category
	node.merchant = merchant
}

// LookupExact performs an exact match lookup (case-insensitive)
func (md *MerchantDictionary) LookupExact(merchant string) *CategorizationResult {
	cacheKey := strings.ToLower(merchant)

	// Check cache first
	if cached, ok := md.cache[cacheKey]; ok {
		return cached
	}

	lower := strings.ToLower(merchant)
	node := md.trie.root

	for _, char := range lower {
		if n, exists := node.children[char]; exists {
			node = n
		} else {
			return nil
		}
	}

	if node.isEnd && node.category != "" {
		result := &CategorizationResult{
			Category:   node.category,
			Method:     "rule_based",
			Confidence: 1.0,
		}
		md.cache[cacheKey] = result
		return result
	}

	return nil
}

// LookupFuzzy performs a fuzzy match lookup using prefix/contains and Levenshtein distance
func (md *MerchantDictionary) LookupFuzzy(merchant string) *CategorizationResult {
	const threshold = 0.85

	lower := strings.ToLower(strings.TrimSpace(merchant))
	if result := md.lookupPrefixOrContains(lower); result != nil {
		return result
	}

	bestMatch := ""
	bestCategory := ""
	bestDistance := 0.0

	md.traverseTrie(md.trie.root, lower, threshold, &bestMatch, &bestCategory, &bestDistance)

	if bestDistance >= threshold && bestCategory != "" {
		return &CategorizationResult{
			Category:      bestCategory,
			Method:        "fuzzy",
			Confidence:    bestDistance,
			matchDistance: bestDistance,
			Reason:        "Fuzzy match: " + bestMatch,
		}
	}

	return nil
}

func (md *MerchantDictionary) lookupPrefixOrContains(target string) *CategorizationResult {
	var bestMatch string
	var bestCategory string

	var walk func(node *TrieNode)
	walk = func(node *TrieNode) {
		if node.isEnd && node.merchant != "" {
			merchantLower := strings.ToLower(node.merchant)
			if strings.Contains(target, merchantLower) || strings.HasPrefix(target, merchantLower) {
				if len(merchantLower) > len(bestMatch) {
					bestMatch = node.merchant
					bestCategory = node.category
				}
			}
		}
		for _, child := range node.children {
			walk(child)
		}
	}
	walk(md.trie.root)

	if bestCategory == "" {
		return nil
	}

	return &CategorizationResult{
		Category:      bestCategory,
		Method:        "fuzzy",
		matchDistance: 0.9,
		Reason:        "Fuzzy match: " + bestMatch,
	}
}

// traverseTrie recursively traverses the trie to find fuzzy matches
func (md *MerchantDictionary) traverseTrie(node *TrieNode, target string, threshold float64, bestMatch *string, bestCategory *string, bestDistance *float64) {
	if node.isEnd && node.merchant != "" {
		dist := levenshteinDistance(strings.ToLower(node.merchant), target)
		if dist > *bestDistance && dist >= threshold {
			*bestDistance = dist
			*bestMatch = node.merchant
			*bestCategory = node.category
		}
	}

	for _, child := range node.children {
		md.traverseTrie(child, target, threshold, bestMatch, bestCategory, bestDistance)
	}
}

// levenshteinDistance computes normalized Levenshtein distance (0.0-1.0)
// 1.0 = perfect match, 0.0 = completely different
func levenshteinDistance(s1, s2 string) float64 {
	maxLen := len(s1)
	if len(s2) > maxLen {
		maxLen = len(s2)
	}

	if maxLen == 0 {
		return 1.0
	}

	distance := computeLevenshtein(s1, s2)
	return 1.0 - float64(distance)/float64(maxLen)
}

// computeLevenshtein computes raw Levenshtein edit distance
func computeLevenshtein(s1, s2 string) int {
	if len(s1) == 0 {
		return len(s2)
	}
	if len(s2) == 0 {
		return len(s1)
	}

	prev := make([]int, len(s2)+1)
	curr := make([]int, len(s2)+1)

	for j := range prev {
		prev[j] = j
	}

	for i := 1; i <= len(s1); i++ {
		curr[0] = i

		for j := 1; j <= len(s2); j++ {
			cost := 0
			if s1[i-1] != s2[j-1] {
				cost = 1
			}

			curr[j] = min(
				curr[j-1]+1,    // insertion
				prev[j]+1,      // deletion
				prev[j-1]+cost, // substitution
			)
		}

		prev, curr = curr, prev
	}

	return prev[len(s2)]
}

func min(a, b, c int) int {
	if a < b && a < c {
		return a
	}
	if b < c {
		return b
	}
	return c
}

// Clear resets the cache
func (md *MerchantDictionary) Clear() {
	md.cache = make(map[string]*CategorizationResult)
}
