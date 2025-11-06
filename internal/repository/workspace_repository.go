package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)



type workspaceRepository struct {
	db *pgxpool.Pool
}

// NewWorkspaceRepository creates a new workspace repository
func NewWorkspaceRepository(db *pgxpool.Pool) domain.WorkspaceRepository {
	return &workspaceRepository{db: db}
}

// Create creates a new workspace
func (r *workspaceRepository) Create(ctx context.Context, workspace *domain.Workspace) error {
	query := `
		INSERT INTO workspaces (id, owner_user_id, name, created_at)
		VALUES ($1, $2, $3, $4)
	`

	workspace.ID = uuid.New()
	workspace.CreatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		workspace.ID,
		workspace.OwnerUserID,
		workspace.Name,
		workspace.CreatedAt,
	)

	if err != nil {
		return domain.ParseDBError(err)
	}

	return nil
}

// GetByID retrieves a workspace by ID
func (r *workspaceRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Workspace, error) {
	query := `
		SELECT id, owner_user_id, name, created_at
		FROM workspaces
		WHERE id = $1
	`

	var workspace domain.Workspace
	err := r.db.QueryRow(ctx, query, id).Scan(
		&workspace.ID,
		&workspace.OwnerUserID,
		&workspace.Name,
		&workspace.CreatedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &workspace, nil
}

// GetByOwnerID retrieves all workspaces owned by a user
func (r *workspaceRepository) GetByOwnerID(ctx context.Context, ownerID uuid.UUID) ([]*domain.Workspace, error) {
	query := `
		SELECT id, owner_user_id, name, created_at
		FROM workspaces
		WHERE owner_user_id = $1
		ORDER BY created_at DESC
	`

	rows, err := r.db.Query(ctx, query, ownerID)
	if err != nil {
		return nil, fmt.Errorf("failed to get workspaces: %w", err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.OwnerUserID,
			&workspace.Name,
			&workspace.CreatedAt,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return workspaces, nil
}

// GetAll retrieves all workspaces with pagination
func (r *workspaceRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.Workspace, error) {
	query := `
		SELECT id, owner_user_id, name, created_at
		FROM workspaces
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, domain.ParseDBError(err)
	}
	defer rows.Close()

	var workspaces []*domain.Workspace
	for rows.Next() {
		var workspace domain.Workspace
		err := rows.Scan(
			&workspace.ID,
			&workspace.OwnerUserID,
			&workspace.Name,
			&workspace.CreatedAt,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		workspaces = append(workspaces, &workspace)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return workspaces, nil
}

// Update updates a workspace
func (r *workspaceRepository) Update(ctx context.Context, workspace *domain.Workspace) error {
	query := `
		UPDATE workspaces
		SET name = $1
		WHERE id = $2
	`

	result, err := r.db.Exec(ctx, query,
		workspace.Name,
		workspace.ID,
	)

	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete deletes a workspace
func (r *workspaceRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM workspaces WHERE id = $1`

	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Count returns the total number of workspaces
func (r *workspaceRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM workspaces`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, domain.ParseDBError(err)
	}

	return count, nil
}

// IsOwner checks if a user is the owner of a workspace
func (r *workspaceRepository) IsOwner(ctx context.Context, workspaceID, userID uuid.UUID) (bool, error) {
	query := `
		SELECT EXISTS(
			SELECT 1 FROM workspaces 
			WHERE id = $1 AND owner_user_id = $2
		)
	`

	var isOwner bool
	err := r.db.QueryRow(ctx, query, workspaceID, userID).Scan(&isOwner)
	if err != nil {
		return false, domain.ParseDBError(err)
	}

	return isOwner, nil
}
