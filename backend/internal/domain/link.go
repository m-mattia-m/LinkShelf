//go:generate mockgen -source=link.go -destination=mocks/link_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

// linkHostPattern matches a domain.tld host (TLD of 2+ letters).
var linkHostPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// normalizeLinkURL validates a link (https by default) and returns the URL to store.
// Never store the raw input: a javascript: URL can pass as userinfo.
func normalizeLinkURL(value string) (string, error) {
	candidate := value
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	parsed, err := url.Parse(candidate)
	if err != nil {
		return "", fmt.Errorf("%w: %q is not a valid URL", ErrInvalidInput, value)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return "", fmt.Errorf("%w: scheme must be \"http\" or \"https\", got %q", ErrInvalidInput, parsed.Scheme)
	}

	// Reject userinfo; it's how a non-http payload hides before a valid host.
	if parsed.User != nil {
		return "", fmt.Errorf("%w: %q must not contain a username or password", ErrInvalidInput, value)
	}

	if !linkHostPattern.MatchString(parsed.Hostname()) {
		return "", fmt.Errorf("%w: %q is not a valid domain", ErrInvalidInput, value)
	}

	return parsed.String(), nil
}

type LinkService interface {
	List(shelfId string) ([]model.Link, error)
	Get(linkId string) (*model.Link, error)
	Create(callerUserId string, isAdmin bool, u *model.Link) (*model.Link, error)
	Update(linkId, callerUserId string, isAdmin bool, linkRequest *model.Link) (*model.Link, error)
	Delete(linkId, callerUserId string, isAdmin bool) error
	// UpdateOrder saves valid items and reports the rest as failures.
	UpdateOrder(callerUserId string, isAdmin bool, items []model.LinkOrderItem) []model.LinkOrderFailure
}

type linkServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewLinkService(repository *repository.Repository, domain *Service) LinkService {
	return &linkServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

// List is the public lookup for a shelf's links; no ownership check.
func (s *linkServiceImpl) List(shelfId string) ([]model.Link, error) {
	return s.Repository.LinkRepository.ListByShelfId(shelfId)
}

func (s *linkServiceImpl) Get(linkId string) (*model.Link, error) {
	link, err := s.Repository.LinkRepository.Get(linkId)
	if err != nil {
		return nil, err
	}
	link.Color = strings.TrimSpace(link.Color)
	return link, nil
}

// shelfOwnerOfSection resolves the user_id of the shelf that owns the given section.
func (s *linkServiceImpl) shelfOwnerOfSection(sectionId string) (*model.Shelf, error) {
	section, err := s.Repository.SectionRepository.Get(sectionId)
	if err != nil {
		return nil, err
	}
	if section == nil {
		return nil, nil
	}
	return s.Repository.ShelfRepository.Get(section.ShelfId)
}

func (s *linkServiceImpl) Create(callerUserId string, isAdmin bool, u *model.Link) (*model.Link, error) {
	shelf, err := s.shelfOwnerOfSection(u.SectionId)
	if err != nil {
		return nil, err
	}
	if shelf == nil {
		return nil, ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return nil, ErrForbidden
	}

	normalized, err := normalizeLinkURL(u.Link)
	if err != nil {
		return nil, err
	}
	u.Link = normalized

	linkId, err := s.Repository.LinkRepository.Create(u)
	if err != nil {
		return nil, err
	}

	return s.Get(linkId)
}

func (s *linkServiceImpl) Update(linkId, callerUserId string, isAdmin bool, linkRequest *model.Link) (*model.Link, error) {
	existing, err := s.Repository.LinkRepository.Get(linkId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	shelf, err := s.shelfOwnerOfSection(existing.SectionId)
	if err != nil {
		return nil, err
	}
	if shelf == nil {
		return nil, ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return nil, ErrForbidden
	}

	normalized, err := normalizeLinkURL(linkRequest.Link)
	if err != nil {
		return nil, err
	}
	linkRequest.Link = normalized

	linkRequest.Id = linkId
	err = s.Repository.LinkRepository.Update(linkRequest)
	if err != nil {
		return nil, err
	}

	return s.Get(linkId)
}

func (s *linkServiceImpl) Delete(linkId, callerUserId string, isAdmin bool) error {
	existing, err := s.Repository.LinkRepository.Get(linkId)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	shelf, err := s.shelfOwnerOfSection(existing.SectionId)
	if err != nil {
		return err
	}
	if shelf == nil {
		return ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return ErrForbidden
	}

	return s.Repository.LinkRepository.Delete(&model.Link{Id: linkId})
}

func (s *linkServiceImpl) UpdateOrder(callerUserId string, isAdmin bool, items []model.LinkOrderItem) []model.LinkOrderFailure {
	var failures []model.LinkOrderFailure

	for _, item := range items {
		if err := s.canReorder(item.Id, callerUserId, isAdmin); err != nil {
			failures = append(failures, model.LinkOrderFailure{Id: item.Id, Reason: err.Error()})
			continue
		}

		if err := s.Repository.LinkRepository.UpdateOrder(item.Id, item.Order); err != nil {
			failures = append(failures, model.LinkOrderFailure{Id: item.Id, Reason: err.Error()})
		}
	}

	return failures
}

// canReorder checks the caller may write to the link's shelf.
func (s *linkServiceImpl) canReorder(linkId, callerUserId string, isAdmin bool) error {
	existing, err := s.Repository.LinkRepository.Get(linkId)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	shelf, err := s.shelfOwnerOfSection(existing.SectionId)
	if err != nil {
		return err
	}
	if shelf == nil {
		return ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return ErrForbidden
	}

	return nil
}
