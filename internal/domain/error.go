package domain

import (
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// Common repository errors
var (
	ErrNotFound          = errors.New("resource not found")
	ErrAlreadyExists     = errors.New("resource already exists")
	ErrInvalidInput      = errors.New("invalid input")
	ErrForeignKeyViolation = errors.New("foreign key constraint violation")
	ErrUniqueViolation   = errors.New("unique constraint violation")
	ErrCheckViolation    = errors.New("check constraint violation")
	ErrDeadlock          = errors.New("database deadlock")
	ErrConnectionFailed  = errors.New("database connection failed")
)

// PostgreSQL error codes
const (
	PgErrCodeUniqueViolation      = "23505"
	PgErrCodeForeignKeyViolation  = "23503"
	PgErrCodeCheckViolation       = "23514"
	PgErrCodeNotNullViolation     = "23502"
	PgErrCodeDeadlock             = "40P01"
	PgErrCodeSerializationFailure = "40001"
)

// ParseDBError converts PostgreSQL errors to domain errors
func ParseDBError(err error) error {
	if err == nil {
		return nil
	}

	// Check for no rows error
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}

	// Check for pgconn.PgError
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case PgErrCodeUniqueViolation:
			return fmt.Errorf("%w: %s", ErrUniqueViolation, pgErr.ConstraintName)
		case PgErrCodeForeignKeyViolation:
			return fmt.Errorf("%w: %s", ErrForeignKeyViolation, pgErr.ConstraintName)
		case PgErrCodeCheckViolation:
			return fmt.Errorf("%w: %s", ErrCheckViolation, pgErr.ConstraintName)
		case PgErrCodeNotNullViolation:
			return fmt.Errorf("%w: column %s cannot be null", ErrInvalidInput, pgErr.ColumnName)
		case PgErrCodeDeadlock, PgErrCodeSerializationFailure:
			return ErrDeadlock
		}
		
		// Return generic error with PostgreSQL details for debugging
		return fmt.Errorf("database error [%s]: %s", pgErr.Code, pgErr.Message)
	}

	// Return original error if not a known type
	return err
}

// IsNotFoundError checks if error is a not found error
func IsNotFoundError(err error) bool {
	return errors.Is(err, ErrNotFound)
}

// IsAlreadyExistsError checks if error is an already exists error
func IsAlreadyExistsError(err error) bool {
	return errors.Is(err, ErrAlreadyExists) || errors.Is(err, ErrUniqueViolation)
}

// IsForeignKeyError checks if error is a foreign key violation
func IsForeignKeyError(err error) bool {
	return errors.Is(err, ErrForeignKeyViolation)
}
