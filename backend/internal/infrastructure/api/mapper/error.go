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

// mysqlDuplicateEntry is the MySQL error number (ER_DUP_ENTRY) for a unique
// constraint / duplicate key violation.
const mysqlDuplicateEntry = 1062

// MapWriteError turns a unique-constraint violation into a 409 Conflict and
// falls back to a 400 Bad Request for anything else. Recognizes both the
// Postgres and MySQL driver error types since the app supports either as its
// configured database.engine.
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

// MapOwnershipError turns the domain package's ownership-related sentinel
// errors into the right HTTP status, falling back to a 400 Bad Request.
func MapOwnershipError(fallbackMessage string, err error) error {
	switch {
	case errors.Is(err, domain.ErrForbidden):
		return huma.Error403Forbidden("you do not have access to this resource", err)
	case errors.Is(err, domain.ErrNotFound):
		return huma.Error404NotFound("resource not found", err)
	default:
		return huma.Error400BadRequest(fallbackMessage, err)
	}
}
