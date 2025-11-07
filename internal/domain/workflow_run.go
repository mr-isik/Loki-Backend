package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrWorkflowRunNotFound = errors.New("workflow run not found")
)

type WorkflowRunStatus string

const (
	WorkflowRunStatusPending   WorkflowRunStatus = "pending"
	WorkflowRunStatusRunning   WorkflowRunStatus = "running"
	WorkflowRunStatusCompleted WorkflowRunStatus = "completed"
	WorkflowRunStatusFailed    WorkflowRunStatus = "failed"
	WorkflowRunStatusCancelled WorkflowRunStatus = "cancelled"
)

type WorkflowRun struct {
	ID         uuid.UUID         `json:"id"`
	WorkflowID uuid.UUID         `json:"workflow_id"`
	Status     WorkflowRunStatus `json:"status"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt *time.Time        `json:"finished_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

type CreateWorkflowRunRequest struct {
	WorkflowID uuid.UUID `json:"workflow_id" validate:"required,uuid4"`
}

type WorkflowRunResponse struct {
	ID         uuid.UUID         `json:"id"`
	WorkflowID uuid.UUID         `json:"workflow_id"`
	Status     WorkflowRunStatus `json:"status"`
	StartedAt  time.Time         `json:"started_at"`
	FinishedAt *time.Time        `json:"finished_at,omitempty"`
	CreatedAt  time.Time         `json:"created_at"`
	UpdatedAt  time.Time         `json:"updated_at"`
}

func (wr *WorkflowRun) ToResponse() *WorkflowRunResponse {
	return &WorkflowRunResponse{
		ID:         wr.ID,
		WorkflowID: wr.WorkflowID,
		Status:     wr.Status,
		StartedAt:  wr.StartedAt,
		FinishedAt: wr.FinishedAt,
		CreatedAt:  wr.CreatedAt,
		UpdatedAt:  wr.UpdatedAt,
	}
}

type WorkflowRunRepository interface {
	Create(ctx context.Context, workflowID uuid.UUID) (*WorkflowRun, error)
	GetByID(ctx context.Context, id uuid.UUID) (*WorkflowRun, error)
	UpdateStatus(ctx context.Context, id uuid.UUID, status WorkflowRunStatus, finishedAt *time.Time) error
	ListByWorkflowID(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]*WorkflowRun, int, error)
}

type WorkflowRunService interface {
	StartWorkflowRun(ctx context.Context, workflowID uuid.UUID) (*WorkflowRunResponse, error)
	GetWorkflowRun(ctx context.Context, id uuid.UUID) (*WorkflowRunResponse, error)
	ListWorkflowRuns(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]*WorkflowRunResponse, int, error)
	UpdateRunStatus(ctx context.Context, id uuid.UUID, status WorkflowRunStatus) error
}