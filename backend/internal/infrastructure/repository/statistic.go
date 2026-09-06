//go:generate mockgen -source=statistic.go -destination=mocks/statistic_repository.go -package=mocks

package repository

import (
	"context"
	"database/sql"
	"errors"
)

type StatisticRepository interface {
	GetShelfAmount(userId string) (*int, error)
	GetSectionAmount(userId string) (*int, error)
	GetLinkAmount(userId string) (*int, error)
}

type statisticRepository struct {
	Engine *sql.DB
}

func NewStatisticRepository(engine *sql.DB) (StatisticRepository, error) {
	return &statisticRepository{
		Engine: engine,
	}, nil
}

func (r *statisticRepository) GetShelfAmount(userId string) (*int, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM shelf
		WHERE user_id = ?
	`)

	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, userId)

	var count int
	err = row.Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &count, err
}

func (r *statisticRepository) GetSectionAmount(userId string) (*int, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM section
		WHERE shelf_id IN (SELECT id FROM shelf WHERE user_id = ?)
	`)
	if err != nil {
		return nil, err
	}
	row := r.Engine.QueryRowContext(context.TODO(), query, userId)

	var count int
	err = row.Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &count, err
}

func (r *statisticRepository) GetLinkAmount(userId string) (*int, error) {
	query, err := buildSqlStatements(`
		SELECT COUNT(*)
		FROM link
		WHERE section_id IN (SELECT id FROM section WHERE shelf_id IN (SELECT id FROM shelf WHERE user_id = ?))
	`)
	if err != nil {
		return nil, err
	}
	row := r.Engine.QueryRowContext(context.TODO(), query, userId)

	var count int
	err = row.Scan(&count)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return &count, err
}
