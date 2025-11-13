package service

import (
	"context"

	"github.com/mr-isik/loki-backend/internal/domain"

	"github.com/google/uuid"
)

type WorkflowEdgeService struct {
	repo domain.WorkflowEdgeRepository
}

func NewWorkflowEdgeService(repo domain.WorkflowEdgeRepository) *WorkflowEdgeService {
	return &WorkflowEdgeService{repo: repo}
}

func (s *WorkflowEdgeService) CreateWorkflowEdge(ctx context.Context, req *domain.CreateWorkflowEdgeRequest) (*domain.WorkflowEdge, error) {
	return s.repo.Create(ctx, req)
}

func (s *WorkflowEdgeService) UpdateWorkflowEdge(ctx context.Context, id uuid.UUID, req *domain.UpdateWorkflowEdgeRequest) (*domain.WorkflowEdge, error) {
	return s.repo.Update(ctx, id, req)
}

func (s *WorkflowEdgeService) GetWorkflowEdge(ctx context.Context, id uuid.UUID) (*domain.WorkflowEdge, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *WorkflowEdgeService) GetWorkflowEdgesByWorkflow(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowEdge, error) {
	return s.repo.GetByWorkflowID(ctx, workflowID)
}

func (s *WorkflowEdgeService) DeleteWorkflowEdge(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
