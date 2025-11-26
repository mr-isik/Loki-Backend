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

func (s *WorkflowEdgeService) CreateWorkflowEdge(ctx context.Context, req *domain.CreateWorkflowEdgeRequest) (*domain.WorkflowEdgeResponse, error) {
	edge, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}
	return edge.ToResponse(), nil
}

func (s *WorkflowEdgeService) UpdateWorkflowEdge(ctx context.Context, id uuid.UUID, req *domain.UpdateWorkflowEdgeRequest) (*domain.WorkflowEdgeResponse, error) {
	edge, err := s.repo.Update(ctx, id, req)
	if err != nil {
		return nil, err
	}
	return edge.ToResponse(), nil
}

func (s *WorkflowEdgeService) GetWorkflowEdgeByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowEdgeResponse, error) {
	edge, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return edge.ToResponse(), nil
}

func (s *WorkflowEdgeService) GetWorkflowEdgesByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowEdgeResponse, error) {
	edges, err := s.repo.GetByWorkflowID(ctx, workflowID)
	if err != nil {
		return nil, err
	}
	var responses []*domain.WorkflowEdgeResponse
	for _, edge := range edges {
		responses = append(responses, edge.ToResponse())
	}
	return responses, nil
}

func (s *WorkflowEdgeService) DeleteWorkflowEdge(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
