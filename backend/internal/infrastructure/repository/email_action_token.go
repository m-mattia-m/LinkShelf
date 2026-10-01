//go:generate mockgen -source=email_action_token.go -destination=mocks/email_action_token_repository.go -package=mocks

package repository

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
)

const (
	EmailActionVerify      = "verify"
	EmailActionSetPassword = "set_password"
	// EmailActionResetPassword is a forgot-password link.
	EmailActionResetPassword = "reset_password"
	// EmailActionChangeEmail confirms a pending email change.
	EmailActionChangeEmail = "change_email"
)

type EmailActionToken struct {
	Id        string
	UserId    string
	TokenHash string
	Action    string
	ExpiresAt time.Time
	CreatedAt time.Time
}

type EmailActionTokenRepository interface {
	Create(userId, tokenHash, action string, expiresAt time.Time) error
	GetByHash(tokenHash string) (*EmailActionToken, error)
	// GetLatestByUserIdAndAction is used to rate-limit resends.
	GetLatestByUserIdAndAction(userId, action string) (*EmailActionToken, error)
	DeleteByUserIdAndAction(userId, action string) error
}

type emailActionTokenRepository struct {
	Engine *sql.DB
	Table  string
}

func NewEmailActionTokenRepository(engine *sql.DB, table string) (EmailActionTokenRepository, error) {
	return &emailActionTokenRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *emailActionTokenRepository) Create(userId, tokenHash, action string, expiresAt time.Time) error {
	query, err := buildSqlStatements(`
		INSERT INTO email_action_token (id, user_id, token_hash, action, expires_at)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return err
	}

	id, err := uuid.NewV7()
	if err != nil {
		return err
	}

	// expires_at is timezone-naive, so store UTC.
	_, err = r.Engine.ExecContext(context.TODO(), query, id.String(), userId, tokenHash, action, expiresAt.UTC())
	return err
}

func (r *emailActionTokenRepository) GetByHash(tokenHash string) (*EmailActionToken, error) {
	query, err := buildSqlStatements(`
		SELECT id, user_id, token_hash, action, expires_at, created_at
		FROM email_action_token
		WHERE token_hash = ?
	`)
	if err != nil {
		return nil, err
	}

	var token EmailActionToken
	err = r.Engine.QueryRowContext(context.TODO(), query, tokenHash).Scan(
		&token.Id,
		&token.UserId,
		&token.TokenHash,
		&token.Action,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *emailActionTokenRepository) GetLatestByUserIdAndAction(userId, action string) (*EmailActionToken, error) {
	query, err := buildSqlStatements(`
		SELECT id, user_id, token_hash, action, expires_at, created_at
		FROM email_action_token
		WHERE user_id = ? AND action = ?
		ORDER BY created_at DESC
		LIMIT 1
	`)
	if err != nil {
		return nil, err
	}

	var token EmailActionToken
	err = r.Engine.QueryRowContext(context.TODO(), query, userId, action).Scan(
		&token.Id,
		&token.UserId,
		&token.TokenHash,
		&token.Action,
		&token.ExpiresAt,
		&token.CreatedAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (r *emailActionTokenRepository) DeleteByUserIdAndAction(userId, action string) error {
	query, err := buildSqlStatements(`
		DELETE FROM email_action_token
		WHERE user_id = ? AND action = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, userId, action)
	return err
}
