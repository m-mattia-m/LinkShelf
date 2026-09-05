package controller

import (
	"errors"

	"github.com/danielgtaylor/huma/v2"
	"github.com/jackc/pgx/v5/pgconn"
)

// pgUniqueViolation is the Postgres SQLSTATE code for a unique constraint violation.
const pgUniqueViolation = "23505"

// mapWriteError turns a unique-constraint violation into a 409 Conflict and
// falls back to a 400 Bad Request for anything else.
func mapWriteError(fallbackMessage string, err error) error {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgUniqueViolation {
		return huma.Error409Conflict("a resource with this value already exists", err)
	}

	return huma.Error400BadRequest(fallbackMessage, err)
}
