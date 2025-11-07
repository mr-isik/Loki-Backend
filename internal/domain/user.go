package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)


var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user with this email already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
)

// User represents a user entity
type User struct {
	ID        uuid.UUID  `json:"id"`
	Email     string     `json:"email"`
	Name      string     `json:"name"`
	Password  string     `json:"-"` // Never expose password in JSON
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at,omitempty"`
}

// CreateUserRequest represents the request to create a user
type CreateUserRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Name     string `json:"name" validate:"required,min=2,max=100"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateUserRequest represents the request to update a user
type UpdateUserRequest struct {
	Email string `json:"email,omitempty" validate:"omitempty,email"`
	Name  string `json:"name,omitempty" validate:"omitempty,min=2,max=100"`
}

// UserResponse represents the user response (without sensitive data)
type UserResponse struct {
	ID        uuid.UUID `json:"id"`
	Email     string    `json:"email"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts User to UserResponse
func (u *User) ToResponse() *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		CreatedAt: u.CreatedAt,
		UpdatedAt: u.UpdatedAt,
	}
}

// UserRepository defines the interface for user data operations
type UserRepository interface {
	Create(ctx context.Context, user *User) error
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
	Update(ctx context.Context, user *User) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
}

// UserService defines the interface for user business logic
type UserService interface {
	CreateUser(ctx context.Context, req *CreateUserRequest) (*UserResponse, error)
	GetUser(ctx context.Context, id uuid.UUID) (*UserResponse, error)
	GetUserByEmail(ctx context.Context, email string) (*UserResponse, error)
	UpdateUser(ctx context.Context, id uuid.UUID, req *UpdateUserRequest) (*UserResponse, error)
	DeleteUser(ctx context.Context, id uuid.UUID) error
}
