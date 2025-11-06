package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrUserExists         = errors.New("user already exists")
)

type userService struct {
	repo domain.UserRepository
}

// NewUserService creates a new user service
func NewUserService(repo domain.UserRepository) domain.UserService {
	return &userService{
		repo: repo,
	}
}

// CreateUser creates a new user
func (s *userService) CreateUser(ctx context.Context, req *domain.CreateUserRequest) (*domain.UserResponse, error) {
	existingUser, err := s.repo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		return nil, fmt.Errorf("failed to check existing user: %w", err)
	}
	if existingUser != nil {
		return nil, ErrUserExists
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	user := &domain.User{
		Email:    req.Email,
		Name:     req.Name,
		Password: string(hashedPassword),
	}

	if err := s.repo.Create(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return user.ToResponse(), nil
}

func (s *userService) GetUser(ctx context.Context, id uuid.UUID) (*domain.UserResponse, error) {
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToResponse(), nil
}

func (s *userService) GetUserByEmail(ctx context.Context, email string) (*domain.UserResponse, error) {
	user, err := s.repo.GetByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	return user.ToResponse(), nil
}

// UpdateUser updates a user
func (s *userService) UpdateUser(ctx context.Context, id uuid.UUID, req *domain.UpdateUserRequest) (*domain.UserResponse, error) {
	// Get existing user
	user, err := s.repo.GetByID(ctx, id)
	if err != nil {	
		if errors.Is(err, domain.ErrUserNotFound) {
			return nil, domain.ErrUserNotFound
		}
		return nil, fmt.Errorf("failed to get user: %w", err)
	}

	// Update fields if provided
	if req.Email != "" {
		// Check if new email is already taken by another user
		existingUser, err := s.repo.GetByEmail(ctx, req.Email)
		if err != nil && !errors.Is(err, domain.ErrUserNotFound) {
			return nil, fmt.Errorf("failed to check email: %w", err)
		}
		if existingUser != nil && existingUser.ID != id {
			return nil, ErrUserExists
		}
		user.Email = req.Email
	}

	if req.Name != "" {
		user.Name = req.Name
	}

	// Update user
	if err := s.repo.Update(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	return user.ToResponse(), nil
}

// DeleteUser deletes a user
func (s *userService) DeleteUser(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return domain.ErrUserNotFound
		}
		return fmt.Errorf("failed to delete user: %w", err)
	}

	return nil
}
