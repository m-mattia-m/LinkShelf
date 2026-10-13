//go:generate mockgen -source=shelf.go -destination=mocks/shelf_repository.go -package=mocks

package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type ShelfRepository interface {
	List() ([]model.Shelf, error)
	ListByUserId(userId string) ([]model.Shelf, error)
	Get(id string) (*model.Shelf, error)
	GetByPath(path string) (*model.Shelf, error)
	// GetByDomain resolves a shelf by its normalized domain.
	GetByDomain(domain string) (*model.Shelf, error)
	// GetByUsernameAndPath resolves /<username>/<path>.
	GetByUsernameAndPath(username, path string) (*model.Shelf, error)
	// PathInUse reports whether another shelf has this path (case-insensitive).
	PathInUse(path, exceptShelfId string) (bool, error)
	PathInUseByUser(userId, path, exceptShelfId string) (bool, error)
	// GetReadyByDomain resolves a shelf by its normalized domain, only while its state is ready.
	GetReadyByDomain(domain string) (*model.Shelf, error)
	// DomainInUse reports whether a shelf other than exceptShelfId has this domain.
	DomainInUse(domain, exceptShelfId string) (bool, error)
	// DomainReadyInUse is DomainInUse, counting only shelves whose domain state is ready.
	DomainReadyInUse(domain, exceptShelfId string) (bool, error)
	// ListDomains returns the shelves with a domain ordered by id, starting after afterId, at most limit.
	ListDomains(afterId string, limit int) ([]model.Shelf, error)
	// SetDomainStatus stores the status of a shelf's domain. It reports false when the
	// shelf does not exist or does not have domain (anymore).
	SetDomainStatus(shelfId, domain string, status model.DomainStatus, now int64) (bool, error)
	// ListPathCollisions returns shelves sharing a path.
	ListPathCollisions() ([]PathCollision, error)
	CountByUserId(userId string) (int, error)
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

// shelfSelect is the shared SELECT for shelf reads, including the owner's username.
const shelfSelect = `
		SELECT s.id, s.title, s.path, s.domain, s.description, s.theme_id, s.icon, s.user_id, u.username, s.created_user_based_paths, s.no_index, s.footer_enabled, s.footer_custom_text,
			s.domain_state, s.domain_reason, s.domain_message, s.domain_records, s.domain_updated_at
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

// nullIfEmpty stores "" as NULL so UNIQUE allows many unset values.
func nullIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// scanShelf reads a shelf row, mapping NULLs to "".
func scanShelf(scan func(dest ...any) error) (model.Shelf, error) {
	var (
		shelf    model.Shelf
		path     sql.NullString
		domain   sql.NullString
		themeId  sql.NullString
		username sql.NullString
		records  string
		state    string
		reason   string
		message  string
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
		&shelf.NoIndex,
		&shelf.FooterEnabled,
		&shelf.FooterCustomText,
		&state,
		&reason,
		&message,
		&records,
		&shelf.DomainUpdatedAt,
	)
	if err != nil {
		return model.Shelf{}, err
	}

	if state != "" {
		status := &model.DomainStatus{State: state, Reason: reason, Message: message, Records: []model.DomainRecord{}}
		if records != "" {
			// A damaged value only costs the records, never the shelf.
			_ = json.Unmarshal([]byte(records), &status.Records)
		}
		shelf.DomainStatus = status
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

func (r *shelfRepository) GetReadyByDomain(domain string) (*model.Shelf, error) {
	query, err := buildSqlStatements(shelfSelect + `WHERE s.domain = ? AND s.domain_state = 'ready'`)
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

func (r *shelfRepository) DomainReadyInUse(domain, exceptShelfId string) (bool, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM shelf
		WHERE domain = ? AND domain_state = 'ready' AND id <> ?
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

func (r *shelfRepository) ListDomains(afterId string, limit int) ([]model.Shelf, error) {
	query, err := buildSqlStatements(shelfSelect + `WHERE s.domain IS NOT NULL AND s.id > ? ORDER BY s.id LIMIT ?`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query, afterId, limit)
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
	return shelves, rows.Err()
}

func (r *shelfRepository) SetDomainStatus(shelfId, domain string, status model.DomainStatus, now int64) (bool, error) {
	query, err := buildSqlStatements(`
		UPDATE shelf
		SET domain_state = ?,
			domain_reason = ?,
			domain_message = ?,
			domain_records = ?,
			domain_updated_at = ?
		WHERE id = ? AND domain = ?
	`)
	if err != nil {
		return false, err
	}

	records, err := json.Marshal(status.Records)
	if err != nil {
		return false, err
	}

	result, err := r.Engine.ExecContext(context.TODO(), query,
		status.State, status.Reason, status.Message, string(records), now, shelfId, domain)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
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
		INSERT INTO shelf (id, title, path, domain, description, theme_id, icon, user_id, created_user_based_paths, no_index, footer_enabled, footer_custom_text)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
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
		s.NoIndex,
		s.FooterEnabled,
		s.FooterCustomText,
	)
	if err != nil {
		return "", err
	}

	return s.Id, nil
}

func (r *shelfRepository) Update(s *model.Shelf) error {
	query, err := buildSqlStatements(`
		UPDATE shelf
		SET domain_state = CASE WHEN COALESCE(domain, '') = ? THEN domain_state ELSE '' END,
			domain_reason = CASE WHEN COALESCE(domain, '') = ? THEN domain_reason ELSE '' END,
			domain_message = CASE WHEN COALESCE(domain, '') = ? THEN domain_message ELSE '' END,
			domain_records = CASE WHEN COALESCE(domain, '') = ? THEN domain_records ELSE '' END,
			domain_updated_at = CASE WHEN COALESCE(domain, '') = ? THEN domain_updated_at ELSE ? END,
			title = ?,
			path = ?,
			domain = ?,
			description = ?,
			theme_id = ?,
			icon = ?,
			no_index = ?,
			footer_enabled = ?,
			footer_custom_text = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		s.Domain, s.Domain, s.Domain, s.Domain, s.Domain, time.Now().Unix(),
		s.Title,
		nullIfEmpty(s.Path),
		nullIfEmpty(s.Domain),
		s.Description,
		nullIfEmpty(s.ThemeId),
		s.Icon,
		s.NoIndex,
		s.FooterEnabled,
		s.FooterCustomText,
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

func (r *shelfRepository) CountByUserId(userId string) (int, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM shelf
		WHERE user_id = ?
	`)
	if err != nil {
		return 0, err
	}

	var count int
	err = r.Engine.QueryRowContext(context.TODO(), query, userId).Scan(&count)
	return count, err
}
