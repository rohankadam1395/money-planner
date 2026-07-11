package main

import (
	"encoding/json"
	"net/http"

	"money-planner/backend/internal/api/middleware"
	"money-planner/backend/internal/auth"
)

// RegisterHandler handles POST /api/auth/register
func RegisterHandler(authService *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
			return
		}

		var req auth.RegisterRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body", "INVALID_REQUEST")
			return
		}

		resp, err := authService.Register(&req)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusBadRequest, err.Error(), "REGISTRATION_FAILED")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}
}

// LoginHandler handles POST /api/auth/login
func LoginHandler(authService *auth.AuthService) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			middleware.WriteJSONError(w, http.StatusMethodNotAllowed, "method not allowed", "METHOD_NOT_ALLOWED")
			return
		}

		var req auth.LoginRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			middleware.WriteJSONError(w, http.StatusBadRequest, "invalid request body", "INVALID_REQUEST")
			return
		}

		resp, err := authService.Login(&req)
		if err != nil {
			middleware.WriteJSONError(w, http.StatusUnauthorized, err.Error(), "LOGIN_FAILED")
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
