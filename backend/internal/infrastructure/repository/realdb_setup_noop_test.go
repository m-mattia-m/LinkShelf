//go:build !realdb

package repository

// setupRealDB is a no-op without -tags=realdb.
func setupRealDB() (cleanup func(), err error) {
	return func() {}, nil
}
