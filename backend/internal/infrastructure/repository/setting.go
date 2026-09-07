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
	Upsert(key string, language string, value string) error
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

// settingKeyColumn is "key" quoted for whichever engine is configured - MySQL
// reserves KEY (rejects it as a bare column reference), Postgres doesn't.
func settingKeyColumn() (string, error) {
	_, driver, _, err := getConnectionInformation()
	if err != nil {
		return "", err
	}
	if driver == "mysql" {
		return "`key`", nil
	}
	return "key", nil
}

func (r *settingRepository) List() ([]model.Setting, error) {
	keyColumn, err := settingKeyColumn()
	if err != nil {
		return nil, err
	}

	query, err := buildSqlStatements(`
		SELECT ` + keyColumn + `, language, value
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
	keyColumn, err := settingKeyColumn()
	if err != nil {
		return nil, err
	}

	query, err := buildSqlStatements(`
		SELECT ` + keyColumn + `, language, value
		FROM setting
		WHERE ` + keyColumn + ` = ?
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

// Upsert's conflict-handling clause has no shared syntax between engines
// (Postgres's ON CONFLICT/EXCLUDED vs MySQL's ON DUPLICATE KEY UPDATE), so
// this builds the two dialects' queries directly rather than going through
// buildSqlStatements's generic ?-to-$N conversion.
func (r *settingRepository) Upsert(key string, language string, value string) error {
	_, driver, _, err := getConnectionInformation()
	if err != nil {
		return err
	}

	var query string
	if driver == "mysql" {
		query = "INSERT INTO setting (`key`, language, value) VALUES (?, ?, ?) ON DUPLICATE KEY UPDATE value = VALUES(value)"
	} else {
		query = "INSERT INTO setting (key, language, value) VALUES ($1, $2, $3) ON CONFLICT (key, language) DO UPDATE SET value = EXCLUDED.value"
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, key, language, value)
	if err != nil {
		return err
	}

	return nil
}
