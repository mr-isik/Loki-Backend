package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
)

var (
	ErrNodeRunLogNotFound = errors.New("node run log not found")
)

type NodeRunLogStatus string

const (
	NodeRunLogStatusPending   NodeRunLogStatus = "pending"
	NodeRunLogStatusRunning   NodeRunLogStatus = "running"
	NodeRunLogStatusCompleted NodeRunLogStatus = "completed"
	NodeRunLogStatusFailed    NodeRunLogStatus = "failed"
	NodeRunLogStatusSkipped   NodeRunLogStatus = "skipped"
)

type NodeRunLog struct {
	ID         uuid.UUID        `json:"id"`
	RunID      uuid.UUID        `json:"run_id"`
	NodeID     uuid.UUID        `json:"node_id"`
	Status     NodeRunLogStatus `json:"status"`
	LogOutput  string           `json:"log_output,omitempty"`
	ErrorMsg   string           `json:"error_msg,omitempty"`
	StartedAt  time.Time        `json:"started_at"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

type CreateNodeRunLogRequest struct {
	RunID  uuid.UUID        `json:"run_id" validate:"required,uuid4"`
	NodeID uuid.UUID        `json:"node_id" validate:"required,uuid4"`
	Status NodeRunLogStatus `json:"status" validate:"required"`
}

type UpdateNodeRunLogRequest struct {
	Status    NodeRunLogStatus `json:"status" validate:"omitempty"`
	LogOutput string           `json:"log_output" validate:"omitempty"`
	ErrorMsg  string           `json:"error_msg" validate:"omitempty"`
}

type NodeRunLogResponse struct {
	ID         uuid.UUID        `json:"id"`
	RunID      uuid.UUID        `json:"run_id"`
	NodeID     uuid.UUID        `json:"node_id"`
	Status     NodeRunLogStatus `json:"status"`
	LogOutput  string           `json:"log_output,omitempty"`
	ErrorMsg   string           `json:"error_msg,omitempty"`
	StartedAt  time.Time        `json:"started_at"`
	FinishedAt *time.Time       `json:"finished_at,omitempty"`
	CreatedAt  time.Time        `json:"created_at"`
	UpdatedAt  time.Time        `json:"updated_at"`
}

func (nrl *NodeRunLog) ToResponse() *NodeRunLogResponse {
	return &NodeRunLogResponse{
		ID:         nrl.ID,
		RunID:      nrl.RunID,
		NodeID:     nrl.NodeID,
		Status:     nrl.Status,
		LogOutput:  nrl.LogOutput,
		ErrorMsg:   nrl.ErrorMsg,
		StartedAt:  nrl.StartedAt,
		FinishedAt: nrl.FinishedAt,
		CreatedAt:  nrl.CreatedAt,
		UpdatedAt:  nrl.UpdatedAt,
	}
}

type NodeRunLogRepository interface {
	Create(ctx context.Context, req *CreateNodeRunLogRequest) (*NodeRunLog, error)
	GetByID(ctx context.Context, id uuid.UUID) (*NodeRunLog, error)
	GetByRunID(ctx context.Context, runID uuid.UUID) ([]*NodeRunLog, error)
	Update(ctx context.Context, id uuid.UUID, req *UpdateNodeRunLogRequest) error
}

type NodeRunLogService interface {
	CreateNodeRunLog(ctx context.Context, req *CreateNodeRunLogRequest) (*NodeRunLogResponse, error)
	GetNodeRunLog(ctx context.Context, id uuid.UUID) (*NodeRunLogResponse, error)
	GetNodeRunLogsByRunID(ctx context.Context, runID uuid.UUID) ([]*NodeRunLogResponse, error)
	UpdateNodeRunLog(ctx context.Context, id uuid.UUID, req *UpdateNodeRunLogRequest) error
}
