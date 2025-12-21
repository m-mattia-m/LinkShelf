package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type SectionRepository interface {
	ListByShelfId(id string) ([]model.Section, error)
	Get(id string) (*model.Section, error)
	Create(s *model.Section) (string, error)
	Update(s *model.Section) error
	Delete(s *model.Section) error
}

type sectionRepository struct {
	Engine *sql.DB
	Table  string
}

func NewSectionRepository(engine *sql.DB, table string) (SectionRepository, error) {
	return &sectionRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *sectionRepository) ListByShelfId(id string) ([]model.Section, error) {
	query, err := buildSqlStatements(`
		SELECT *
		FROM section
		WHERE shelf_id = ?
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query, id)
	if err != nil {
		return nil, err
	}

	var sections []model.Section
	for rows.Next() {
		var section model.Section
		err := rows.Scan(
			&section.Id,
			&section.Title,
			&section.ShelfId,
		)
		if err != nil {
			return nil, err
		}
		sections = append(sections, section)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return sections, err
}

func (r *sectionRepository) Get(id string) (*model.Section, error) {
	query, err := buildSqlStatements(`
		SELECT *
		FROM section
		WHERE id = ?
		LIMIT 1
	`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, id)

	var section model.Section
	err = row.Scan(
		&section.Id,
		&section.Title,
		&section.ShelfId,
	)
	if err != nil {
		return nil, err
	}

	return &section, nil
}

func (r *sectionRepository) Create(s *model.Section) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO section (id, title, shelf_id)
		VALUES (?, ?, ?)
	`)
	if err != nil {
		return "", err
	}

	s.Id = uuid.New().String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		s.Id,
		s.Title,
		s.ShelfId,
	)
	if err != nil {
		return "", err
	}

	return s.Id, nil
}

func (r *sectionRepository) Update(s *model.Section) error {
	query, err := buildSqlStatements(`
		UPDATE section
		SET title = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		s.Title,
		s.Id,
	)
	if err != nil {
		return err
	}

	return nil
}

func (r *sectionRepository) Delete(s *model.Section) error {
	query, err := buildSqlStatements(`
		DELETE FROM section
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
