package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type nodeRunLogService struct {
	repo domain.NodeRunLogRepository
}

func NewNodeRunLogService(repo domain.NodeRunLogRepository) domain.NodeRunLogService {
	return &nodeRunLogService{
		repo: repo,
	}
}

func (s *nodeRunLogService) CreateNodeRunLog(ctx context.Context, req *domain.CreateNodeRunLogRequest) (*domain.NodeRunLogResponse, error) {
	log, err := s.repo.Create(ctx, req)
	if err != nil {
		return nil, err
	}

	return log.ToResponse(), nil
}

func (s *nodeRunLogService) GetNodeRunLog(ctx context.Context, id uuid.UUID) (*domain.NodeRunLogResponse, error) {
	log, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return log.ToResponse(), nil
}

func (s *nodeRunLogService) GetNodeRunLogsByRunID(ctx context.Context, runID uuid.UUID) ([]*domain.NodeRunLogResponse, error) {
	logs, err := s.repo.GetByRunID(ctx, runID)
	if err != nil {
		return nil, err
	}

	responses := make([]*domain.NodeRunLogResponse, len(logs))
	for i, log := range logs {
		responses[i] = log.ToResponse()
	}

	return responses, nil
}

func (s *nodeRunLogService) UpdateNodeRunLog(ctx context.Context, id uuid.UUID, req *domain.UpdateNodeRunLogRequest) error {
	return s.repo.Update(ctx, id, req)
}
