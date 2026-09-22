//go:generate mockgen -source=shelf.go -destination=mocks/shelf_service.go -package=mocks

package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
	"strings"
)

type ShelfService interface {
	Get(id, callerUserId string, isAdmin bool) (*model.Shelf, error)
	GetByPath(path string) (*model.Shelf, error)
	// GetByDomain is the public lookup for a shelf served on its own domain.
	// It returns nil for a domain no shelf has, and for a value that can't be
	// a domain at all.
	GetByDomain(domain string) (*model.Shelf, error)
	// IsDomainRegistered reports whether any shelf is served on domain, which
	// is what decides if a browser origin on that domain may call the API
	// (see app.strictOrigins).
	IsDomainRegistered(domain string) (bool, error)
	GetByUsernameAndPath(username, path string) (*model.Shelf, error)
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
// public link page - it intentionally performs no ownership check. It only
// answers while app.userBasedPaths is off: with it on, /<path> without a
// username is no longer a valid URL and must not resolve.
func (s *shelfServiceImpl) GetByPath(path string) (*model.Shelf, error) {
	if config.Bool("app.userBasedPaths") {
		return nil, nil
	}

	shelf, err := s.Repository.ShelfRepository.GetByPath(path)
	if err != nil || shelf == nil {
		return shelf, err
	}
	if err := s.annotatePublicTheme(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

// GetByDomain is GetByPath for a shelf that is served on a domain of its own.
// Unlike a path lookup it doesn't depend on app.userBasedPaths, since a domain
// is unique on the whole instance either way.
func (s *shelfServiceImpl) GetByDomain(domain string) (*model.Shelf, error) {
	domain = NormalizeDomain(domain)
	if ValidateDomain(domain) != nil {
		return nil, nil
	}

	shelf, err := s.Repository.ShelfRepository.GetByDomain(domain)
	if err != nil || shelf == nil {
		return shelf, err
	}
	if err := s.annotatePublicTheme(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

func (s *shelfServiceImpl) IsDomainRegistered(domain string) (bool, error) {
	domain = NormalizeDomain(domain)
	if ValidateDomain(domain) != nil {
		return false, nil
	}
	return s.Repository.ShelfRepository.DomainInUse(domain, "")
}

// GetByUsernameAndPath is GetByPath for /<username>/<path>, and only answers
// while app.userBasedPaths is on.
func (s *shelfServiceImpl) GetByUsernameAndPath(username, path string) (*model.Shelf, error) {
	if !config.Bool("app.userBasedPaths") {
		return nil, nil
	}

	shelf, err := s.Repository.ShelfRepository.GetByUsernameAndPath(strings.ToLower(username), path)
	if err != nil || shelf == nil {
		return shelf, err
	}
	if err := s.annotatePublicTheme(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

// validatePath checks a shelf's path before it is saved. Which rules apply
// depends on app.userBasedPaths:
//   - on: the path lives behind the owner's username, so it only has to be
//     unique among that owner's shelves and no word is off limits.
//   - off: the path is a top-level URL, so it must be unique across the whole
//     instance and must not be a word the frontend or backend already routes.
//
// An empty path (a shelf that is only reachable through its domain) is always
// fine.
func (s *shelfServiceImpl) validatePath(path, ownerId, exceptShelfId string) error {
	if path == "" {
		return nil
	}

	if config.Bool("app.userBasedPaths") {
		taken, err := s.Repository.ShelfRepository.PathInUseByUser(ownerId, path, exceptShelfId)
		if err != nil {
			return err
		}
		if taken {
			return fmt.Errorf("%w: you already have a shelf with the path %q", ErrConflict, path)
		}
		return nil
	}

	if isRouteReserved(path) {
		return fmt.Errorf("%w: the path %q is reserved and can't be used for a shelf", ErrInvalidInput, path)
	}
	taken, err := s.Repository.ShelfRepository.PathInUse(path, exceptShelfId)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("%w: the path %q is already in use", ErrConflict, path)
	}
	return nil
}

// validateDomain checks a shelf's domain before it is saved: the format, that
// it isn't one of this instance's own hosts, and that no other shelf has it.
// An empty domain (a shelf that is only reachable through its path) is always
// fine.
func (s *shelfServiceImpl) validateDomain(domain, exceptShelfId string) error {
	domain, err := checkShelfDomain(domain)
	if err != nil || domain == "" {
		return err
	}

	taken, err := s.Repository.ShelfRepository.DomainInUse(domain, exceptShelfId)
	if err != nil {
		return err
	}
	if taken {
		return fmt.Errorf("%w: the domain %q is already in use", ErrConflict, domain)
	}
	return nil
}

// validateLocation checks how a shelf is reached before it is saved: through
// exactly one of a path or a domain, never both and never neither. It also
// leaves request.Domain in its normalized form, which is what gets stored.
//
// existing is the shelf as it is now, or nil when creating one. Only a
// changed path or domain is checked, so a shelf that already sits on a value
// that is no longer allowed (a path a later release reserved, or a shelf
// created before a shelf had to choose one of the two) can still have its
// title or theme edited.
func (s *shelfServiceImpl) validateLocation(request, existing *model.Shelf, ownerId string) error {
	request.Domain = NormalizeDomain(request.Domain)

	exceptShelfId := ""
	pathChanged, domainChanged := true, true
	if existing != nil {
		exceptShelfId = existing.Id
		pathChanged = request.Path != existing.Path
		domainChanged = request.Domain != existing.Domain
	}
	if !pathChanged && !domainChanged {
		return nil
	}

	switch {
	case request.Path != "" && request.Domain != "":
		return fmt.Errorf("%w: a shelf has either a path or a domain, not both", ErrInvalidInput)
	case request.Path == "" && request.Domain == "":
		return fmt.Errorf("%w: a shelf needs a path or a domain", ErrInvalidInput)
	}

	if pathChanged {
		if err := s.validatePath(request.Path, ownerId, exceptShelfId); err != nil {
			return err
		}
	}
	if domainChanged {
		if err := s.validateDomain(request.Domain, exceptShelfId); err != nil {
			return err
		}
	}
	return nil
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

	// A username is normally always set (every creation path requires one and
	// startup backfills the rest), so this only guards a state that should
	// not exist: a URL with the username missing.
	if config.Bool("app.userBasedPaths") && shelfRequest.Path != "" && user.Username == "" {
		return "", fmt.Errorf("%w: set a username before creating a shelf with a path", ErrInvalidInput)
	}
	if err := s.validateLocation(shelfRequest, nil, callerUserId); err != nil {
		return "", err
	}

	shelfRequest.UserId = callerUserId
	// Stamped here, never taken from the client, and left alone by Update.
	shelfRequest.CreatedWithUserBasedPaths = config.Bool("app.userBasedPaths")
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

	if err := s.validateLocation(shelfRequest, existing, existing.UserId); err != nil {
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
