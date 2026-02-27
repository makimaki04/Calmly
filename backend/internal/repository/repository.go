package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/makimaki04/Calmly/internal/models"
	"go.uber.org/zap"
)

type Dump interface {
	CreateDump(ctx context.Context, dump models.Dump) (uuid.UUID, error)
	UpdateStatus(ctx context.Context, dumpID uuid.UUID, status models.DumpStatus) error
	ClearRawText(ctx context.Context, dumpID uuid.UUID) error
	GetActiveDump(ctx context.Context, userID uuid.UUID) (*models.Dump, error)
}

type DumpAnalysis interface {
	SaveAnalysis(ctx context.Context, dumpAnalysis models.DumpAnalysis) error
	GetAnalysis(ctx context.Context, dumpID uuid.UUID) (*models.DumpAnalysis, error)
}

type DumpAnswers interface {
}

type Plan interface {
}

type PlanItem interface {
}

type Repository struct {
	Dump
	DumpAnalysis
	DumpAnswers
	Plan
	PlanItem
}

func NewRepository(db *sql.DB, logger *zap.Logger) *Repository {
	return &Repository{
		Dump:         NewDumpRepository(db, logger),
		DumpAnalysis: NewDumpAnalysisRepo(db, logger),
	}
}

var (
	// ErrForeignKeyViolation is returned when the referenced user does not exist.
	ErrForeignKeyViolation = errors.New("user missing")
	// ErrBadItemType is returned when the item type is invalid.
	ErrBadItemType = errors.New("bad item type")
	// ErrRetryableDB is returned for retryable database errors.
	ErrRetryableDB = errors.New("retryable db error")
	// ErrSchemaMismatch is returned when the database schema is incompatible with the query.
	ErrSchemaMismatch = errors.New("schema mismatch")
	// ErrNotFound is returned when an item can't be found.
	ErrNotFound = errors.New("item not found")
	// ErrDB is returned for non-specific database errors.
	ErrDB = errors.New("db error")
)

// checkErr classifies a database error into a domain sentinel.
// It does NOT log — the calling method owns the log.
func checkErr(err error) error {
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return err
	}

	if errors.Is(err, sql.ErrNoRows) {
		return ErrNotFound
	}

	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return ErrDB
	}

	switch pgErr.Code {
	case pgerrcode.ForeignKeyViolation:
		return ErrForeignKeyViolation
	case pgerrcode.NotNullViolation:
		return ErrDB
	case pgerrcode.InvalidTextRepresentation:
		return ErrBadItemType
	case pgerrcode.SerializationFailure, pgerrcode.DeadlockDetected, pgerrcode.LockNotAvailable:
		return ErrRetryableDB
	case pgerrcode.InvalidColumnReference:
		return ErrSchemaMismatch
	default:
		return ErrDB
	}
}
