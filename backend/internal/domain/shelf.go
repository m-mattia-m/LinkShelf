//go:generate mockgen -source=shelf.go -destination=mocks/shelf_service.go -package=mocks

package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
)

type ShelfService interface {
	Get(id, callerUserId string, isAdmin bool) (*model.Shelf, error)
	GetByPath(path string) (*model.Shelf, error)
	// GetByDomain returns the shelf served on domain, or nil.
	GetByDomain(domain string) (*model.Shelf, error)
	// IsDomainRegistered reports whether any shelf is served on domain (used for CORS).
	IsDomainRegistered(domain string) (bool, error)
	GetByUsernameAndPath(username, path string) (*model.Shelf, error)
	List(callerUserId string, isAdmin bool) ([]model.Shelf, error)
	Create(callerUserId string, u *model.Shelf) (string, error)
	Update(shelfId, callerUserId string, isAdmin bool, shelfRequest *model.Shelf) (*model.Shelf, error)
	Delete(u *model.Shelf) error
	// ListDomains is one page of the shelves with a domain, for the service API. An empty
	// next cursor means the last page.
	ListDomains(cursor string, limit int) (items []model.ServiceDomain, next string, err error)
	// SetDomainStatus stores what an external controller reports about shelfId's domain.
	SetDomainStatus(domain, shelfId string, status model.DomainStatus) (*model.ServiceDomain, error)
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

// annotateThemeMissing flags shelves whose selected theme no longer exists.
func (s *shelfServiceImpl) annotateThemeMissing(shelf *model.Shelf) error {
	_, missing, err := s.Domain.ThemeService.Resolve(shelf.ThemeId)
	if err != nil {
		return err
	}
	shelf.ThemeMissing = missing
	return nil
}

// annotatePublicTheme resolves the shelf's theme into CSS properties, or nil.
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

// GetByPath is the public lookup by path. It only answers while app.userBasedPaths is off.
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

// GetByDomain is the public lookup by domain.
func (s *shelfServiceImpl) GetByDomain(domain string) (*model.Shelf, error) {
	domain = NormalizeDomain(domain)
	if ValidateDomain(domain) != nil {
		return nil, nil
	}
	// The instance's own hosts never resolve to a shelf.
	if _, reserved := reservedShelfDomains()[domain]; reserved {
		return nil, nil
	}

	var shelf *model.Shelf
	var err error
	if config.CustomDomainsRequireReady() {
		shelf, err = s.Repository.ShelfRepository.GetReadyByDomain(domain)
	} else {
		shelf, err = s.Repository.ShelfRepository.GetByDomain(domain)
	}
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
	if config.CustomDomainsRequireReady() {
		return s.Repository.ShelfRepository.DomainReadyInUse(domain, "")
	}
	return s.Repository.ShelfRepository.DomainInUse(domain, "")
}

// GetByUsernameAndPath is the public lookup while app.userBasedPaths is on.
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

// validatePath checks that a path is unique and, without user-based paths, not reserved.
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

// validateDomain checks the format, reserved hosts and uniqueness of a domain.
func (s *shelfServiceImpl) validateDomain(domain, exceptShelfId string) error {
	domain, err := checkShelfDomain(domain)
	if err != nil || domain == "" {
		return err
	}

	// With requireReady a saved domain is a claim, not yet a right to serve: the
	// uniqueness applies when a controller marks it ready (SetDomainStatus).
	if config.CustomDomainsRequireReady() {
		return nil
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

// validateLocation requires exactly one of path and domain; unchanged values aren't re-checked.
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

// Create assigns the caller as owner, returning 404 if the caller no longer exists.
func (s *shelfServiceImpl) Create(callerUserId string, shelfRequest *model.Shelf) (string, error) {
	user, err := s.Repository.UserRepository.Get(callerUserId)
	if err != nil {
		return "", err
	}
	if user == nil {
		return "", ErrNotFound
	}

	if user.MaxShelves != nil {
		count, err := s.Repository.ShelfRepository.CountByUserId(callerUserId)
		if err != nil {
			return "", err
		}
		if count >= *user.MaxShelves {
			return "", shelfLimitError(*user.MaxShelves)
		}
	}

	if err := s.Domain.ThemeService.ValidateAssignable(shelfRequest.ThemeId, callerUserId); err != nil {
		return "", err
	}

	// Should not happen: every user has a username.
	if config.Bool("app.userBasedPaths") && shelfRequest.Path != "" && user.Username == "" {
		return "", fmt.Errorf("%w: set a username before creating a shelf with a path", ErrInvalidInput)
	}
	if err := s.validateLocation(shelfRequest, nil, callerUserId); err != nil {
		return "", err
	}

	shelfRequest.UserId = callerUserId
	// Stamped here, never taken from the client, and left alone by Update.
	shelfRequest.CreatedWithUserBasedPaths = config.Bool("app.userBasedPaths")
	id, err := s.Repository.ShelfRepository.Create(shelfRequest)
	if err != nil {
		return "", err
	}
	s.Domain.publishDomainChanged(id, callerUserId, shelfRequest.Domain, "")
	return id, nil
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

	s.Domain.publishDomainChanged(shelfId, existing.UserId, shelfRequest.Domain, existing.Domain)

	shelf, err := s.Repository.ShelfRepository.Get(shelfId)
	if err != nil {
		return nil, err
	}
	if err := s.annotateThemeMissing(shelf); err != nil {
		return nil, err
	}
	return shelf, nil
}

// Delete expects the caller to have checked ownership via Get.
func (s *shelfServiceImpl) Delete(shelfRequest *model.Shelf) error {
	if err := s.Repository.ShelfRepository.Delete(shelfRequest); err != nil {
		return err
	}
	s.Domain.publishDomainChanged(shelfRequest.Id, shelfRequest.UserId, "", shelfRequest.Domain)
	return nil
}

// ListDomains reads one page after cursor, a shelf id.
func (s *shelfServiceImpl) ListDomains(cursor string, limit int) ([]model.ServiceDomain, string, error) {
	shelves, err := s.Repository.ShelfRepository.ListDomains(cursor, limit+1)
	if err != nil {
		return nil, "", err
	}

	next := ""
	if len(shelves) > limit {
		shelves = shelves[:limit]
		next = shelves[limit-1].Id
	}

	items := make([]model.ServiceDomain, 0, len(shelves))
	for _, shelf := range shelves {
		items = append(items, toServiceDomain(shelf))
	}
	return items, next, nil
}

// maxDomainRecordsBytes is what fits into shelf.domain_records.
const maxDomainRecordsBytes = 4096

// SetDomainStatus answers ErrConflict when shelfId does not have domain (anymore), so
// a stale controller can not label a new domain, or when another shelf is ready on it.
func (s *shelfServiceImpl) SetDomainStatus(domain, shelfId string, status model.DomainStatus) (*model.ServiceDomain, error) {
	switch status.State {
	case model.DomainStatePending, model.DomainStateReady, model.DomainStateError:
	default:
		return nil, fmt.Errorf("%w: state must be pending, ready or error", ErrInvalidInput)
	}
	if status.Records == nil {
		status.Records = []model.DomainRecord{}
	}
	encoded, err := json.Marshal(status.Records)
	if err != nil {
		return nil, err
	}
	if len(encoded) > maxDomainRecordsBytes {
		return nil, fmt.Errorf("%w: the records are larger than %d bytes", ErrInvalidInput, maxDomainRecordsBytes)
	}

	domain = NormalizeDomain(domain)
	shelf, err := s.Repository.ShelfRepository.Get(shelfId)
	if err != nil {
		return nil, err
	}
	if shelf == nil || domain == "" || shelf.Domain != domain {
		return nil, fmt.Errorf("%w: the shelf does not have the domain %q anymore", ErrConflict, domain)
	}

	if status.State == model.DomainStateReady {
		taken, err := s.Repository.ShelfRepository.DomainReadyInUse(domain, shelfId)
		if err != nil {
			return nil, err
		}
		if taken {
			return nil, fmt.Errorf("%w: another shelf is already ready on the domain %q", ErrConflict, domain)
		}
	}

	now := time.Now().Unix()
	updated, err := s.Repository.ShelfRepository.SetDomainStatus(shelfId, domain, status, now)
	if err != nil {
		return nil, err
	}
	if !updated {
		return nil, fmt.Errorf("%w: the shelf does not have the domain %q anymore", ErrConflict, domain)
	}

	shelf.DomainStatus, shelf.DomainUpdatedAt = &status, now
	item := toServiceDomain(*shelf)
	return &item, nil
}

func toServiceDomain(shelf model.Shelf) model.ServiceDomain {
	item := model.ServiceDomain{
		Domain:    shelf.Domain,
		ShelfId:   shelf.Id,
		OwnerId:   shelf.UserId,
		CreatedAt: unixToRfc3339(shelfCreatedAt(shelf.Id)),
		UpdatedAt: unixToRfc3339(max(shelf.DomainUpdatedAt, shelfCreatedAt(shelf.Id))),
		Status:    model.ServiceDomainStatus{Records: []model.DomainRecord{}, ObservedDomain: shelf.Domain},
	}
	if shelf.DomainStatus != nil {
		item.Status.State = shelf.DomainStatus.State
		item.Status.Reason = shelf.DomainStatus.Reason
		item.Status.Message = shelf.DomainStatus.Message
		if shelf.DomainStatus.Records != nil {
			item.Status.Records = shelf.DomainStatus.Records
		}
	}
	return item
}

// shelfCreatedAt reads the creation time out of the UUIDv7 id, 0 for any other id.
func shelfCreatedAt(id string) int64 {
	parsed, err := uuid.Parse(id)
	if err != nil || parsed.Version() != 7 {
		return 0
	}
	seconds, _ := parsed.Time().UnixTime()
	return seconds
}

func shelfLimitError(limit int) error {
	if u := strings.TrimSpace(config.String("limits.upgradeUrl")); u != "" {
		return fmt.Errorf("%w: your account is limited to %d shelves, upgrade at %s", ErrShelfLimitReached, limit, u)
	}
	return fmt.Errorf("%w: your account is limited to %d shelves", ErrShelfLimitReached, limit)
}
