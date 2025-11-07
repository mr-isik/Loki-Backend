package service

import (
	"context"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type nodeTemplateService struct {
	repo domain.NodeTemplateRepository
}

func NewNodeTemplateService(repo domain.NodeTemplateRepository) domain.NodeTemplateService {
	return &nodeTemplateService{
		repo: repo,
	}
}

func (s *nodeTemplateService) ListNodeTemplates(ctx context.Context) ([]*domain.NodeTemplateResponse, error) {
	templates, err := s.repo.GetAll(ctx)
	if err != nil {
		return nil, err
	}

	responses := make([]*domain.NodeTemplateResponse, len(templates))
	for i, template := range templates {
		responses[i] = template.ToResponse()
	}

	return responses, nil
}

func (s *nodeTemplateService) GetNodeTemplate(ctx context.Context, id uuid.UUID) (*domain.NodeTemplateResponse, error) {
	template, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	return template.ToResponse(), nil
}
