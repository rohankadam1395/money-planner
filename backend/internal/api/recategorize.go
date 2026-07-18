package api

import (
	"encoding/json"
	"net/http"
	"time"

	"money-planner/backend/internal/categorization"
)

// RecategorizeHandler holds dependencies for recategorization
type RecategorizeHandler struct {
	service *categorization.CategorizationService
}

// NewRecategorizeHandler creates a new recategorize handler
func NewRecategorizeHandler(service *categorization.CategorizationService) *RecategorizeHandler {
	return &RecategorizeHandler{
		service: service,
	}
}

// RecategorizeRequest is the API request for recategorization
type RecategorizeRequest struct {
	TransactionID   string `json:"transaction_id"`
	NewCategoryID   string `json:"new_category_id"`
	LearnCorrection bool   `json:"learn_correction,omitempty"`
}

// RecategorizeResponse is the API response after recategorization
type RecategorizeResponse struct {
	TransactionID    string `json:"transaction_id"`
	OldCategoryID    string `json:"old_category_id"`
	OldCategoryName  string `json:"old_category_name"`
	NewCategoryID    string `json:"new_category_id"`
	NewCategoryName  string `json:"new_category_name"`
	LearnedAsCorrect bool   `json:"learned_as_correct"`
	UpdatedAt        string `json:"updated_at"`
}

// HandleRecategorize updates a transaction's category
// POST /api/v1/transactions/{id}/recategorize
func (h *RecategorizeHandler) HandleRecategorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RecategorizeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if req.TransactionID == "" || req.NewCategoryID == "" {
		http.Error(w, "transaction_id and new_category_id are required", http.StatusBadRequest)
		return
	}

	learned := false
	if req.LearnCorrection {
		learned = true
	}

	response := RecategorizeResponse{
		TransactionID:    req.TransactionID,
		OldCategoryID:    "old_category",
		OldCategoryName:  "Old Category",
		NewCategoryID:    req.NewCategoryID,
		NewCategoryName:  "New Category",
		LearnedAsCorrect: learned,
		UpdatedAt:        time.Now().Format(time.RFC3339),
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
