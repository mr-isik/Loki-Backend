package domain

import (
	"context"

	"github.com/google/uuid"
)

type WorkflowNode struct {
	ID         uuid.UUID      `json:"id"`
	WorkflowID uuid.UUID      `json:"workflow_id"`
	TemplateID uuid.UUID      `json:"template_id"`
	PositionX  float64        `json:"position_x"`
	PositionY  float64        `json:"position_y"`
	Data       map[string]any `json:"data"`
}

type CreateWorkflowNodeRequest struct {
	WorkflowID uuid.UUID      `json:"workflow_id" validate:"required"`
	TemplateID uuid.UUID      `json:"template_id" validate:"required"`
	PositionX  float64        `json:"position_x" validate:"required"`
	PositionY  float64        `json:"position_y" validate:"required"`
	Data       map[string]any `json:"data,omitempty"`
}

type UpdateWorkflowNodeRequest struct {
	ID        uuid.UUID       `json:"id" validate:"required"`
	PositionX *float64        `json:"position_x,omitempty"`
	PositionY *float64        `json:"position_y,omitempty"`
	Data      *map[string]any `json:"data,omitempty"`
}

type WorkflowNodeResponse struct {
	ID         uuid.UUID      `json:"id"`
	WorkflowID uuid.UUID      `json:"workflow_id"`
	TemplateID uuid.UUID      `json:"template_id"`
	PositionX  float64        `json:"position_x"`
	PositionY  float64        `json:"position_y"`
	Data       map[string]any `json:"data"`
}

func (wn *WorkflowNode) ToResponse() *WorkflowNodeResponse {
	return &WorkflowNodeResponse{
		ID:         wn.ID,
		WorkflowID: wn.WorkflowID,
		TemplateID: wn.TemplateID,
		PositionX:  wn.PositionX,
		PositionY:  wn.PositionY,
		Data:       wn.Data,
	}
}

type WorkflowNodeRepository interface {
	Create(ctx context.Context, workflowNode *CreateWorkflowNodeRequest) (*WorkflowNode, error)
	GetByID(ctx context.Context, id uuid.UUID) (*WorkflowNode, error)
	Update(ctx context.Context, workflowNode *UpdateWorkflowNodeRequest) error
	Delete(ctx context.Context, id uuid.UUID) error
	GetByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*WorkflowNode, error)
}

type WorkflowNodeService interface {
	CreateWorkflowNode(ctx context.Context, req *CreateWorkflowNodeRequest) (*WorkflowNodeResponse, error)
	GetWorkflowNode(ctx context.Context, id uuid.UUID) (*WorkflowNodeResponse, error)
	UpdateWorkflowNode(ctx context.Context, id uuid.UUID, req *UpdateWorkflowNodeRequest) error
	DeleteWorkflowNode(ctx context.Context, id uuid.UUID) error
	GetWorkflowNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*WorkflowNodeResponse, error)
}
