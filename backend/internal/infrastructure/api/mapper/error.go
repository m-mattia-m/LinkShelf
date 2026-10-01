package mapper

import (
	"backend/internal/domain"
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
)

// pgUniqueViolation is the Postgres SQLSTATE code for a unique constraint violation.
const pgUniqueViolation = "23505"

// mysqlDuplicateEntry is ER_DUP_ENTRY.
const mysqlDuplicateEntry = 1062

// MapWriteError maps unique violations (Postgres or MySQL) to 409, everything else to 400.
func MapWriteError(fallbackMessage string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return huma.Error409Conflict("a resource with this value already exists", err)
	}

	var myErr *mysql.MySQLError
	if errors.As(err, &myErr) && myErr.Number == mysqlDuplicateEntry {
		return huma.Error409Conflict("a resource with this value already exists", err)
	}

	return huma.Error400BadRequest(fallbackMessage, err)
}

// MapOwnershipError maps ownership errors to HTTP statuses, otherwise 400.
func MapOwnershipError(fallbackMessage string, err error) error {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return huma.Error403Forbidden("you do not have access to this resource", err)
	case errors.Is(err, domain.ErrNotFound):
		return huma.Error404NotFound("resource not found", err)
	case errors.Is(err, domain.ErrConflict):
		return huma.Error409Conflict("a resource with this value already exists", err)
	default:
		return huma.Error400BadRequest(fallbackMessage, err)
	}
}
