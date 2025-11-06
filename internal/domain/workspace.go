package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkspaceNotFound = errors.New("workspace not found")
	ErrUnauthorized      = errors.New("unauthorized: user is not the owner")
)

type Workspace struct {
	ID          uuid.UUID `json:"id"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
}

type CreateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type UpdateWorkspaceRequest struct {
	Name string `json:"name" validate:"required,min=2,max=100"`
}

type WorkspaceResponse struct {
	ID          uuid.UUID `json:"id"`
	OwnerUserID uuid.UUID `json:"owner_user_id"`
	Name        string    `json:"name"`
	CreatedAt   time.Time `json:"created_at"`
}

func (w *Workspace) ToResponse() *WorkspaceResponse {
	return &WorkspaceResponse{
		ID:          w.ID,
		OwnerUserID: w.OwnerUserID,
		Name:        w.Name,
		CreatedAt:   w.CreatedAt,
	}
}

type WorkspaceRepository interface {
	Create(ctx context.Context, workspace *Workspace) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workspace, error)
	GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*Workspace, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Workspace, error)
	Update(ctx context.Context, workspace *Workspace) error
	Delete(ctx context.Context, id uuid.UUID) error
	Count(ctx context.Context) (int64, error)
	IsOwner(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error)
}

type WorkspaceService interface {
	CreateWorkspace(ctx context.Context, ownerID uuid.UUID, req *CreateWorkspaceRequest) (*WorkspaceResponse, error)
	GetWorkspace(ctx context.Context, id uuid.UUID) (*WorkspaceResponse, error)
	GetUserWorkspaces(ctx context.Context, userID uuid.UUID) ([]*WorkspaceResponse, error)
	ListWorkspaces(ctx context.Context, page, pageSize int) ([]*WorkspaceResponse, int64, error)
	UpdateWorkspace(ctx context.Context, id, userID uuid.UUID, req *UpdateWorkspaceRequest) (*WorkspaceResponse, error)
	DeleteWorkspace(ctx context.Context, id, userID uuid.UUID) error
}
