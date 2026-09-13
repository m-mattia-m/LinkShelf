//go:generate mockgen -source=shelf.go -destination=mocks/shelf_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type ShelfService interface {
	Get(id, callerUserId string, isAdmin bool) (*model.Shelf, error)
	GetByPath(path string) (*model.Shelf, error)
	List(callerUserId string, isAdmin bool) ([]model.Shelf, error)
	Create(callerUserId string, u *model.Shelf) (string, error)
	Update(shelfId, callerUserId string, isAdmin bool, shelfRequest *model.Shelf) (*model.Shelf, error)
	Delete(u *model.Shelf) error
}

type shelfServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewShelfService(repository *repository.Repository, domain *Service) ShelfService {
	return &shelfServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

// annotateThemeMissing sets ThemeMissing so the edit page can tell "never
// picked a theme" apart from "the one picked is gone" - it deliberately
// leaves the resolved Theme config map unset here, since the authenticated
// edit view doesn't render with it (no live preview - see design notes),
// only the public view (annotatePublicTheme) does.
func (s *shelfServiceImpl) annotateThemeMissing(shelf *model.Shelf) error {
	_, missing, err := s.Domain.ThemeService.Resolve(shelf.ThemeId)
	if err != nil {
		return err
	}
	shelf.ThemeMissing = missing
	return nil
}

// annotatePublicTheme resolves the shelf's theme into its rendered CSS
// property map for the public link page - nil (silently falling back to the
// page's own default look) if none is selected or the selected one no
// longer exists.
func (s *shelfServiceImpl) annotatePublicTheme(shelf *model.Shelf) error {
	config, _, err := s.Domain.ThemeService.Resolve(shelf.ThemeId)
	if err != nil {
		return err
	}
	shelf.Theme = config
	return nil
}

// Get returns a shelf, enforcing that only its owner or an admin may see it.
func (s *shelfServiceImpl) Get(id, callerUserId string, isAdmin bool) (*model.Shelf, error) {
	shelf, err := s.Repository.ShelfRepository.Get(id)
	if err != nil || shelf == nil {
		return shelf, err
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return nil, ErrForbidden
	}
	if err := s.annotateThemeMissing(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

// GetByPath is the public, unauthenticated lookup used to render a shelf's
// public link page - it intentionally performs no ownership check.
func (s *shelfServiceImpl) GetByPath(path string) (*model.Shelf, error) {
	shelf, err := s.Repository.ShelfRepository.GetByPath(path)
	if err != nil || shelf == nil {
		return shelf, err
	}
	if err := s.annotatePublicTheme(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

// List returns every shelf for an admin, or only the caller's own shelves otherwise.
func (s *shelfServiceImpl) List(callerUserId string, isAdmin bool) ([]model.Shelf, error) {
	if isAdmin {
		return s.Repository.ShelfRepository.List()
	}
	return s.Repository.ShelfRepository.ListByUserId(callerUserId)
}

// Create always assigns ownership to the caller - a client can never create a
// shelf on someone else's behalf. It first re-verifies the caller's user
// still exists, guarding the rare case of a still-valid JWT for a since-
// deleted user (the DB's user_id foreign key would otherwise surface as a
// raw constraint-violation error instead of a clean 404).
func (s *shelfServiceImpl) Create(callerUserId string, shelfRequest *model.Shelf) (string, error) {
	user, err := s.Repository.UserRepository.Get(callerUserId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrNotFound
	}

	if err := s.Domain.ThemeService.ValidateAssignable(shelfRequest.ThemeId, callerUserId); err != nil {
		return "", err
	}

	shelfRequest.UserId = callerUserId
	return s.Repository.ShelfRepository.Create(shelfRequest)
}

func (s *shelfServiceImpl) Update(shelfId, callerUserId string, isAdmin bool, shelfRequest *model.Shelf) (*model.Shelf, error) {
	user, err := s.Repository.UserRepository.Get(callerUserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrNotFound
	}

	existing, err := s.Repository.ShelfRepository.Get(shelfId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}
	if !isAdmin && existing.UserId != callerUserId {
		return nil, ErrForbidden
	}

	if err := s.Domain.ThemeService.ValidateAssignable(shelfRequest.ThemeId, existing.UserId); err != nil {
		return nil, err
	}

	shelfRequest.Id = shelfId
	err = s.Repository.ShelfRepository.Update(shelfRequest)
	if err != nil {
		return nil, err
	}

	shelf, err := s.Repository.ShelfRepository.Get(shelfId)
	if err != nil {
		return nil, err
	}
	if err := s.annotateThemeMissing(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

// Delete relies on the caller having already fetched the shelf through Get,
// which enforces ownership.
func (s *shelfServiceImpl) Delete(shelfRequest *model.Shelf) error {
	return s.Repository.ShelfRepository.Delete(shelfRequest)
}
