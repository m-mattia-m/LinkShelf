package repository

import (
	"testing"

	"backend/internal/config"

	"github.com/stretchr/testify/require"
)

// setDatabaseTestConfig mutates the package-global config singleton, so it
// restores the normal test config afterward - otherwise a leftover
// database.engine=mysql would leak into every other test in this package
// that relies on the default (Postgres) test config, e.g. user_test.go's
// sqlmock expectations.
func setDatabaseTestConfig(t *testing.T, engine string) {
	t.Helper()
	config.Reset()
	require.NoError(t, config.LoadConfig())
	config.Set("database.engine", engine)
	config.Set("database.host", "db-host")
	config.Set("database.port", "5432")
	config.Set("database.name", "db-name")
	config.Set("database.username", "db-user")
	config.Set("database.password", "db-pass")

	t.Cleanup(func() {
		config.Reset()
		_ = config.LoadConfig()
	})
}

func Test_Unit_GetConnectionInformation_Postgres(t *testing.T) {
	setDatabaseTestConfig(t, "postgres")
	config.Set("database.params", "sslmode=disable")

	sqlDSN, driver, migrateDSN, err := getConnectionInformation()

	require.NoError(t, err)
	require.Equal(t, "pgx", driver)
	require.Contains(t, sqlDSN, "host=db-host")
	require.Contains(t, sqlDSN, "dbname=db-name")
	require.Contains(t, sqlDSN, "sslmode=disable", "database.params must reach the real runtime DSN, not just migrateDSN")
	require.Equal(t, "postgres://db-user:db-pass@db-host:5432/db-name?sslmode=disable", migrateDSN)
}

func Test_Unit_GetConnectionInformation_Postgres_NoParams(t *testing.T) {
	setDatabaseTestConfig(t, "postgres")
	config.Set("database.params", "")

	sqlDSN, _, _, err := getConnectionInformation()

	require.NoError(t, err)
	require.NotContains(t, sqlDSN, "  ", "no leftover double space when params is empty")
	require.Regexp(t, `^host=\S+ port=\S+ user=\S+ password=\S+ dbname=\S+$`, sqlDSN)
}

func Test_Unit_GetConnectionInformation_MySQL(t *testing.T) {
	setDatabaseTestConfig(t, "mysql")
	config.Set("database.params", "charset=utf8mb4&parseTime=true")

	sqlDSN, driver, migrateDSN, err := getConnectionInformation()

	require.NoError(t, err)
	require.Equal(t, "mysql", driver)
	require.Equal(t, "db-user:db-pass@tcp(db-host:5432)/db-name?charset=utf8mb4&parseTime=true", sqlDSN)
	require.Contains(t, sqlDSN, "parseTime=true", "without this in the runtime DSN, DATETIME/TIMESTAMP columns can't scan into time.Time")
	require.Equal(t, "mysql://db-user:db-pass@tcp(db-host:5432)/db-name?charset=utf8mb4&parseTime=true", migrateDSN)
}

func Test_Unit_GetConnectionInformation_MySQL_NoParams(t *testing.T) {
	setDatabaseTestConfig(t, "mysql")
	config.Set("database.params", "")

	sqlDSN, _, _, err := getConnectionInformation()

	require.NoError(t, err)
	require.Equal(t, "db-user:db-pass@tcp(db-host:5432)/db-name", sqlDSN, "trailing ? must be trimmed when params is empty")
}

func Test_Unit_BuildSqlStatements_Postgres_ConvertsPlaceholdersAndQuoting(t *testing.T) {
	setDatabaseTestConfig(t, "postgres")

	query, err := buildSqlStatements(`SELECT * FROM "user" WHERE id = ? AND email = ?`)

	require.NoError(t, err)
	require.Equal(t, `SELECT * FROM "user" WHERE id = $1 AND email = $2`, query)
}

func Test_Unit_BuildSqlStatements_MySQL_LeavesPlaceholdersAndBackticksUser(t *testing.T) {
	setDatabaseTestConfig(t, "mysql")

	query, err := buildSqlStatements(`SELECT * FROM "user" WHERE id = ? AND email = ?`)

	require.NoError(t, err)
	require.Equal(t, "SELECT * FROM `user` WHERE id = ? AND email = ?", query)
}

func Test_Unit_SettingKeyColumn(t *testing.T) {
	setDatabaseTestConfig(t, "postgres")
	col, err := settingKeyColumn()
	require.NoError(t, err)
	require.Equal(t, "key", col)

	setDatabaseTestConfig(t, "mysql")
	col, err = settingKeyColumn()
	require.NoError(t, err)
	require.Equal(t, "`key`", col)
}
