package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/google/uuid"
)

type LinkRepository interface {
	ListByShelfId(id string) ([]model.Link, error)
	Get(id string) (*model.Link, error)
	Create(l *model.Link) (string, error)
	Update(l *model.Link) error
	Delete(l *model.Link) error
}

type linkRepository struct {
	Engine *sql.DB
	Table  string
}

func NewLinkRepository(engine *sql.DB, table string) (LinkRepository, error) {
	return &linkRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *linkRepository) ListByShelfId(id string) ([]model.Link, error) {
	query, err := buildSqlStatements(fmt.Sprintf(`
		SELECT l.*
		FROM link l
		JOIN section s ON l.section_id = s.id
		WHERE s.shelf_id = %s;
	`, id))
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	var links []model.Link
	for rows.Next() {
		var link model.Link
		err := rows.Scan(
			&link.Id,
			&link.Title,
			&link.Link,
			&link.Icon,
			&link.Color,
			&link.SectionId,
		)
		if err != nil {
			return nil, err
		}
		links = append(links, link)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return links, err
}

func (r *linkRepository) Get(id string) (*model.Link, error) {
	query, err := buildSqlStatements(fmt.Sprintf(`
		SELECT *
		FROM link
		WHERE id = %s
		LIMIT 1
	`, id))
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query)

	var link model.Link
	err = row.Scan(
		&link.Id,
		&link.Title,
		&link.Link,
		&link.Icon,
		&link.Color,
		&link.SectionId,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	return &link, nil
}

func (r *linkRepository) Create(l *model.Link) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO links (id, title, link, icon, color, section_id)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}

	l.Id = uuid.New().String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		l.Id,
		l.Title,
		l.Link,
		l.Icon,
		l.Color,
		l.SectionId,
	)
	if err != nil {
		return "", err
	}

	return l.Id, nil
}

func (r *linkRepository) Update(l *model.Link) error {
	query, err := buildSqlStatements(`
		UPDATE links
		SET title = ?,
			link = ?,
			icon = ?,
			color = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		l.Title,
		l.Link,
		l.Icon,
		l.Color,
		l.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *linkRepository) Delete(l *model.Link) error {
	query, err := buildSqlStatements(`
		DELETE FROM links
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		l.Id,
	)
	if err != nil {
		return err
	}

	return nil
}
