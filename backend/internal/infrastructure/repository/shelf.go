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
	// GetByDomain resolves a shelf by the domain it is served on. domain must
	// already be normalized (see domain.NormalizeDomain), which is what makes
	// a plain equality check enough.
	GetByDomain(domain string) (*model.Shelf, error)
	// GetByUsernameAndPath resolves /<username>/<path>, used instead of
	// GetByPath while app.userBasedPaths is enabled.
	GetByUsernameAndPath(username, path string) (*model.Shelf, error)
	// PathInUse reports whether another shelf (not exceptShelfId, which may
	// be empty) already has this path, compared case-insensitively like the
	// public lookup does. PathInUseByUser is the same restricted to one owner.
	PathInUse(path, exceptShelfId string) (bool, error)
	PathInUseByUser(userId, path, exceptShelfId string) (bool, error)
	// DomainInUse reports whether a shelf other than exceptShelfId (which may
	// be empty) already has this normalized domain, so an empty exceptShelfId
	// asks whether any shelf has it.
	DomainInUse(domain, exceptShelfId string) (bool, error)
	// ListPathCollisions returns every shelf whose path is also used by at
	// least one other shelf - only possible after user-based paths were
	// switched off again.
	ListPathCollisions() ([]PathCollision, error)
	Create(s *model.Shelf) (string, error)
	Update(s *model.Shelf) error
	Delete(s *model.Shelf) error
}

// PathCollision is one shelf out of a group sharing the same path.
type PathCollision struct {
	Path     string
	ShelfId  string
	UserId   string
	Username string
}

// shelfSelect is shared by every read so they all fill the owner's username
// the same way. It is a plain fragment rather than a column list so each
// query can append its own WHERE.
const shelfSelect = `
		SELECT s.id, s.title, s.path, s.domain, s.description, s.theme_id, s.icon, s.user_id, u.username, s.created_user_based_paths
		FROM shelf s
		JOIN "user" u ON u.id = s.user_id
`

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

// scanShelf reads a shelf row where path/domain/theme_id may be SQL NULL,
// translating NULL back to "" so nothing above the repository layer has to
// know about it.
func scanShelf(scan func(dest ...any) error) (model.Shelf, error) {
	var (
		shelf    model.Shelf
		path     sql.NullString
		domain   sql.NullString
		themeId  sql.NullString
		username sql.NullString
	)

	err := scan(
		&shelf.Id,
		&shelf.Title,
		&path,
		&domain,
		&shelf.Description,
		&themeId,
		&shelf.Icon,
		&shelf.UserId,
		&username,
		&shelf.CreatedWithUserBasedPaths,
	)
	if err != nil {
		return model.Shelf{}, err
	}

	shelf.Path = path.String
	shelf.Domain = domain.String
	shelf.ThemeId = themeId.String
	shelf.Username = username.String
	return shelf, nil
}

func (r *shelfRepository) List() ([]model.Shelf, error) {
	query, err := buildSqlStatements(shelfSelect)
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
	query, err := buildSqlStatements(shelfSelect + `WHERE s.user_id = ?`)
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
	query, err := buildSqlStatements(shelfSelect + `WHERE s.id = ?`)
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
	query, err := buildSqlStatements(shelfSelect + `WHERE LOWER(s.path) = LOWER(?)`)
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

func (r *shelfRepository) GetByDomain(domain string) (*model.Shelf, error) {
	query, err := buildSqlStatements(shelfSelect + `WHERE s.domain = ?`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, domain)
	shelf, err := scanShelf(row.Scan)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &shelf, nil
}

func (r *shelfRepository) GetByUsernameAndPath(username, path string) (*model.Shelf, error) {
	query, err := buildSqlStatements(shelfSelect + `WHERE u.username = ? AND LOWER(s.path) = LOWER(?)`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, username, path)
	shelf, err := scanShelf(row.Scan)

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &shelf, nil
}

func (r *shelfRepository) PathInUse(path, exceptShelfId string) (bool, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM shelf
		WHERE LOWER(path) = LOWER(?) AND id <> ?
	`)
	if err != nil {
		return false, err
	}

	var count int
	if err := r.Engine.QueryRowContext(context.TODO(), query, path, exceptShelfId).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shelfRepository) PathInUseByUser(userId, path, exceptShelfId string) (bool, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM shelf
		WHERE user_id = ? AND LOWER(path) = LOWER(?) AND id <> ?
	`)
	if err != nil {
		return false, err
	}

	var count int
	if err := r.Engine.QueryRowContext(context.TODO(), query, userId, path, exceptShelfId).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shelfRepository) DomainInUse(domain, exceptShelfId string) (bool, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM shelf
		WHERE domain = ? AND id <> ?
	`)
	if err != nil {
		return false, err
	}

	var count int
	if err := r.Engine.QueryRowContext(context.TODO(), query, domain, exceptShelfId).Scan(&count); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *shelfRepository) ListPathCollisions() ([]PathCollision, error) {
	query, err := buildSqlStatements(`
		SELECT s.path, s.id, s.user_id, u.username
		FROM shelf s
		JOIN "user" u ON u.id = s.user_id
		JOIN (
			SELECT LOWER(path) AS lower_path
			FROM shelf
			WHERE path IS NOT NULL
			GROUP BY lower_path
			HAVING COUNT(*) > 1
		) dup ON dup.lower_path = LOWER(s.path)
		ORDER BY dup.lower_path, s.id
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	collisions := make([]PathCollision, 0)
	for rows.Next() {
		var c PathCollision
		var username sql.NullString
		if err := rows.Scan(&c.Path, &c.ShelfId, &c.UserId, &username); err != nil {
			return nil, err
		}
		c.Username = username.String
		collisions = append(collisions, c)
	}

	return collisions, rows.Err()
}

func (r *shelfRepository) Create(s *model.Shelf) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO shelf (id, title, path, domain, description, theme_id, icon, user_id, created_user_based_paths)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		nullIfEmpty(s.ThemeId),
		s.Icon,
		s.UserId,
		s.CreatedWithUserBasedPaths,
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
			theme_id = ?,
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
		nullIfEmpty(s.ThemeId),
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
