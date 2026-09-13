//go:generate mockgen -source=theme.go -destination=mocks/theme_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
)

type ThemeService interface {
	// ListGrouped returns instance-provided themes alongside the caller's own
	// - the grouping the shelf editor's theme picker and the "Themes" nav
	// page both render.
	ListGrouped(callerUserId string) (model.ThemeGroupedResponseBody, error)
	Get(themeId, callerUserId string, isAdmin bool) (*model.Theme, error)
	Create(callerUserId string, base model.ThemeBase) (*model.Theme, error)
	// Update only ever succeeds for the theme's own owner - not even an
	// admin may edit another user's theme content (they may only delete it,
	// see Delete), and instance-scoped themes can never be edited through
	// the API at all since they're owned by the mounted config directory.
	Update(themeId, callerUserId string, base model.ThemeBase) (*model.Theme, error)
	// Delete allows the owner or an admin to remove a user-scoped theme.
	// Instance-scoped themes can never be deleted through the API - removing
	// one means deleting its file from the mounted directory and restarting.
	Delete(themeId, callerUserId string, isAdmin bool) error
	// ListAllUserScoped is the admin moderation view across every user's
	// themes. Route-level Metadata gates this to admins.
	ListAllUserScoped() ([]model.Theme, error)
	// Resolve returns a theme's validated property map for rendering. An
	// empty themeId resolves to (nil, false, nil) - no theme selected, render
	// the built-in default look. A themeId that no longer matches any theme
	// resolves to (nil, true, nil) - missing, same default look, but the
	// caller (the shelf edit page) should tell the owner to pick a new one.
	Resolve(themeId string) (config map[string]string, missing bool, err error)
	// ValidateAssignable checks that shelfOwnerUserId may set a shelf's theme
	// to themeId: unset, any instance theme, or a user theme owned by
	// shelfOwnerUserId. Prevents assigning someone else's private theme.
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
		// Defensive only: config is validated before it's ever stored, so
		// this can't happen in practice. Treat it the same as "gone" rather
		// than surfacing a 500 to a page visitor.
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
