package service

import (
	"context"
	"errors"
	"time"

	"github.com/mr-isik/loki-backend/internal/domain"
	"github.com/mr-isik/loki-backend/internal/util"
)

type authService struct {
	userRepo   domain.UserRepository
	jwtManager *util.JWTManager
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo domain.UserRepository, jwtManager *util.JWTManager) domain.AuthService {
	return &authService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

// Register handles user registration
func (s *authService) Register(ctx context.Context, req *domain.RegisterRequest) (*domain.RegisterResponse, error) {
	// Check if user already exists
	existingUser, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if existingUser != nil {
		return nil, domain.ErrUserAlreadyExists
	}

	// Hash password
	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &domain.User{
		Email:     req.Email,
		Name:      req.Name,
		Password:  hashedPassword,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := s.userRepo.Create(ctx, user); err != nil {
		return nil, err
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	return &domain.RegisterResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Login handles user authentication
func (s *authService) Login(ctx context.Context, req *domain.LoginRequest) (*domain.LoginResponse, error) {
	// Get user by email
	user, err := s.userRepo.GetByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}

	// Verify password
	if err := util.CheckPassword(user.Password, req.Password); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	// Generate tokens
	accessToken, err := s.jwtManager.GenerateAccessToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	refreshToken, err := s.jwtManager.GenerateRefreshToken(user.ID, user.Email, user.Name)
	if err != nil {
		return nil, err
	}

	return &domain.LoginResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// RefreshToken refreshes the access token using a valid refresh token
func (s *authService) RefreshToken(ctx context.Context, req *domain.RefreshTokenRequest) (*domain.RefreshTokenResponse, error) {

	accessToken, err := s.jwtManager.RefreshAccessToken(req.RefreshToken)
	if err != nil {
		return nil, domain.ErrInvalidRefreshToken
	}

	return &domain.RefreshTokenResponse{
		AccessToken: accessToken,
	}, nil
}
