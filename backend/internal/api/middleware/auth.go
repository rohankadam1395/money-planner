package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

const (
	UserIDContextKey = "user_id"
	BearerScheme     = "Bearer"
)

type AuthMiddleware struct {
	jwtSecret string
}

func NewAuthMiddleware(jwtSecret string) *AuthMiddleware {
	return &AuthMiddleware{
		jwtSecret: jwtSecret,
	}
}

func (m *AuthMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != BearerScheme {
			http.Error(w, "invalid authorization header format", http.StatusUnauthorized)
			return
		}

		tokenString := parts[1]
		userID, err := m.validateToken(tokenString)
		if err != nil {
			http.Error(w, fmt.Sprintf("invalid token: %v", err), http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), UserIDContextKey, userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func (m *AuthMiddleware) validateToken(tokenString string) (string, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(m.jwtSecret), nil
	})

	if err != nil {
		return "", fmt.Errorf("failed to parse token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return "", fmt.Errorf("invalid token claims")
	}

	return claims.UserID, nil
}

type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func GetUserID(r *http.Request) (string, error) {
	userID := r.Context().Value(UserIDContextKey)
	if userID == nil {
		return "", fmt.Errorf("user_id not found in context")
	}

	userIDStr, ok := userID.(string)
	if !ok {
		return "", fmt.Errorf("user_id is not a string")
	}

	return userIDStr, nil
}
