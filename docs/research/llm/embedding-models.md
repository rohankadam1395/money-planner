# Embedding Models Research

**Source**: Agent research on embedding models for transaction analysis  
**Date**: 2026-07-18  
**Status**: Deferred to Phase 5+

## Summary

Embedding models (semantic vectors) are **NOT needed for Phase 002 (transaction categorization MVP)**. They become valuable in Phase 5+ for anomaly detection, semantic search, and merchant clustering.

## What Are Embeddings?

Embedding models convert text into fixed-size vectors (e.g., 768 dimensions) that capture semantic meaning. Similar concepts have similar vectors.

```
"Swiggy Food Delivery" → [0.12, -0.45, 0.89, ...768 values...]
"Food ordering service" → [0.11, -0.44, 0.88, ...768 values...]
# Vectors are close (high cosine similarity ≈ 0.98)

"Uber Ride" → [-0.45, 0.23, 0.12, ...768 values...]
# Vector is far from Swiggy (low cosine similarity ≈ 0.15)
```

## When Embeddings ARE Useful

### 1. Anomaly Detection (Phase 007)

**Use Case**: Detect unusual transactions
```
"Detect transactions that deviate from user's spending patterns"

1. Embed all historical transactions: "Swiggy ₹450", "Amazon ₹2500", "Uber ₹150"
2. Embed current transaction: "Unknown vendor ₹50000"
3. Calculate distance to normal pattern
4. Flag if distance > threshold
```

**Example Implementation**:
```go
// Calculate centroid of "normal" transactions in category
normalCentroid := averageEmbedding(userHistoricalFood)

// Current transaction embedding
currentVector := embed("Unknown restaurant ₹50000")

// Distance from normal
distance := cosineDistance(currentVector, normalCentroid)
if distance > 0.8 {
    // Anomalous: ₹50000 is unusual for food
    flagAsAnomaly()
}
```

### 2. Semantic Search (Phase 005+)

**Use Case**: Find all food-related merchants
```
"Find merchants semantically similar to 'restaurant'"

1. Embed search query: "restaurant" → vector
2. Embed all merchants: "Swiggy", "Zomato", "Dominios", etc. → vectors
3. Find top-K nearest neighbors (highest cosine similarity)
4. Return: [Swiggy 0.95, Zomato 0.94, Dominios 0.92, ...]
```

### 3. Merchant Clustering (Phase 5+)

**Use Case**: Unsupervised grouping of similar merchants
```
"Group merchants without pre-defined categories"

1. Embed all merchants
2. Use clustering algorithm (K-means, DBSCAN)
3. Discover natural groups: food, transport, utilities, etc.
4. Use clusters to suggest new categories
```

### 4. Transaction Similarity

**Use Case**: Find transactions similar to a given one
```
"Find all transactions like 'Swiggy ₹450'"

1. Embed user's transaction
2. Find nearest neighbors in user's historical data
3. Show similar transactions (recurring patterns, habits)
4. Enable "spend more like this" or "spend less on this" features
```

## When Embeddings Are NOT Needed

### Phase 002 (Transaction Categorization)

❌ **Don't use embeddings**:
- Classification task is simple (10 categories)
- LLM (Ollama/Claude) encodes semantics internally
- Embedding lookup would be slower than direct text classification
- Adds storage overhead (vector for each transaction)

✅ **Use direct text classification** (current approach):
```
Merchant: "Swiggy"
→ Direct LLM inference: "Food & Dining"
```

### Phase 003-004 (Dashboard, Budget Planning)

❌ **Embeddings overkill**:
- Basic analytics (sums, counts, filters) don't need semantic matching
- Rule-based budget rules sufficient

### Phase 006 (Forecasting)

⚠️ **Maybe**: Time-series models with semantic features
- Hybrid: Use embeddings as input features to LSTM/transformer
- But: Not required; pure time-series (past spending patterns) often sufficient

## Embedding Models Comparison

| Model | Dimensions | Latency | VRAM | Open-Source | Cost |
|-------|-----------|---------|------|-------------|------|
| **Ollama (snowflake-arctic-embed)** | 768 | 50-100ms | 2GB | ✅ Yes | $0 |
| OpenAI ada-002 | 1536 | 200ms | N/A (API) | ❌ No | $0.02 per 1M |
| Sentence-Transformers (BERT) | 384-768 | 50-150ms | 1-3GB | ✅ Yes | $0 |
| Cohere | 1024 | 100ms | N/A (API) | ❌ No | $0.10-1.00 per 1M |

**For Money Planner**:
- **Best local**: Ollama `snowflake-arctic-embed` (small, fast, free)
- **Best cloud**: OpenAI ada-002 (high quality, widely available)

## Implementation Pattern (for Phase 5+)

```go
// Embedding service (similar to LLM provider abstraction)
type EmbeddingProvider interface {
    Embed(ctx context.Context, text string) ([]float64, error)
    EmbedBatch(ctx context.Context, texts []string) ([][]float64, error)
}

// Usage in anomaly detection
func detectAnomaly(
    ctx context.Context,
    embeddingProvider EmbeddingProvider,
    historicalTransactions []Transaction,
    currentTransaction Transaction,
) (bool, error) {
    // 1. Embed historical transactions
    historicalTexts := make([]string, len(historicalTransactions))
    for i, tx := range historicalTransactions {
        historicalTexts[i] = fmt.Sprintf("%s %f", tx.Merchant, tx.Amount)
    }
    
    historicalEmbeddings, err := embeddingProvider.EmbedBatch(ctx, historicalTexts)
    if err != nil {
        return false, err
    }
    
    // 2. Embed current transaction
    currentText := fmt.Sprintf("%s %f", currentTransaction.Merchant, currentTransaction.Amount)
    currentEmbedding, err := embeddingProvider.Embed(ctx, currentText)
    if err != nil {
        return false, err
    }
    
    // 3. Calculate average embedding (centroid)
    centroid := averageEmbedding(historicalEmbeddings)
    
    // 4. Calculate distance
    distance := cosineDistance(currentEmbedding, centroid)
    
    // 5. Threshold (tunable)
    return distance > 0.7, nil
}

// Helper: cosine similarity
func cosineDistance(a, b []float64) float64 {
    dotProduct := 0.0
    normA, normB := 0.0, 0.0
    
    for i := range a {
        dotProduct += a[i] * b[i]
        normA += a[i] * a[i]
        normB += b[i] * b[i]
    }
    
    return dotProduct / (math.Sqrt(normA) * math.Sqrt(normB))
}
```

## Storage Considerations

**Embedding storage overhead**:
```
Per transaction:
- Merchant: 50 bytes (avg)
- Amount: 8 bytes
- Embedding (768 dims × 4 bytes float32): 3,072 bytes

1M transactions: ~3GB (embeddings only)
```

**When to store embeddings**:
- ✅ Phase 005+: If implementing anomaly detection, semantic search
- ❌ Phase 002: Not needed, compute on-demand if rare queries

## Recommendation

**For Money Planner MVP (Phase 002)**:
- ✅ Skip embeddings
- ✅ Use direct LLM classification (Ollama/Claude)
- ✅ Defer until Phase 005+ when anomaly detection needed

**Future (Phase 005+)**:
- Consider `ollama embed` (free, local) or OpenAI ada-002
- Use for anomaly detection, semantic search
- Follow same provider abstraction pattern as LLM

**Budget**:
- Ollama embeddings: Free (local)
- OpenAI embeddings: ~$0.02 per 1M tokens (cheap)

---

## References

- [Sentence-Transformers](https://www.sbert.net/)
- [OpenAI Embeddings API](https://platform.openai.com/docs/guides/embeddings)
- [Ollama Embedding Models](https://ollama.ai/library?sort=trending&search=embed)
- [Cosine Similarity](https://en.wikipedia.org/wiki/Cosine_similarity)
- [FAISS (Similarity Search Library)](https://github.com/facebookresearch/faiss)
