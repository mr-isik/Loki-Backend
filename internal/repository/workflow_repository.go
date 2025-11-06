package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type workflowRepository struct {
	db *pgxpool.Pool
}

// NewWorkflowRepository creates a new workflow repository
func NewWorkflowRepository(db *pgxpool.Pool) domain.WorkflowRepository {
	return &workflowRepository{db: db}
}

// Create creates a new workflow
func (r *workflowRepository) Create(ctx context.Context, workflow *domain.Workflow) error {
	query := `
		INSERT INTO workflows (id, workspace_id, title, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	workflow.ID = uuid.New()
	workflow.CreatedAt = time.Now()
	workflow.UpdatedAt = time.Now()

	if workflow.Title == "" {
		workflow.Title = "Untitled Workflow"
	}

	if workflow.Status == "" {
		workflow.Status = domain.WorkflowStatusDraft
	}

	_, err := r.db.Exec(ctx, query,
		workflow.ID,
		workflow.WorkspaceID,
		workflow.Title,
		workflow.Status,
		workflow.CreatedAt,
		workflow.UpdatedAt,
	)

	if err != nil {
		return domain.ParseDBError(err)
	}

	return nil
}

// GetByID retrieves a workflow by ID
func (r *workflowRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workflow, error) {
	query := `
		SELECT id, workspace_id, title, status, created_at, updated_at
		FROM workflows
		WHERE id = $1
	`

	var workflow domain.Workflow
	err := r.db.QueryRow(ctx, query, id).Scan(
		&workflow.ID,
		&workflow.WorkspaceID,
		&workflow.Title,
		&workflow.Status,
		&workflow.CreatedAt,
		&workflow.UpdatedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &workflow, nil
}

// GetByWorkspaceID retrieves workflows by workspace ID with pagination
func (r *workflowRepository) GetByWorkspaceID(ctx context.Context, workspaceID uuid.UUID, limit, offset int) ([]*domain.Workflow, error) {
	query := `
		SELECT id, workspace_id, title, status, created_at, updated_at
		FROM workflows
		WHERE workspace_id = $1
		ORDER BY updated_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, workspaceID, limit, offset)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		var workflow domain.Workflow
		err := rows.Scan(
			&workflow.ID,
			&workflow.WorkspaceID,
			&workflow.Title,
			&workflow.Status,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		workflows = append(workflows, &workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return workflows, nil
}

// GetAll retrieves all workflows with pagination
func (r *workflowRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Workflow, error) {
	query := `
		SELECT id, workspace_id, title, status, created_at, updated_at
		FROM workflows
		ORDER BY updated_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()

	var workflows []*domain.Workflow
	for rows.Next() {
		var workflow domain.Workflow
		err := rows.Scan(
			&workflow.ID,
			&workflow.WorkspaceID,
			&workflow.Title,
			&workflow.Status,
			&workflow.CreatedAt,
			&workflow.UpdatedAt,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		workflows = append(workflows, &workflow)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return workflows, nil
}

// Update updates a workflow
func (r *workflowRepository) Update(ctx context.Context, workflow *domain.Workflow) error {
	query := `
		UPDATE workflows
		SET title = $1, status = $2, updated_at = $3
		WHERE id = $4
	`

	workflow.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		workflow.Title,
		workflow.Status,
		workflow.UpdatedAt,
		workflow.ID,
	)

	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a workflow
func (r *workflowRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workflows WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// CountByWorkspace returns the total number of workflows in a workspace
func (r *workflowRepository) CountByWorkspace(ctx context.Context, workspaceID uuid.UUID) (int64, error) {
	query := `SELECT COUNT(*) FROM workflows WHERE workspace_id = $1`

	var count int64
	err := r.db.QueryRow(ctx, query, workspaceID).Scan(&count)
	if err != nil {
		return 0, domain.ParseDBError(err)
	}

	return count, nil
}

// UpdateStatus updates only the status of a workflow
func (r *workflowRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.WorkflowStatus) error {
	query := `
		UPDATE workflows
		SET status = $1, updated_at = $2
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, status, time.Now(), id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}
