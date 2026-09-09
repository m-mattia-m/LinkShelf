//go:build !realdb

package repository

// setupRealDB is a no-op outside the realdb-tagged test tier - see
// realdb_setup_test.go (built only with -tags=realdb) for the real
// Postgres/MySQL container implementation.
func setupRealDB() (cleanup func(), err error) {
	return func() {}, nil
}
