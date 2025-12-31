//go:generate mockgen -source=setting.go -destination=mocks/setting_repository.go -package=mocks

package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"
)

type SettingRepository interface {
	List() ([]model.Setting, error)
	GetByKey(key string) (*model.Setting, error)
	Update(key string, value string) error
}

type settingRepository struct {
	Engine *sql.DB
	Table  string
}

func NewSettingRepository(engine *sql.DB, table string) (SettingRepository, error) {
	return &settingRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func (r *settingRepository) List() ([]model.Setting, error) {
	query, err := buildSqlStatements(`
		SELECT *
		FROM setting
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query)
	if err != nil {
		return nil, err
	}

	var settings []model.Setting
	for rows.Next() {
		var setting model.Setting
		err := rows.Scan(
			&setting.Key,
			&setting.LanguageCode,
			&setting.Value,
		)
		if err != nil {
			return nil, err
		}
		settings = append(settings, setting)
	}

	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}

	return settings, err
}

func (r *settingRepository) GetByKey(key string) (*model.Setting, error) {

	query, err := buildSqlStatements(`
		SELECT *
		FROM setting
		WHERE key = ?
	`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, key)

	var setting model.Setting
	err = row.Scan(
		&setting.Key,
		&setting.LanguageCode,
		&setting.Value,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	return &setting, nil
}

func (r *settingRepository) Update(key string, value string) error {
	query, err := buildSqlStatements(`
		UPDATE setting
		SET value = ?
		WHERE key = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, value, key)
	if err != nil {
		return err
	}

	return nil
}
