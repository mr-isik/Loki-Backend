package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type WorkflowRunRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowRunRepository(db *pgxpool.Pool) domain.WorkflowRunRepository {
	return &WorkflowRunRepository{db: db}
}

func (r *WorkflowRunRepository) Create(ctx context.Context, workflowID uuid.UUID) (*domain.WorkflowRun, error) {
	query := `
		INSERT INTO workflow_runs (id, workflow_id, status, started_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, NOW(), NOW(), NOW())
		RETURNING id, workflow_id, status, started_at, finished_at, created_at, updated_at
	`

	var run domain.WorkflowRun
	err := r.db.QueryRow(ctx, query, workflowID, domain.WorkflowRunStatusRunning).Scan(
		&run.ID,
		&run.WorkflowID,
		&run.Status,
		&run.StartedAt,
		&run.FinishedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &run, nil
}

func (r *WorkflowRunRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowRun, error) {
	query := `
		SELECT id, workflow_id, status, started_at, finished_at, created_at, updated_at
		FROM workflow_runs
		WHERE id = $1
	`

	var run domain.WorkflowRun
	err := r.db.QueryRow(ctx, query, id).Scan(
		&run.ID,
		&run.WorkflowID,
		&run.Status,
		&run.StartedAt,
		&run.FinishedAt,
		&run.CreatedAt,
		&run.UpdatedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &run, nil
}

func (r *WorkflowRunRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status domain.WorkflowRunStatus, finishedAt *time.Time) error {
	query := `
		UPDATE workflow_runs
		SET status = $1, finished_at = $2, updated_at = NOW()
		WHERE id = $3
	`

	result, err := r.db.Exec(ctx, query, status, finishedAt, id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrWorkflowRunNotFound
	}

	return nil
}

func (r *WorkflowRunRepository) ListByWorkflowID(ctx context.Context, workflowID uuid.UUID, limit, offset int) ([]*domain.WorkflowRun, int, error) {
	// Get total count
	countQuery := `SELECT COUNT(*) FROM workflow_runs WHERE workflow_id = $1`
	var total int
	if err := r.db.QueryRow(ctx, countQuery, workflowID).Scan(&total); err != nil {
		return nil, 0, domain.ParseDBError(err)
	}

	// Get paginated results
	query := `
		SELECT id, workflow_id, status, started_at, finished_at, created_at, updated_at
		FROM workflow_runs
		WHERE workflow_id = $1
		ORDER BY started_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(ctx, query, workflowID, limit, offset)
	if err != nil {
		return nil, 0, domain.ParseDBError(err)
	}
	defer rows.Close()

	var runs []*domain.WorkflowRun
	for rows.Next() {
		var run domain.WorkflowRun
		if err := rows.Scan(
			&run.ID,
			&run.WorkflowID,
			&run.Status,
			&run.StartedAt,
			&run.FinishedAt,
			&run.CreatedAt,
			&run.UpdatedAt,
		); err != nil {
			return nil, 0, domain.ParseDBError(err)
		}
		runs = append(runs, &run)
	}

	if err := rows.Err(); err != nil {
		return nil, 0, domain.ParseDBError(err)
	}

	return runs, total, nil
}
