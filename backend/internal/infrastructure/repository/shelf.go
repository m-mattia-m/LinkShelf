//go:generate mockgen -source=shelf.go -destination=mocks/shelf_repository.go -package=mocks

package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type ShelfRepository interface {
	List() ([]model.Shelf, error)
	Get(id string) (*model.Shelf, error)
	Create(s *model.Shelf) (string, error)
	Update(s *model.Shelf) error
	Delete(s *model.Shelf) error
}

type shelfRepository struct {
	Engine *sql.DB
	Table  string
}

func NewShelfRepository(engine *sql.DB, table string) (ShelfRepository, error) {
	return &shelfRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *shelfRepository) List() ([]model.Shelf, error) {
	query, err := buildSqlStatements(`
		SELECT *
		FROM shelf
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shelves := make([]model.Shelf, 0)

	for rows.Next() {
		var shelf model.Shelf
		err := rows.Scan(
			&shelf.Id,
			&shelf.Title,
			&shelf.Description,
			&shelf.Theme,
			&shelf.Icon,
			&shelf.UserId,
		)
		if err != nil {
			return nil, err
		}

		shelves = append(shelves, shelf)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return shelves, nil
}

func (r *shelfRepository) Get(id string) (*model.Shelf, error) {
	query, err := buildSqlStatements(`
		SELECT *
		FROM shelf
		WHERE id = ?
	`)
	if err != nil {
		return nil, err
	}

	var shelf model.Shelf
	err = r.Engine.QueryRowContext(context.TODO(), query, id).Scan(
		&shelf.Id,
		&shelf.Title,
		&shelf.Path,
		&shelf.Domain,
		&shelf.Description,
		&shelf.Theme,
		&shelf.Icon,
		&shelf.UserId,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &shelf, err
}

func (r *shelfRepository) Create(s *model.Shelf) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO shelf (id, title, path, domain, description, theme, icon, user_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}

	generatedShelfId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	s.Id = generatedShelfId.String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		s.Id,
		s.Title,
		s.Path,
		s.Domain,
		s.Description,
		s.Theme,
		s.Icon,
		s.UserId,
	)
	if err != nil {
		return "", err
	}

	return s.Id, nil
}

func (r *shelfRepository) Update(s *model.Shelf) error {
	query, err := buildSqlStatements(`
		UPDATE shelf
		SET title = ?,
			path = ?,
			domain = ?,
			description = ?,
			theme = ?,
			icon = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		s.Title,
		s.Path,
		s.Domain,
		s.Description,
		s.Theme,
		s.Icon,
		s.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *shelfRepository) Delete(s *model.Shelf) error {
	query, err := buildSqlStatements(`
		DELETE FROM shelf
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		s.Id,
	)
	if err != nil {
		return err
	}

	return nil
}
