package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type workflowNodeRepository struct {
	db *pgxpool.Pool
}

func NewWorkflowNodeRepository(db *pgxpool.Pool) domain.WorkflowNodeRepository {
	return &workflowNodeRepository{db: db}
}

func (r *workflowNodeRepository) Create(ctx context.Context, workflowNode *domain.CreateWorkflowNodeRequest) (*domain.WorkflowNode, error) {
	query := `
		INSERT INTO workflow_nodes (id, workflow_id, template_id, position_x, position_y, data)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	id := uuid.New()
	_, err := r.db.Exec(ctx, query, id, workflowNode.WorkflowID, workflowNode.TemplateID, workflowNode.PositionX, workflowNode.PositionY, workflowNode.Data)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &domain.WorkflowNode{
		ID:         id,
		WorkflowID: workflowNode.WorkflowID,
		TemplateID: workflowNode.TemplateID,
		PositionX:  workflowNode.PositionX,
		PositionY:  workflowNode.PositionY,
		Data:       workflowNode.Data,
	}, nil
}

func (r *workflowNodeRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.WorkflowNode, error) {
	query := `
		SELECT id, workflow_id, template_id, position_x, position_y, data
		FROM workflow_nodes
		WHERE id = $1
	`
	var workflowNode domain.WorkflowNode
	err := r.db.QueryRow(ctx, query, id).Scan(
		&workflowNode.ID,
		&workflowNode.WorkflowID,
		&workflowNode.TemplateID,
		&workflowNode.PositionX,
		&workflowNode.PositionY,
		&workflowNode.Data,
	)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	return &workflowNode, nil
}

func (r *workflowNodeRepository) Update(ctx context.Context, workflowNode *domain.UpdateWorkflowNodeRequest) error {
	query := `
		UPDATE workflow_nodes
		SET position_x = $1, position_y = $2, data = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, workflowNode.PositionX, workflowNode.PositionY, workflowNode.Data, workflowNode.ID)
	return domain.ParseDBError(err)
}

func (r *workflowNodeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		DELETE FROM workflow_nodes
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, id)
	return domain.ParseDBError(err)
}

func (r *workflowNodeRepository) GetByWorkflowID(ctx context.Context, workflowID uuid.UUID) ([]*domain.WorkflowNode, error) {
	query := `
		SELECT id, workflow_id, template_id, position_x, position_y, data
		FROM workflow_nodes
		WHERE workflow_id = $1
	`
	rows, err := r.db.Query(ctx, query, workflowID)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()
	var workflowNodes []*domain.WorkflowNode
	for rows.Next() {
		var workflowNode domain.WorkflowNode
		err := rows.Scan(
			&workflowNode.ID,
			&workflowNode.WorkflowID,
			&workflowNode.TemplateID,
			&workflowNode.PositionX,
			&workflowNode.PositionY,
			&workflowNode.Data,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		workflowNodes = append(workflowNodes, &workflowNode)
	}
	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}
	return workflowNodes, nil
}
