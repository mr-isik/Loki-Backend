package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

type WorkflowStatus string

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
)

const (
	WorkflowStatusDraft     WorkflowStatus = "draft"
	WorkflowStatusPublished WorkflowStatus = "published"
	WorkflowStatusArchived  WorkflowStatus = "archived"
)

type Workflow struct {
	ID          uuid.UUID      `json:"id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Title       string         `json:"title"`
	Status      WorkflowStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// CreateWorkflowRequest represents the request to create a workflow
type CreateWorkflowRequest struct {
	Title string `json:"title" validate:"omitempty,max=255"`
}

// UpdateWorkflowRequest represents the request to update a workflow
type UpdateWorkflowRequest struct {
	Title  string         `json:"title,omitempty" validate:"omitempty,max=255"`
	Status WorkflowStatus `json:"status,omitempty" validate:"omitempty,oneof=draft published archived"`
}

// WorkflowResponse represents the workflow response
type WorkflowResponse struct {
	ID          uuid.UUID      `json:"id"`
	WorkspaceID uuid.UUID      `json:"workspace_id"`
	Title       string         `json:"title"`
	Status      WorkflowStatus `json:"status"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
}

// ToResponse converts Workflow to WorkflowResponse
func (w *Workflow) ToResponse() *WorkflowResponse {
	return &WorkflowResponse{
		ID:          w.ID,
		WorkspaceID: w.WorkspaceID,
		Title:       w.Title,
		Status:      w.Status,
		CreatedAt:   w.CreatedAt,
		UpdatedAt:   w.UpdatedAt,
	}
}

type WorkflowRepository interface {
	Create(ctx context.Context, workflow *Workflow) error
	GetByID(ctx context.Context, id uuid.UUID) (*Workflow, error)
	GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*Workflow, error)
	GetAll(ctx context.Context, limit, offset int) ([]*Workflow, error)
	Update(ctx context.Context, workflow *Workflow) error
	Delete(ctx context.Context, id uuid.UUID) error
	CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status WorkflowStatus) error
}

type WorkflowService interface {
	CreateWorkflow(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, req *CreateWorkflowRequest) (*WorkflowResponse, error)
	GetWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*WorkflowResponse, error)
	GetWorkspaceWorkflows(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, page, pageSize int) ([]*WorkflowResponse, int64, error)
	UpdateWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *UpdateWorkflowRequest) (*WorkflowResponse, error)
	DeleteWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) error
	PublishWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*WorkflowResponse, error)
	ArchiveWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*WorkflowResponse, error)
}