package categorization

// ConfidenceScorer computes confidence scores for different categorization methods
// - Exact match: 1.0 (100%)
// - Fuzzy match: 0.85-0.99 (based on Levenshtein distance)
// - Uncategorized: 0.0
type ConfidenceScorer struct {
	exactMatchScore float64
	fuzzyMin        float64
	fuzzyMax        float64
}

// NewConfidenceScorer creates a new confidence scorer
func NewConfidenceScorer() *ConfidenceScorer {
	return &ConfidenceScorer{
		exactMatchScore: 1.0,
		fuzzyMin:        0.85,
		fuzzyMax:        0.99,
	}
}

// ScoreExactMatch returns confidence for exact merchant match
func (cs *ConfidenceScorer) ScoreExactMatch() float64 {
	return cs.exactMatchScore
}

// ScoreFuzzyMatch returns confidence for fuzzy match based on distance (0.0-1.0)
// Maps distance to confidence range [fuzzyMin, fuzzyMax]
func (cs *ConfidenceScorer) ScoreFuzzyMatch(distance float64) float64 {
	if distance < 0 {
		return cs.fuzzyMin
	}
	if distance > 1.0 {
		return cs.fuzzyMax
	}

	// Scale distance (0.85-1.0 range) to confidence (fuzzyMin-fuzzyMax)
	// distance=0.85 → confidence=fuzzyMin
	// distance=1.0 → confidence=fuzzyMax
	scaled := cs.fuzzyMin + (distance-0.85)*(cs.fuzzyMax-cs.fuzzyMin)/(1.0-0.85)
	if scaled < cs.fuzzyMin {
		return cs.fuzzyMin
	}
	if scaled > cs.fuzzyMax {
		return cs.fuzzyMax
	}
	return scaled
}

// ScoreUncategorized returns confidence for uncategorized transactions
func (cs *ConfidenceScorer) ScoreUncategorized() float64 {
	return 0.0
}

// Validate checks if a confidence score is valid (0.0-1.0)
func (cs *ConfidenceScorer) Validate(score float64) bool {
	return score >= 0.0 && score <= 1.0
}
