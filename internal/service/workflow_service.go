package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type workflowService struct {
	workflowRepo  domain.WorkflowRepository
	workspaceRepo domain.WorkspaceRepository
}

// NewWorkflowService creates a new workflow service
func NewWorkflowService(workflowRepo domain.WorkflowRepository, workspaceRepo domain.WorkspaceRepository) domain.WorkflowService {
	return &workflowService{
		workflowRepo:  workflowRepo,
		workspaceRepo: workspaceRepo,
	}
}

// CreateWorkflow creates a new workflow
func (s *workflowService) CreateWorkflow(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, req *domain.CreateWorkflowRequest) (*domain.WorkflowResponse, error) {
	// Check if user is the owner of the workspace
	isOwner, err := s.workspaceRepo.IsOwner(ctx, workspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrUnauthorized
	}

	workflow := &domain.Workflow{
		WorkspaceID: workspaceID,
		Title:       req.Title,
		Status:      domain.WorkflowStatusDraft,
	}

	if workflow.Title == "" {
		workflow.Title = "Untitled Workflow"
	}

	if err := s.workflowRepo.Create(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to create workflow: %w", err)
	}

	return workflow.ToResponse(), nil
}

func (s *workflowService) GetWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) (*domain.WorkflowResponse, error) {
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return nil, domain.ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	isOwner, err := s.workspaceRepo.IsOwner(ctx, workflow.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrUnauthorized
	}

	return workflow.ToResponse(), nil
}

func (s *workflowService) GetWorkspaceWorkflows(ctx context.Context, workspaceID uuid.UUID, userID uuid.UUID, page, pageSize int) ([]*domain.WorkflowResponse, int64, error) {
	isOwner, err := s.workspaceRepo.IsOwner(ctx, workspaceID, userID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return nil, 0, domain.ErrUnauthorized
	}

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 10
	}

	offset := (page - 1) * pageSize

	workflows, err := s.workflowRepo.GetByWorkspaceID(ctx, workspaceID, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to get workflows: %w", err)
	}

	total, err := s.workflowRepo.CountByWorkspace(ctx, workspaceID)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count workflows: %w", err)
	}

	responses := make([]*domain.WorkflowResponse, len(workflows))
	for i, workflow := range workflows {
		responses[i] = workflow.ToResponse()
	}

	return responses, total, nil
}

func (s *workflowService) UpdateWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID, req *domain.UpdateWorkflowRequest) (*domain.WorkflowResponse, error) {
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return nil, domain.ErrWorkflowNotFound
		}
		return nil, fmt.Errorf("failed to get workflow: %w", err)
	}

	isOwner, err := s.workspaceRepo.IsOwner(ctx, workflow.WorkspaceID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return nil, domain.ErrUnauthorized
	}

	if req.Title != "" {
		workflow.Title = req.Title
	}

	if req.Status != "" {
		workflow.Status = req.Status
	}

	if err := s.workflowRepo.Update(ctx, workflow); err != nil {
		return nil, fmt.Errorf("failed to update workflow: %w", err)
	}

	return workflow.ToResponse(), nil
}

// DeleteWorkflow deletes a workflow
func (s *workflowService) DeleteWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get existing workflow
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return domain.ErrWorkflowNotFound
		}
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Check if user is the owner of the workspace
	isOwner, err := s.workspaceRepo.IsOwner(ctx, workflow.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return domain.ErrUnauthorized
	}

	if err := s.workflowRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete workflow: %w", err)
	}

	return nil
}

// PublishWorkflow publishes a workflow
func (s *workflowService) PublishWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get existing workflow
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return domain.ErrWorkflowNotFound
		}
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Check if user is the owner of the workspace
	isOwner, err := s.workspaceRepo.IsOwner(ctx, workflow.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return domain.ErrUnauthorized
	}

	// Update status
	if err := s.workflowRepo.UpdateStatus(ctx, id, domain.WorkflowStatusPublished); err != nil {
		return fmt.Errorf("failed to publish workflow: %w", err)
	}

	// Get updated workflow
	workflow, err = s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get updated workflow: %w", err)
	}

	return nil
}

// ArchiveWorkflow archives a workflow
func (s *workflowService) ArchiveWorkflow(ctx context.Context, id uuid.UUID, userID uuid.UUID) error {
	// Get existing workflow
	workflow, err := s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		if errors.Is(err, domain.ErrWorkflowNotFound) {
			return domain.ErrWorkflowNotFound
		}
		return fmt.Errorf("failed to get workflow: %w", err)
	}

	// Check if user is the owner of the workspace
	isOwner, err := s.workspaceRepo.IsOwner(ctx, workflow.WorkspaceID, userID)
	if err != nil {
		return fmt.Errorf("failed to check workspace ownership: %w", err)
	}
	if !isOwner {
		return domain.ErrUnauthorized
	}

	// Update status
	if err := s.workflowRepo.UpdateStatus(ctx, id, domain.WorkflowStatusArchived); err != nil {
		return fmt.Errorf("failed to archive workflow: %w", err)
	}

	// Get updated workflow
	workflow, err = s.workflowRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get updated workflow: %w", err)
	}

	return nil
}
