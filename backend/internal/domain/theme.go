//go:generate mockgen -source=theme.go -destination=mocks/theme_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
)

type ThemeService interface {
	// ListGrouped returns instance themes and the caller's own.
	ListGrouped(callerUserId string) (model.ThemeGroupedResponseBody, error)
	Get(themeId, callerUserId string, isAdmin bool) (*model.Theme, error)
	Create(callerUserId string, base model.ThemeBase) (*model.Theme, error)
	// Update is owner-only; instance themes can't be edited via the API.
	Update(themeId, callerUserId string, base model.ThemeBase) (*model.Theme, error)
	// Delete allows the owner or an admin; instance themes can't be deleted via the API.
	Delete(themeId, callerUserId string, isAdmin bool) error
	// ListAllUserScoped lists every user's themes (admin-only).
	ListAllUserScoped() ([]model.Theme, error)
	// Resolve returns a theme's properties and whether a set themeId is missing.
	Resolve(themeId string) (config map[string]string, missing bool, err error)
	// ValidateAssignable checks the theme is an instance theme or owned by shelfOwnerUserId.
	ValidateAssignable(themeId, shelfOwnerUserId string) error
}

type themeServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewThemeService(repository *repository.Repository, domain *Service) ThemeService {
	return &themeServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

func (s *themeServiceImpl) ListGrouped(callerUserId string) (model.ThemeGroupedResponseBody, error) {
	instance, err := s.Repository.ThemeRepository.ListInstance()
	if err != nil {
		return model.ThemeGroupedResponseBody{}, err
	}

	mine, err := s.Repository.ThemeRepository.ListByOwner(callerUserId)
	if err != nil {
		return model.ThemeGroupedResponseBody{}, err
	}

	return model.ThemeGroupedResponseBody{Instance: instance, Mine: mine}, nil
}

func (s *themeServiceImpl) Get(themeId, callerUserId string, isAdmin bool) (*model.Theme, error) {
	theme, err := s.Repository.ThemeRepository.Get(themeId)
	if err != nil {
		return nil, err
	}
	if theme == nil {
		return nil, ErrNotFound
	}
	if theme.Scope == model.ThemeScopeUser && !isAdmin && theme.OwnerUserId != callerUserId {
		return nil, ErrForbidden
	}
	return theme, nil
}

func (s *themeServiceImpl) Create(callerUserId string, base model.ThemeBase) (*model.Theme, error) {
	canonical, err := ValidateThemeConfig(base.Config)
	if err != nil {
		return nil, err
	}

	theme := &model.Theme{
		Scope:       model.ThemeScopeUser,
		OwnerUserId: callerUserId,
		Name:        base.Name,
		Config:      canonical,
	}

	themeId, err := s.Repository.ThemeRepository.Create(theme)
	if err != nil {
		return nil, err
	}

	return s.Repository.ThemeRepository.Get(themeId)
}

func (s *themeServiceImpl) Update(themeId, callerUserId string, base model.ThemeBase) (*model.Theme, error) {
	existing, err := s.Repository.ThemeRepository.Get(themeId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}
	if existing.Scope != model.ThemeScopeUser || existing.OwnerUserId != callerUserId {
		return nil, ErrForbidden
	}

	canonical, err := ValidateThemeConfig(base.Config)
	if err != nil {
		return nil, err
	}

	existing.Name = base.Name
	existing.Config = canonical
	if err := s.Repository.ThemeRepository.Update(existing); err != nil {
		return nil, err
	}

	return s.Repository.ThemeRepository.Get(themeId)
}

func (s *themeServiceImpl) Delete(themeId, callerUserId string, isAdmin bool) error {
	existing, err := s.Repository.ThemeRepository.Get(themeId)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}
	if existing.Scope != model.ThemeScopeUser {
		return fmt.Errorf("%w: instance themes are managed via the mounted themes directory, not the API", ErrForbidden)
	}
	if !isAdmin && existing.OwnerUserId != callerUserId {
		return ErrForbidden
	}

	return s.Repository.ThemeRepository.Delete(themeId)
}

func (s *themeServiceImpl) ListAllUserScoped() ([]model.Theme, error) {
	return s.Repository.ThemeRepository.ListAllUserScoped()
}

func (s *themeServiceImpl) Resolve(themeId string) (map[string]string, bool, error) {
	if themeId == "" {
		return nil, false, nil
	}

	theme, err := s.Repository.ThemeRepository.Get(themeId)
	if err != nil {
		return nil, false, err
	}
	if theme == nil {
		return nil, true, nil
	}

	config, err := ParseThemeConfig(theme.Config)
	if err != nil {
		// Stored config is always valid; treat a failure as a missing theme.
		return nil, true, nil
	}

	return config, false, nil
}

func (s *themeServiceImpl) ValidateAssignable(themeId, shelfOwnerUserId string) error {
	if themeId == "" {
		return nil
	}

	theme, err := s.Repository.ThemeRepository.Get(themeId)
	if err != nil {
		return err
	}
	if theme == nil {
		return fmt.Errorf("%w: theme not found", ErrInvalidInput)
	}
	if theme.Scope == model.ThemeScopeInstance {
		return nil
	}
	if theme.OwnerUserId != shelfOwnerUserId {
		return fmt.Errorf("%w: cannot use another user's theme", ErrForbidden)
	}

	return nil
}
