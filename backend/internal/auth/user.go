package auth

import (
	"database/sql"
	"time"

	"github.com/google/uuid"
)

// User represents a user account
type User struct {
	UserID       uuid.UUID
	Email        string
	PasswordHash string
	FullName     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
	LastLoginAt  *time.Time
}

// UserRepository handles database operations for users
type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

// Create creates a new user
func (ur *UserRepository) Create(user *User) error {
	query := `
		INSERT INTO users (user_id, email, password_hash, full_name, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	user.UserID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := ur.db.Exec(query,
		user.UserID, user.Email, user.PasswordHash, user.FullName,
		user.CreatedAt, user.UpdatedAt,
	)

	return err
}

// GetByEmail fetches a user by email
func (ur *UserRepository) GetByEmail(email string) (*User, error) {
	query := `
		SELECT user_id, email, password_hash, full_name, created_at, updated_at, last_login_at
		FROM users
		WHERE email = $1
	`

	user := &User{}
	err := ur.db.QueryRow(query, email).Scan(
		&user.UserID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// GetByID fetches a user by ID
func (ur *UserRepository) GetByID(userID uuid.UUID) (*User, error) {
	query := `
		SELECT user_id, email, password_hash, full_name, created_at, updated_at, last_login_at
		FROM users
		WHERE user_id = $1
	`

	user := &User{}
	err := ur.db.QueryRow(query, userID).Scan(
		&user.UserID, &user.Email, &user.PasswordHash, &user.FullName,
		&user.CreatedAt, &user.UpdatedAt, &user.LastLoginAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return user, nil
}

// UpdateLastLogin updates the last login timestamp
func (ur *UserRepository) UpdateLastLogin(userID uuid.UUID) error {
	query := `
		UPDATE users
		SET last_login_at = $1, updated_at = $2
		WHERE user_id = $3
	`

	now := time.Now()
	_, err := ur.db.Exec(query, now, now, userID)
	return err
}
