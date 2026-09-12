//go:generate mockgen -source=theme.go -destination=mocks/theme_repository.go -package=mocks

package repository

import (
	"backend/internal/infrastructure/api/model"
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
)

type ThemeRepository interface {
	Get(id string) (*model.Theme, error)
	GetBySourceFile(sourceFile string) (*model.Theme, error)
	ListInstance() ([]model.Theme, error)
	ListByOwner(ownerUserId string) ([]model.Theme, error)
	ListAllUserScoped() ([]model.Theme, error)
	Create(t *model.Theme) (string, error)
	Update(t *model.Theme) error
	Delete(id string) error
	// UpsertInstanceBySourceFile creates or updates the instance-scoped theme
	// matching sourceFile, used by the startup directory sync.
	UpsertInstanceBySourceFile(t *model.Theme) error
	// DeleteInstanceNotIn removes every instance-scoped theme whose
	// source_file is not in keepSourceFiles, used by the startup directory
	// sync to drop themes whose file was removed. An empty keepSourceFiles
	// removes all instance-scoped themes.
	DeleteInstanceNotIn(keepSourceFiles []string) error
}

type themeRepository struct {
	Engine *sql.DB
	Table  string
}

func NewThemeRepository(engine *sql.DB, table string) (ThemeRepository, error) {
	return &themeRepository{
		Engine: engine,
		Table:  table,
	}, nil
}

func scanTheme(row interface{ Scan(...any) error }) (*model.Theme, error) {
	var t model.Theme
	var ownerUserId, sourceFile sql.NullString

	err := row.Scan(&t.Id, &t.Scope, &ownerUserId, &t.Name, &sourceFile, &t.Config)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}

	t.OwnerUserId = ownerUserId.String
	t.SourceFile = sourceFile.String
	return &t, nil
}

func (r *themeRepository) Get(id string) (*model.Theme, error) {
	query, err := buildSqlStatements(`
		SELECT id, scope, owner_user_id, name, source_file, config
		FROM theme
		WHERE id = ?
		LIMIT 1
	`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, id)
	return scanTheme(row)
}

func (r *themeRepository) GetBySourceFile(sourceFile string) (*model.Theme, error) {
	query, err := buildSqlStatements(`
		SELECT id, scope, owner_user_id, name, source_file, config
		FROM theme
		WHERE source_file = ?
		LIMIT 1
	`)
	if err != nil {
		return nil, err
	}

	row := r.Engine.QueryRowContext(context.TODO(), query, sourceFile)
	return scanTheme(row)
}

func (r *themeRepository) listWhere(condition string, args ...any) ([]model.Theme, error) {
	query, err := buildSqlStatements(`
		SELECT id, scope, owner_user_id, name, source_file, config
		FROM theme
		WHERE ` + condition + `
		ORDER BY name
	`)
	if err != nil {
		return nil, err
	}

	rows, err := r.Engine.QueryContext(context.TODO(), query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	themes := make([]model.Theme, 0)
	for rows.Next() {
		t, err := scanTheme(rows)
		if err != nil {
			return nil, err
		}
		themes = append(themes, *t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return themes, nil
}

func (r *themeRepository) ListInstance() ([]model.Theme, error) {
	return r.listWhere("scope = ?", model.ThemeScopeInstance)
}

func (r *themeRepository) ListByOwner(ownerUserId string) ([]model.Theme, error) {
	return r.listWhere("scope = ? AND owner_user_id = ?", model.ThemeScopeUser, ownerUserId)
}

func (r *themeRepository) ListAllUserScoped() ([]model.Theme, error) {
	return r.listWhere("scope = ?", model.ThemeScopeUser)
}

func (r *themeRepository) Create(t *model.Theme) (string, error) {
	query, err := buildSqlStatements(`
		INSERT INTO theme (id, scope, owner_user_id, name, source_file, config)
		VALUES (?, ?, ?, ?, ?, ?)
	`)
	if err != nil {
		return "", err
	}

	generatedId, err := uuid.NewV7()
	if err != nil {
		return "", err
	}
	t.Id = generatedId.String()

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		t.Id,
		t.Scope,
		nullableString(t.OwnerUserId),
		t.Name,
		nullableString(t.SourceFile),
		t.Config,
	)
	if err != nil {
		return "", err
	}

	return t.Id, nil
}

func (r *themeRepository) Update(t *model.Theme) error {
	query, err := buildSqlStatements(`
		UPDATE theme
		SET name = ?,
			config = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(
		context.TODO(),
		query,
		t.Name,
		t.Config,
		t.Id,
	)
	return err
}

func (r *themeRepository) Delete(id string) error {
	query, err := buildSqlStatements(`
		DELETE FROM theme
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, id)
	return err
}

func (r *themeRepository) UpsertInstanceBySourceFile(t *model.Theme) error {
	existing, err := r.GetBySourceFile(t.SourceFile)
	if err != nil {
		return err
	}

	if existing == nil {
		t.Scope = model.ThemeScopeInstance
		t.OwnerUserId = ""
		_, err := r.Create(t)
		return err
	}

	query, err := buildSqlStatements(`
		UPDATE theme
		SET name = ?,
			config = ?
		WHERE id = ?
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, t.Name, t.Config, existing.Id)
	return err
}

func (r *themeRepository) DeleteInstanceNotIn(keepSourceFiles []string) error {
	if len(keepSourceFiles) == 0 {
		query, err := buildSqlStatements(`DELETE FROM theme WHERE scope = ?`)
		if err != nil {
			return err
		}
		_, err = r.Engine.ExecContext(context.TODO(), query, model.ThemeScopeInstance)
		return err
	}

	placeholders := ""
	args := make([]any, 0, len(keepSourceFiles)+1)
	args = append(args, model.ThemeScopeInstance)
	for i, f := range keepSourceFiles {
		if i > 0 {
			placeholders += ", "
		}
		placeholders += "?"
		args = append(args, f)
	}

	query, err := buildSqlStatements(`
		DELETE FROM theme
		WHERE scope = ? AND (source_file IS NULL OR source_file NOT IN (` + placeholders + `))
	`)
	if err != nil {
		return err
	}

	_, err = r.Engine.ExecContext(context.TODO(), query, args...)
	return err
}

func nullableString(s string) any {
	if s == "" {
		return nil
	}
	return s
}
