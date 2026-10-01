package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/knadh/koanf/parsers/yaml"
	"github.com/knadh/koanf/providers/file"
	"github.com/knadh/koanf/v2"
	"go.uber.org/zap"
)

// SyncInstanceThemes syncs instance themes with the YAML files in themes.directory.
func SyncInstanceThemes(repo *repository.Repository) error {
	dir := strings.TrimSpace(config.String("themes.directory"))
	if dir == "" {
		return nil
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			zap.L().Warn("themes.directory is set but does not exist, skipping instance theme sync", zap.String("directory", dir))
			return nil
		}
		return fmt.Errorf("read themes directory %q: %w", dir, err)
	}

	seenSourceFiles := make(map[string]bool)
	var keep []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		ext := strings.ToLower(filepath.Ext(entry.Name()))
		if ext != ".yaml" && ext != ".yml" {
			continue
		}

		sourceFile := strings.TrimSuffix(entry.Name(), filepath.Ext(entry.Name()))
		if seenSourceFiles[sourceFile] {
			zap.L().Error("duplicate instance theme slug, skipping file", zap.String("file", entry.Name()))
			continue
		}

		theme, err := loadInstanceThemeFile(filepath.Join(dir, entry.Name()), sourceFile)
		if err != nil {
			zap.L().Error("skipping invalid instance theme file", zap.String("file", entry.Name()), zap.Error(err))
			continue
		}

		if err := repo.ThemeRepository.UpsertInstanceBySourceFile(theme); err != nil {
			zap.L().Error("failed to sync instance theme", zap.String("file", entry.Name()), zap.Error(err))
			continue
		}

		seenSourceFiles[sourceFile] = true
		keep = append(keep, sourceFile)
	}

	if err := repo.ThemeRepository.DeleteInstanceNotIn(keep); err != nil {
		return fmt.Errorf("prune removed instance themes: %w", err)
	}

	zap.L().Info("Instance themes synced", zap.Int("count", len(keep)))
	return nil
}

func loadInstanceThemeFile(path, sourceFile string) (*model.Theme, error) {
	k := koanf.New(".")
	if err := k.Load(file.Provider(path), yaml.Parser()); err != nil {
		return nil, fmt.Errorf("parse yaml: %w", err)
	}

	name := strings.TrimSpace(k.String("name"))
	if name == "" {
		return nil, fmt.Errorf("missing required \"name\" field")
	}

	canonical, err := ValidateThemeConfig(k.String("config"))
	if err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &model.Theme{
		Scope:      model.ThemeScopeInstance,
		Name:       name,
		SourceFile: sourceFile,
		Config:     canonical,
	}, nil
}
