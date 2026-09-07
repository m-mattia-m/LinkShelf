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
	ListByUserId(userId string) ([]model.Shelf, error)
	Get(id string) (*model.Shelf, error)
	GetByPath(path string) (*model.Shelf, error)
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

// nullIfEmpty maps "" to a SQL NULL, since path/domain are stored as NULL
// (not "") when unset - a plain UNIQUE constraint then allows any number of
// shelves to leave them unset, on both Postgres and MySQL. See
// migrations/postgres/0003_dashboard.up.sql for why.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// scanShelf reads a shelf row where path/domain may be SQL NULL, translating
// NULL back to "" so nothing above the repository layer has to know about it.
func scanShelf(scan func(dest ...any) error) (model.Shelf, error) {
	var (
		shelf  model.Shelf
		path   sql.NullString
		domain sql.NullString
	)

	err := scan(
		&shelf.Id,
		&shelf.Title,
		&path,
		&domain,
		&shelf.Description,
		&shelf.Theme,
		&shelf.Icon,
		&shelf.UserId,
	)
	if err != nil {
		return model.Shelf{}, err
	}

	shelf.Path = path.String
	shelf.Domain = domain.String
	return shelf, nil
}

func (r *shelfRepository) List() ([]model.Shelf, error) {
	query, err := buildSqlStatements(`
		SELECT id, title, path, domain, description, theme, icon, user_id
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
		shelf, err := scanShelf(rows.Scan)
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

func (r *shelfRepository) ListByUserId(userId string) ([]model.Shelf, error) {
	query, err := buildSqlStatements(`
		SELECT id, title, path, domain, description, theme, icon, user_id
		FROM shelf
		WHERE user_id = ?
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	shelves := make([]model.Shelf, 0)

	for rows.Next() {
		shelf, err := scanShelf(rows.Scan)
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
		SELECT id, title, path, domain, description, theme, icon, user_id
		FROM shelf
		WHERE id = ?
	`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, id)
	shelf, err := scanShelf(row.Scan)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &shelf, nil
}

func (r *shelfRepository) GetByPath(path string) (*model.Shelf, error) {
	query, err := buildSqlStatements(`
		SELECT id, title, path, domain, description, theme, icon, user_id
		FROM shelf
		WHERE LOWER(path) = LOWER(?)
	`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, path)
	shelf, err := scanShelf(row.Scan)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &shelf, nil
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
		nullIfEmpty(s.Path),
		nullIfEmpty(s.Domain),
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
		nullIfEmpty(s.Path),
		nullIfEmpty(s.Domain),
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
