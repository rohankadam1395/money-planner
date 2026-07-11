package auth

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

// AuthService handles authentication operations
type AuthService struct {
	userRepo *UserRepository
	jwtSecret string
}

func NewAuthService(userRepo *UserRepository, jwtSecret string) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		jwtSecret: jwtSecret,
	}
}

// RegisterRequest contains registration data
type RegisterRequest struct {
	Email    string
	Password string
	FullName string
}

// RegisterResponse contains the result of registration
type RegisterResponse struct {
	UserID string `json:"user_id"`
	Email  string `json:"email"`
	Message string `json:"message,omitempty"`
}

// LoginRequest contains login credentials
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResponse contains the JWT token after login
type LoginResponse struct {
	Token   string `json:"token"`
	UserID  string `json:"user_id"`
	Email   string `json:"email"`
}

// Register creates a new user account
func (as *AuthService) Register(req *RegisterRequest) (*RegisterResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	if len(req.Password) < 6 {
		return nil, fmt.Errorf("password must be at least 6 characters")
	}

	// Check if email already exists
	existing, err := as.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existing != nil {
		return nil, fmt.Errorf("email already registered")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &User{
		Email:        req.Email,
		PasswordHash: string(hashedPassword),
		FullName:     req.FullName,
	}

	if err := as.userRepo.Create(user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &RegisterResponse{
		UserID:  user.UserID.String(),
		Email:   user.Email,
		Message: "Registration successful",
	}, nil
}

// Login authenticates a user and returns a JWT token
func (as *AuthService) Login(req *LoginRequest) (*LoginResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, fmt.Errorf("email and password are required")
	}

	// Get user by email
	user, err := as.userRepo.GetByEmail(req.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %w", err)
	}

	if user == nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, fmt.Errorf("invalid credentials")
	}

	// Update last login
	_ = as.userRepo.UpdateLastLogin(user.UserID)

	// Generate JWT token
	token, err := GenerateToken(user.UserID.String(), as.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	return &LoginResponse{
		Token:  token,
		UserID: user.UserID.String(),
		Email:  user.Email,
	}, nil
}
