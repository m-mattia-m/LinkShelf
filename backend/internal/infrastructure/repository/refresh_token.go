//go:generate mockgen -source=refresh_token.go -destination=mocks/refresh_token_repository.go -package=mocks

package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

type RefreshToken struct {
	Id        string
	UserId    string
	TokenHash string
	ExpiresAt time.Time
}

type RefreshTokenRepository interface {
	Create(userId, tokenHash string, expiresAt time.Time) error
	GetByHash(tokenHash string) (*RefreshToken, error)
	DeleteByHash(tokenHash string) error
	DeleteByUserId(userId string) error
}

type refreshTokenRepository struct {
	Engine *sql.DB
	Table  string
}

func NewRefreshTokenRepository(engine *sql.DB, table string) (RefreshTokenRepository, error) {
	return &refreshTokenRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *refreshTokenRepository) Create(userId, tokenHash string, expiresAt time.Time) error {
	query, err := buildSqlStatements(`
		INSERT INTO "refresh_token" (id, user_id, token_hash, expires_at)
		VALUES (?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, id.String(), userId, tokenHash, expiresAt)
	return err
}

func (r *refreshTokenRepository) GetByHash(tokenHash string) (*RefreshToken, error) {
	query, err := buildSqlStatements(`
		SELECT id, user_id, token_hash, expires_at
		FROM "refresh_token"
		WHERE token_hash = ?
	`)
	if err != nil {
		return nil, err
	}

	var token RefreshToken
	err = r.Engine.QueryRowContext(context.TODO(), query, tokenHash).Scan(
		&token.Id,
		&token.UserId,
		&token.TokenHash,
		&token.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &token, err
}

func (r *refreshTokenRepository) DeleteByHash(tokenHash string) error {
	query, err := buildSqlStatements(`
		DELETE FROM "refresh_token"
		WHERE token_hash = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, tokenHash)
	return err
}

func (r *refreshTokenRepository) DeleteByUserId(userId string) error {
	query, err := buildSqlStatements(`
		DELETE FROM "refresh_token"
		WHERE user_id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, userId)
	return err
}
