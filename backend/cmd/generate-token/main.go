package main

import (
	"fmt"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Claims represents JWT claims
type Claims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

func main() {
	// Use JWT_SECRET from environment or default for testing
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		secret = "your_jwt_secret_key_here_change_in_production"
		fmt.Println("⚠️  Using default JWT secret (not secure for production)")
	}

	// Generate a test user ID
	userID := uuid.New().String()

	// Create claims with 24-hour expiration
	claims := Claims{
		UserID: userID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	// Create and sign the token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(secret))
	if err != nil {
		fmt.Printf("Error generating token: %v\n", err)
		os.Exit(1)
	}

	fmt.Println("✓ JWT Token Generated Successfully")
	fmt.Println()
	fmt.Printf("User ID:  %s\n", userID)
	fmt.Printf("Expires:  %s\n", time.Now().Add(24*time.Hour).Format(time.RFC3339))
	fmt.Println()
	fmt.Println("Authorization Header:")
	fmt.Printf("Authorization: Bearer %s\n", tokenString)
	fmt.Println()
	fmt.Println("curl command with auth:")
	fmt.Printf("curl -X POST http://localhost:8080/api/v1/transactions/categorize \\\n")
	fmt.Printf("  -H \"Content-Type: application/json\" \\\n")
	fmt.Printf("  -H \"Authorization: Bearer %s\" \\\n", tokenString)
	fmt.Printf("  -d '{\"transactions\": [{\"id\": \"1\", \"merchant\": \"Swiggy\", \"amount\": 500, \"timestamp\": 0}]}'\n")
}
