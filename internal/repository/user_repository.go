package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/mr-isik/loki-backend/internal/domain"
)

type userRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *pgxpool.Pool) domain.UserRepository {
	return &userRepository{db: db}
}

// Create creates a new user
func (r *userRepository) Create(ctx context.Context, user *domain.User) error {
	query := `
		INSERT INTO users (id, email, name, password, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	user.ID = uuid.New()
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	_, err := r.db.Exec(ctx, query,
		user.ID,
		user.Email,
		user.Name,
		user.Password,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		return domain.ParseDBError(err)
	}

	return nil
}

// GetByID retrieves a user by ID
func (r *userRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, name, password, created_at, updated_at, deleted_at
		FROM users
		WHERE id = $1 AND deleted_at IS NULL
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &user, nil
}

// GetByEmail retrieves a user by email
func (r *userRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, name, password, created_at, updated_at, deleted_at
		FROM users
		WHERE email = $1 AND deleted_at IS NULL
	`

	var user domain.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID,
		&user.Email,
		&user.Name,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
		&user.DeletedAt,
	)

	if err != nil {
		return nil, domain.ParseDBError(err)
	}

	return &user, nil
}

// GetAll retrieves all users with pagination
func (r *userRepository) GetAll(ctx context.Context, limit, offset int) ([]*domain.User, error) {
	query := `
		SELECT id, email, name, password, created_at, updated_at, deleted_at
		FROM users
		WHERE deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to get users: %w", err)
	}
	defer rows.Close()

	var users []*domain.User
	for rows.Next() {
		var user domain.User
		err := rows.Scan(
			&user.ID,
			&user.Email,
			&user.Name,
			&user.Password,
			&user.CreatedAt,
			&user.UpdatedAt,
			&user.DeletedAt,
		)
		if err != nil {
			return nil, domain.ParseDBError(err)
		}
		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, domain.ParseDBError(err)
	}

	return users, nil
}

// Update updates a user
func (r *userRepository) Update(ctx context.Context, user *domain.User) error {
	query := `
		UPDATE users
		SET email = $1, name = $2, updated_at = $3
		WHERE id = $4 AND deleted_at IS NULL
	`

	user.UpdatedAt = time.Now()

	result, err := r.db.Exec(ctx, query,
		user.Email,
		user.Name,
		user.UpdatedAt,
		user.ID,
	)

	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Delete soft deletes a user
func (r *userRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `
		UPDATE users
		SET deleted_at = $1
		WHERE id = $2 AND deleted_at IS NULL
	`

	result, err := r.db.Exec(ctx, query, time.Now(), id)
	if err != nil {
		return domain.ParseDBError(err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}

	return nil
}

// Count returns the total number of active users
func (r *userRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM users WHERE deleted_at IS NULL`

	var count int64
	err := r.db.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, domain.ParseDBError(err)
	}

	return count, nil
}
