//go:generate mockgen -source=user.go -destination=mocks/user_repository.go -package=mocks

package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

// AuthRecord carries the fields needed to authenticate and authorize a user.
// It intentionally lives outside the API model package since it must never be
// serialized back to a client (it carries the password hash).
type AuthRecord struct {
	Id         string
	Email      string
	FirstName  string
	LastName   string
	Role       string
	Password   string
	Provider   string
	ProviderId *string
}

type UserRepository interface {
	List() ([]model.User, error)
	Get(id string) (*model.User, error)
	GetPassword(id string) (string, error)
	Create(u model.UserBase, hashedPassword, role string) (string, error)
	Update(u *model.User) error
	PatchPassword(id string, hashedPassword string) error
	Delete(u *model.User) error

	FindByEmail(email string) (*AuthRecord, error)
	FindByProviderId(providerId string) (*AuthRecord, error)
	CreateExternal(email, firstName, lastName, provider, providerId string) (string, error)
	LinkProvider(userId, provider, providerId string) error
	SetPasswordAndRole(userId, hashedPassword, role string) error
}

type userRepository struct {
	Engine *sql.DB
	Table  string
}

func NewUserRepository(engine *sql.DB, table string) (UserRepository, error) {

	return &userRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *userRepository) List() ([]model.User, error) {
	query, err := buildSqlStatements(`
		SELECT id, email, first_name, last_name, role
		FROM "user"
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]model.User, 0)
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.FirstName,
			&user.LastName,
			&user.Role,
		)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}

func (r *userRepository) Get(id string) (*model.User, error) {
	query, err := buildSqlStatements(`
		SELECT id, email, first_name, last_name, role
		FROM "user"
		WHERE id = ?
	`)
	if err != nil {
		return nil, err
	}

	var user model.User
	err = r.Engine.QueryRowContext(context.TODO(), query, id).Scan(
		&user.Id,
		&user.Email,
		&user.FirstName,
		&user.LastName,
		&user.Role,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &user, err
}

func (r *userRepository) GetPassword(id string) (string, error) {
	query, err := buildSqlStatements(`
		SELECT password
		FROM "user"
		WHERE id = ?
	`)
	if err != nil {
		return "", err
	}

	var password string
	err = r.Engine.QueryRowContext(context.TODO(), query, id).Scan(
		&password,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}

	return password, err
}

func (r *userRepository) Create(u model.UserBase, hashedPassword, role string) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO "user" (id, email, first_name, last_name, password, provider, role)
		VALUES (?, ?, ?, ?, ?, 'LOCAL', ?)
	`)
	if err != nil {
		return "", err
	}

	generatedUserId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	id := generatedUserId.String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		id,
		u.Email,
		u.FirstName,
		u.LastName,
		hashedPassword,
		role,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

func (r *userRepository) Update(u *model.User) error {
	query, err := buildSqlStatements(`
		UPDATE "user"
		SET email = ?,
		 	first_name = ?,
			last_name = ?,
			role = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		u.Email,
		u.FirstName,
		u.LastName,
		u.Role,
		u.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) PatchPassword(id string, hashedPassword string) error {
	query, err := buildSqlStatements(`
		UPDATE "user"
		SET password = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		hashedPassword,
		id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) Delete(u *model.User) error {
	query, err := buildSqlStatements(`
		DELETE FROM "user"
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		u.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) FindByEmail(email string) (*AuthRecord, error) {
	query, err := buildSqlStatements(`
		SELECT id, email, first_name, last_name, role, password, provider, provider_id
		FROM "user"
		WHERE LOWER(email) = LOWER(?)
	`)
	if err != nil {
		return nil, err
	}

	var record AuthRecord
	err = r.Engine.QueryRowContext(context.TODO(), query, email).Scan(
		&record.Id,
		&record.Email,
		&record.FirstName,
		&record.LastName,
		&record.Role,
		&record.Password,
		&record.Provider,
		&record.ProviderId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &record, err
}

func (r *userRepository) FindByProviderId(providerId string) (*AuthRecord, error) {
	query, err := buildSqlStatements(`
		SELECT id, email, first_name, last_name, role, password, provider, provider_id
		FROM "user"
		WHERE provider_id = ?
	`)
	if err != nil {
		return nil, err
	}

	var record AuthRecord
	err = r.Engine.QueryRowContext(context.TODO(), query, providerId).Scan(
		&record.Id,
		&record.Email,
		&record.FirstName,
		&record.LastName,
		&record.Role,
		&record.Password,
		&record.Provider,
		&record.ProviderId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &record, err
}

// CreateExternal creates a user with no local password, provisioned from an
// external identity provider login.
func (r *userRepository) CreateExternal(email, firstName, lastName, provider, providerId string) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO "user" (id, email, first_name, last_name, password, provider, provider_id)
		VALUES (?, ?, ?, ?, '', ?, ?)
	`)
	if err != nil {
		return "", err
	}

	generatedUserId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	id := generatedUserId.String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		id,
		email,
		firstName,
		lastName,
		provider,
		providerId,
	)
	if err != nil {
		return "", err
	}

	return id, nil
}

// LinkProvider attaches an external identity to an already-existing user,
// without touching their existing local password.
func (r *userRepository) LinkProvider(userId, provider, providerId string) error {
	query, err := buildSqlStatements(`
		UPDATE "user"
		SET provider = ?,
			provider_id = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		provider,
		providerId,
		userId,
	)
	return err
}

// SetPasswordAndRole is used exclusively to create/refresh the config-driven
// bootstrap admin account idempotently.
func (r *userRepository) SetPasswordAndRole(userId, hashedPassword, role string) error {
	query, err := buildSqlStatements(`
		UPDATE "user"
		SET password = ?,
			role = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		hashedPassword,
		role,
		userId,
	)
	return err
}
