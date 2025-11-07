package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

var (
	ErrWorkflowEdgeNotFound = errors.New("workflow edge not found")
)

type WorkflowEdge struct {
		ID           uuid.UUID `json:"id"`
		WorkflowID   uuid.UUID `json:"workflow_id"`
		SourceNodeID uuid.UUID `json:"source_node_id"`
		TargetNodeID uuid.UUID `json:"target_node_id"`
		SourceHandle string    `json:"source_handle"`
		TargetHandle string    `json:"target_handle"`
}

type CreateWorkflowEdgeRequest struct {
	SourceNodeID uuid.UUID `json:"source_node_id" validate:"required,uuid4"`
	TargetNodeID uuid.UUID `json:"target_node_id" validate:"required,uuid4"`
	SourceHandle string    `json:"source_handle" validate:"required"`
	TargetHandle string    `json:"target_handle" validate:"required"`
}

type UpdateWorkflowEdgeRequest struct {
	SourceNodeID uuid.UUID `json:"source_node_id" validate:"omitempty,uuid4"`
	TargetNodeID uuid.UUID `json:"target_node_id" validate:"omitempty,uuid4"`
	SourceHandle string `json:"source_handle" validate:"omitempty"`
	TargetHandle string `json:"target_handle" validate:"omitempty"`
}

type WorkflowEdgeResponse struct {
	ID           uuid.UUID `json:"id"`
	WorkflowID   uuid.UUID `json:"workflow_id"`
	SourceNodeID uuid.UUID `json:"source_node_id"`
	TargetNodeID uuid.UUID `json:"target_node_id"`
	SourceHandle string `json:"source_handle"`
	TargetHandle string `json:"target_handle"`
}

func (we *WorkflowEdge) ToResponse() *WorkflowEdgeResponse {
	return &WorkflowEdgeResponse{
		ID:           we.ID,
		WorkflowID:   we.WorkflowID,
		SourceNodeID: we.SourceNodeID,
		TargetNodeID: we.TargetNodeID,
		SourceHandle: we.SourceHandle,
		TargetHandle: we.TargetHandle,
	}
}

type WorkflowEdgeRepository interface {
	Create(ctx context.Context, edge *CreateWorkflowEdgeRequest) error
	Update(ctx context.Context, id uuid.UUID, edge *UpdateWorkflowEdgeRequest) error
	GetByID(ctx context.Context, id uuid.UUID) (*WorkflowEdge, error)
	GetByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*WorkflowEdge, error)
	Delete(ctx context.Context, id uuid.UUID) error
}