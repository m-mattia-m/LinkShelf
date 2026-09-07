package mapper_test

import (
	"errors"
	"testing"

	"backend/internal/domain"
	"backend/internal/infrastructure/api/mapper"

	"github.com/danielgtaylor/huma/v2"
	"github.com/go-sql-driver/mysql"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/stretchr/testify/require"
)

func statusOf(t *testing.T, err error) int {
	t.Helper()
	var statusErr huma.StatusError
	require.ErrorAs(t, err, &statusErr)
	return statusErr.GetStatus()
}

func Test_MapWriteError_PostgresUniqueViolation_Maps409(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23505", Message: "duplicate key value"}

	err := mapper.MapWriteError("failed to write", pgErr)

	require.Error(t, err)
	require.Equal(t, 409, statusOf(t, err))
}

func Test_MapWriteError_MySQLDuplicateEntry_Maps409(t *testing.T) {
	myErr := &mysql.MySQLError{Number: 1062, Message: "Duplicate entry"}

	err := mapper.MapWriteError("failed to write", myErr)

	require.Error(t, err)
	require.Equal(t, 409, statusOf(t, err))
}

func Test_MapWriteError_OtherPostgresError_FallsBack400(t *testing.T) {
	pgErr := &pgconn.PgError{Code: "23503", Message: "foreign key violation"}

	err := mapper.MapWriteError("failed to write", pgErr)

	require.Error(t, err)
	require.Equal(t, 400, statusOf(t, err))
}

func Test_MapWriteError_OtherMySQLError_FallsBack400(t *testing.T) {
	myErr := &mysql.MySQLError{Number: 1451, Message: "cannot delete or update a parent row"}

	err := mapper.MapWriteError("failed to write", myErr)

	require.Error(t, err)
	require.Equal(t, 400, statusOf(t, err))
}

func Test_MapWriteError_UnrelatedError_FallsBack400(t *testing.T) {
	err := mapper.MapWriteError("failed to write", errors.New("something else"))

	require.Error(t, err)
	require.Equal(t, 400, statusOf(t, err))
}

func Test_MapOwnershipError_Forbidden_Maps403(t *testing.T) {
	err := mapper.MapOwnershipError("failed", domain.ErrForbidden)

	require.Error(t, err)
	require.Equal(t, 403, statusOf(t, err))
}

func Test_MapOwnershipError_NotFound_Maps404(t *testing.T) {
	err := mapper.MapOwnershipError("failed", domain.ErrNotFound)

	require.Error(t, err)
	require.Equal(t, 404, statusOf(t, err))
}

func Test_MapOwnershipError_OtherError_FallsBack400(t *testing.T) {
	err := mapper.MapOwnershipError("failed", errors.New("something else"))

	require.Error(t, err)
	require.Equal(t, 400, statusOf(t, err))
}
