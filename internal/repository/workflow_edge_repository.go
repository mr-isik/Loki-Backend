package repository

import (
	"context"

	"github.com/mr-isik/loki-backend/internal/domain"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type WorkflowEdgeRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowEdgeRepository(db *pgxpool.Pool) *WorkflowEdgeRepository {
	return &WorkflowEdgeRepository{db: db}
}

func (r *WorkflowEdgeRepository) Create(ctx context.Context, edge *domain.CreateWorkflowEdgeRequest) (*domain.WorkflowEdge, error) {
	query := `
		INSERT INTO workflow_edges (
			id, workflow_id, source_node_id, target_node_id, source_handle, target_handle
		)
		SELECT 
			$1, $2, $3, $4, $5, $6
		FROM workflows wf
		WHERE wf.id = $2
	`

	id := uuid.New()

	_, err := r.db.Exec(ctx, query,
		id,
		edge.WorkflowID,
		edge.SourceNodeID,
		edge.TargetNodeID,
		edge.SourceHandle,
		edge.TargetHandle,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &domain.WorkflowEdge{
		ID:           id,
		WorkflowID:   edge.WorkflowID,
		SourceNodeID: edge.SourceNodeID,
		TargetNodeID: edge.TargetNodeID,
		SourceHandle: edge.SourceHandle,
		TargetHandle: edge.TargetHandle,
	}, nil
}

func (r *WorkflowEdgeRepository) Update(ctx context.Context, id uuid.UUID, edge *domain.UpdateWorkflowEdgeRequest) (*domain.WorkflowEdge, error) {
	var sourceNodeID *uuid.UUID
	if edge.SourceNodeID != uuid.Nil {
		sourceNodeID = &edge.SourceNodeID
	}

	var targetNodeID *uuid.UUID
	if edge.TargetNodeID != uuid.Nil {
		targetNodeID = &edge.TargetNodeID
	}

	query := `
		UPDATE workflow_edges
		SET 
			source_node_id = COALESCE($1, source_node_id),
			target_node_id = COALESCE($2, target_node_id),
			source_handle = COALESCE(NULLIF($3, ''), source_handle),
			target_handle = COALESCE(NULLIF($4, ''), target_handle),
			updated_at = NOW()
		WHERE id = $5
	`

	result, err := r.db.Exec(ctx, query,
		sourceNodeID,
		targetNodeID,
		edge.SourceHandle,
		edge.TargetHandle,
		id,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return nil, domain.ErrWorkflowEdgeNotFound
	}

	return &domain.WorkflowEdge{
		ID:           id,
		SourceNodeID: edge.SourceNodeID,
		TargetNodeID: edge.TargetNodeID,
		SourceHandle: edge.SourceHandle,
		TargetHandle: edge.TargetHandle,
	}, nil
}

func (r *WorkflowEdgeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowEdge, error) {
	query := `
		SELECT id, workflow_id, source_node_id, target_node_id, source_handle, target_handle
		FROM workflow_edges
		WHERE id = $1
	`

	var edge domain.WorkflowEdge
	err := r.db.QueryRow(ctx, query, id).Scan(
		&edge.ID,
		&edge.WorkflowID,
		&edge.SourceNodeID,
		&edge.TargetNodeID,
		&edge.SourceHandle,
		&edge.TargetHandle,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &edge, nil
}

func (r *WorkflowEdgeRepository) GetByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowEdge, error) {
	query := `
		SELECT id, workflow_id, source_node_id, target_node_id, source_handle, target_handle
		FROM workflow_edges
		WHERE workflow_id = $1
	`

	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()

	var edges []*domain.WorkflowEdge
	for rows.Next() {
		var edge domain.WorkflowEdge
		err := rows.Scan(
			&edge.ID,
			&edge.WorkflowID,
			&edge.SourceNodeID,
			&edge.TargetNodeID,
			&edge.SourceHandle,
			&edge.TargetHandle,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		edges = append(edges, &edge)
	}

	if err = rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return edges, nil
}

func (r *WorkflowEdgeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workflow_edges WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrWorkflowEdgeNotFound
	}

	return nil
}
