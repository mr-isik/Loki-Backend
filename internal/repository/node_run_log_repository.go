package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type NodeRunLogRepository struct {
	db *pgxpool.Pool
}

func NewNodeRunLogRepository(db *pgxpool.Pool) domain.NodeRunLogRepository {
	return &NodeRunLogRepository{db: db}
}

func (r *NodeRunLogRepository) Create(ctx context.Context, req *domain.CreateNodeRunLogRequest) (*domain.NodeRunLog, error) {
	query := `
		INSERT INTO node_run_logs (id, run_id, node_id, status, started_at, created_at, updated_at)
		VALUES (gen_random_uuid(), $1, $2, $3, NOW(), NOW(), NOW())
		RETURNING id, run_id, node_id, status, log_output, error_msg, started_at, finished_at, created_at, updated_at
	`

	var log domain.NodeRunLog
	err := r.db.QueryRow(ctx, query, req.RunID, req.NodeID, req.Status).Scan(
		&log.ID,
		&log.RunID,
		&log.NodeID,
		&log.Status,
		&log.LogOutput,
		&log.ErrorMsg,
		&log.StartedAt,
		&log.FinishedAt,
		&log.CreatedAt,
		&log.UpdatedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &log, nil
}

func (r *NodeRunLogRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NodeRunLog, error) {
	query := `
		SELECT id, run_id, node_id, status, log_output, error_msg, started_at, finished_at, created_at, updated_at
		FROM node_run_logs
		WHERE id = $1
	`

	var log domain.NodeRunLog
	err := r.db.QueryRow(ctx, query, id).Scan(
		&log.ID,
		&log.RunID,
		&log.NodeID,
		&log.Status,
		&log.LogOutput,
		&log.ErrorMsg,
		&log.StartedAt,
		&log.FinishedAt,
		&log.CreatedAt,
		&log.UpdatedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &log, nil
}

func (r *NodeRunLogRepository) GetByRunID(ctx context.Context, runID uuid.UUID) ([]*domain.NodeRunLog, error) {
	query := `
		SELECT id, run_id, node_id, status, log_output, error_msg, started_at, finished_at, created_at, updated_at
		FROM node_run_logs
		WHERE run_id = $1
		ORDER BY started_at ASC
	`

	rows, err := r.db.Query(ctx, query, runID)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()

	var logs []*domain.NodeRunLog
	for rows.Next() {
		var log domain.NodeRunLog
		if err := rows.Scan(
			&log.ID,
			&log.RunID,
			&log.NodeID,
			&log.Status,
			&log.LogOutput,
			&log.ErrorMsg,
			&log.StartedAt,
			&log.FinishedAt,
			&log.CreatedAt,
			&log.UpdatedAt,
		); err != nil {
			return nil, domain.ParseDBError(err)
		}
		logs = append(logs, &log)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return logs, nil
}

func (r *NodeRunLogRepository) Update(ctx context.Context, id uuid.UUID, req *domain.UpdateNodeRunLogRequest) error {
	// Build dynamic update query
	query := `
		UPDATE node_run_logs
		SET 
			status = COALESCE(NULLIF($1::text, ''), status::text)::varchar(50),
			log_output = COALESCE(NULLIF($2, ''), log_output),
			error_msg = COALESCE(NULLIF($3, ''), error_msg),
			finished_at = CASE 
				WHEN $1 IN ('completed', 'failed', 'skipped') AND finished_at IS NULL THEN NOW()
				ELSE finished_at
			END,
			updated_at = NOW()
		WHERE id = $4
	`

	result, err := r.db.Exec(ctx, query, req.Status, req.LogOutput, req.ErrorMsg, id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNodeRunLogNotFound
	}

	return nil
}
