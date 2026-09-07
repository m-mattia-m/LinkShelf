//go:generate mockgen -source=oidc_state.go -destination=mocks/oidc_state_repository.go -package=mocks

package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"
)

// OidcState is a single-use PKCE state/verifier pair for an in-flight OIDC
// login. It's persisted (not kept in memory) so multiple backend replicas can
// all handle the callback regardless of which one issued the authorization URL.
type OidcState struct {
	State        string
	CodeVerifier string
	ExpiresAt    time.Time
}

type OidcStateRepository interface {
	Create(state, codeVerifier string, expiresAt time.Time) error
	GetByState(state string) (*OidcState, error)
	DeleteByState(state string) error
}

type oidcStateRepository struct {
	Engine *sql.DB
	Table  string
}

func NewOidcStateRepository(engine *sql.DB, table string) (OidcStateRepository, error) {
	return &oidcStateRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *oidcStateRepository) Create(state, codeVerifier string, expiresAt time.Time) error {
	query, err := buildSqlStatements(`
		INSERT INTO "oidc_state" (state, code_verifier, expires_at)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, state, codeVerifier, expiresAt)
	return err
}

func (r *oidcStateRepository) GetByState(state string) (*OidcState, error) {
	query, err := buildSqlStatements(`
		SELECT state, code_verifier, expires_at
		FROM "oidc_state"
		WHERE state = ?
	`)
	if err != nil {
		return nil, err
	}

	var result OidcState
	err = r.Engine.QueryRowContext(context.TODO(), query, state).Scan(
		&result.State,
		&result.CodeVerifier,
		&result.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &result, err
}

func (r *oidcStateRepository) DeleteByState(state string) error {
	query, err := buildSqlStatements(`
		DELETE FROM "oidc_state"
		WHERE state = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, state)
	return err
}
