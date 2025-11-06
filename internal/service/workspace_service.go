package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type workspaceService struct {
	repo domain.WorkspaceRepository
}

// NewWorkspaceService creates a new workspace service
func NewWorkspaceService(repo domain.WorkspaceRepository) domain.WorkspaceService {
	return &workspaceService{
		repo: repo,
	}
}

// CreateWorkspace creates a new workspace
func (s *workspaceService) CreateWorkspace(ctx context.Context, ownerID uuid.UUID, req *domain.CreateWorkspaceRequest) (*domain.WorkspaceResponse, error) {
	workspace := &domain.Workspace{
		OwnerUserID: ownerID,
		Name:        req.Name,
	}

	if err := s.repo.Create(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to create workspace: %w", err)
	}

	return workspace.ToResponse(), nil
}

// GetWorkspace retrieves a workspace by ID
func (s *workspaceService) GetWorkspace(ctx context.Context, id uuid.UUID) (*domain.WorkspaceResponse, error) {
	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	return workspace.ToResponse(), nil
}

// GetUserWorkspaces retrieves all workspaces owned by a user
func (s *workspaceService) GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*domain.WorkspaceResponse, error) {
	workspaces, err := s.repo.GetByOwnerID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user workspaces: %w", err)
	}

	responses := make([]*domain.WorkspaceResponse, len(workspaces))
	for i, workspace := range workspaces {
		responses[i] = workspace.ToResponse()
	}

	return responses, nil
}

// ListWorkspaces retrieves workspaces with pagination
func (s *workspaceService) ListWorkspaces(ctx context.Context, page, pageSize int) ([]*domain.WorkspaceResponse, int64, error) {
	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	// Get workspaces
	workspaces, err := s.repo.GetAll(ctx, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list workspaces: %w", err)
	}

	// Get total count
	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count workspaces: %w", err)
	}

	// Convert to response
	responses := make([]*domain.WorkspaceResponse, len(workspaces))
	for i, workspace := range workspaces {
		responses[i] = workspace.ToResponse()
	}

	return responses, total, nil
}

// UpdateWorkspace updates a workspace
func (s *workspaceService) UpdateWorkspace(ctx context.Context, id, userID uuid.UUID, req *domain.UpdateWorkspaceRequest) (*domain.WorkspaceResponse, error) {
	// Check if user is the owner
	isOwner, err := s.repo.IsOwner(ctx, id, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check ownership: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrUnauthorized
	}

	// Get existing workspace
	workspace, err := s.repo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return nil, domain.ErrWorkspaceNotFound
		}
		return nil, fmt.Errorf("failed to get workspace: %w", err)
	}

	// Update fields
	workspace.Name = req.Name

	// Update workspace
	if err := s.repo.Update(ctx, workspace); err != nil {
		return nil, fmt.Errorf("failed to update workspace: %w", err)
	}

	return workspace.ToResponse(), nil
}

// DeleteWorkspace deletes a workspace
func (s *workspaceService) DeleteWorkspace(ctx context.Context, id, userID uuid.UUID) error {
	// Check if user is the owner
	isOwner, err := s.repo.IsOwner(ctx, id, userID)
	if err != nil {
		return fmt.Errorf("failed to check ownership: %w", err)
	}
	if !isOwner {
		return domain.ErrUnauthorized
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		if errors.Is(err, domain.ErrWorkspaceNotFound) {
			return domain.ErrWorkspaceNotFound
		}
		return fmt.Errorf("failed to delete workspace: %w", err)
	}

	return nil
}
