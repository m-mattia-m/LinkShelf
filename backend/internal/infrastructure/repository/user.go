package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type UserRepository interface {
	List() ([]model.User, error)
	Get(id string) (*model.User, error)
	GetPassword(id string) (string, error)
	Create(u *model.User) (string, error)
	Update(u *model.User) error
	PatchPassword(u *model.User) error
	Delete(u *model.User) error
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
		SELECT *
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

	var users []model.User
	for rows.Next() {
		var user model.User
		err := rows.Scan(
			&user.Id,
			&user.Email,
			&user.FirstName,
			&user.LastName,
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
		SELECT *
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
		&user.Password,
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

func (r *userRepository) Create(u *model.User) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO "user" (id, email, first_name, last_name, password)
		VALUES (?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}

	u.Id = uuid.New().String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		u.Id,
		u.Email,
		u.FirstName,
		u.LastName,
		u.Password,
	)
	if err != nil {
		return "", err
	}

	return u.Id, nil
}

func (r *userRepository) Update(u *model.User) error {
	query, err := buildSqlStatements(`
		UPDATE "user"
		SET email = ?, 
		 	first_name = ?, 
			last_name = ?
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
		u.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *userRepository) PatchPassword(u *model.User) error {
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
		u.Password,
		u.Id,
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
