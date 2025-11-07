package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type workflowNodeService struct {
	repo domain.WorkflowNodeRepository
}

func NewWorkflowNodeService(repo domain.WorkflowNodeRepository) domain.WorkflowNodeService {
	return &workflowNodeService{
		repo: repo,
	}
}

func (s *workflowNodeService) CreateWorkflowNode(ctx context.Context, req *domain.CreateWorkflowNodeRequest) error {
	workflowNode := &domain.CreateWorkflowNodeRequest{
		WorkflowID: req.WorkflowID,
		TemplateID: req.TemplateID,
		PositionX: req.PositionX,
		PositionY: req.PositionY,
		Data:     req.Data,
	}

	if err := s.repo.Create(ctx, workflowNode); err != nil {
		return fmt.Errorf("failed to create workflow node: %w", err)
	}

	return nil
}

func (s *workflowNodeService) GetWorkflowNode(ctx context.Context, id uuid.UUID) (*domain.WorkflowNodeResponse, error) {
	workflowNode, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow node: %w", err)
	}
	return workflowNode.ToResponse(), nil
}

func (s *workflowNodeService) UpdateWorkflowNode(ctx context.Context, id uuid.UUID, req *domain.UpdateWorkflowNodeRequest) error {
	workflowNode, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get workflow node: %w", err)
	}
	if req.PositionX != nil {
		workflowNode.PositionX = *req.PositionX
	}
	if req.PositionY != nil {
		workflowNode.PositionY = *req.PositionY
	}
	if req.Data != nil {
		workflowNode.Data = *req.Data
	}

	workflowNodeToUpdate := &domain.UpdateWorkflowNodeRequest{
		ID:         workflowNode.ID,
		PositionX: &workflowNode.PositionX,
		PositionY: &workflowNode.PositionY,
		Data:      &workflowNode.Data,
	}
	
	if err := s.repo.Update(ctx, workflowNodeToUpdate); err != nil {
		return fmt.Errorf("failed to update workflow node: %w", err)
	}
	return nil
}

func (s *workflowNodeService) DeleteWorkflowNode(ctx context.Context, id uuid.UUID) error {
	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow node: %w", err)
	}
	return nil
}

func (s *workflowNodeService) GetWorkflowNodesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowNodeResponse, error) {
	workflowNodes, err := s.repo.GetByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workflow nodes: %w", err)
	}
	var responses []*domain.WorkflowNodeResponse
	for _, wn := range workflowNodes {
		responses = append(responses, wn.ToResponse())
	}
	return responses, nil
}
