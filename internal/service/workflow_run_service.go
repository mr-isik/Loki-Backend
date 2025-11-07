package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type workflowRunService struct {
	repo domain.WorkflowRunRepository
}

func NewWorkflowRunService(repo domain.WorkflowRunRepository) domain.WorkflowRunService {
	return &workflowRunService{
		repo: repo,
	}
}

func (s *workflowRunService) StartWorkflowRun(ctx context.Context, workflowID uuid.UUID) (*domain.WorkflowRunResponse, error) {
	run, err := s.repo.Create(ctx, workflowID)
	if err != nil {
		return nil, err
	}

	return run.ToResponse(), nil
}

func (s *workflowRunService) GetWorkflowRun(ctx context.Context, id uuid.UUID) (*domain.WorkflowRunResponse, error) {
	run, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return run.ToResponse(), nil
}

func (s *workflowRunService) ListWorkflowRuns(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]*domain.WorkflowRunResponse, int, error) {
	// Set default limit if not provided
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	runs, total, err := s.repo.ListByWorkflowID(ctx, workflowID, limit, offset)
	if err != nil {
		return nil, 0, err
	}

	responses := make([]*domain.WorkflowRunResponse, len(runs))
	for i, run := range runs {
		responses[i] = run.ToResponse()
	}

	return responses, total, nil
}

func (s *workflowRunService) UpdateRunStatus(ctx context.Context, id uuid.UUID, status domain.WorkflowRunStatus) error {
	var finishedAt *time.Time
	
	// Set finished_at when status is terminal
	if status == domain.WorkflowRunStatusCompleted || 
	   status == domain.WorkflowRunStatusFailed || 
	   status == domain.WorkflowRunStatusCancelled {
		now := time.Now()
		finishedAt = &now
	}

	return s.repo.UpdateStatus(ctx, id, status, finishedAt)
}
