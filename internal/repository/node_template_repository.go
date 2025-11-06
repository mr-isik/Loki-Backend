package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type nodeTemplateRepository struct {
	db *pgxpool.Pool
}

func NewNodeTemplateRepository(db *pgxpool.Pool) domain.NodeTemplateRepository {
	return &nodeTemplateRepository{
		db: db,
	}
}

func (r *nodeTemplateRepository) GetAll(ctx context.Context) ([]*domain.NodeTemplate, error) {
	query := `
		SELECT id, name, description, type_key, category
		FROM node_templates
		ORDER BY category, name
	`
	
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()

	var templates []*domain.NodeTemplate
	for rows.Next() {
		var template domain.NodeTemplate
		if err := rows.Scan(
			&template.ID,
			&template.Name,
			&template.Description,
			&template.TypeKey,
			&template.Category,
		); err != nil {
			return nil, domain.ParseDBError(err)
		}
		templates = append(templates, &template)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return templates, nil
}

func (r *nodeTemplateRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.NodeTemplate, error) {
	query := `
		SELECT id, name, description, type_key, category
		FROM node_templates
		WHERE id = $1
	`

	var template domain.NodeTemplate
	err := r.db.QueryRow(ctx, query, id).Scan(
		&template.ID,
		&template.Name,
		&template.Description,
		&template.TypeKey,
		&template.Category,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &template, nil
}

