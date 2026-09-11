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

// linkHostPattern matches a domain.tld-shaped host: at least two
// dot-separated labels (letters/digits/hyphens, no leading/trailing hyphen)
// with the last one (the TLD) being at least 2 letters. Deliberately not
// restricted to a fixed TLD allowlist, since custom/non-standard TLDs may be
// valid here too.
var linkHostPattern = regexp.MustCompile(`^([a-zA-Z0-9]([a-zA-Z0-9-]*[a-zA-Z0-9])?\.)+[a-zA-Z]{2,}$`)

// validateLinkURL accepts a URL with or without a scheme (defaulting to
// https), but if a scheme is present it must be http/https, and the host
// must look like a real domain.
func validateLinkURL(value string) error {
	candidate := value
	if !strings.Contains(candidate, "://") {
		candidate = "https://" + candidate
	}

	parsed, err := url.Parse(candidate)
	if err != nil {
		return fmt.Errorf("%w: %q is not a valid URL", ErrInvalidInput, value)
	}

	if parsed.Scheme != "http" && parsed.Scheme != "https" {
		return fmt.Errorf("%w: scheme must be \"http\" or \"https\", got %q", ErrInvalidInput, parsed.Scheme)
	}

	if !linkHostPattern.MatchString(parsed.Hostname()) {
		return fmt.Errorf("%w: %q is not a valid domain", ErrInvalidInput, value)
	}

	return nil
}

type LinkService interface {
	List(shelfId string) ([]model.Link, error)
	Get(linkId string) (*model.Link, error)
	Create(callerUserId string, isAdmin bool, u *model.Link) (*model.Link, error)
	Update(linkId, callerUserId string, isAdmin bool, linkRequest *model.Link) (*model.Link, error)
	Delete(linkId, callerUserId string, isAdmin bool) error
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

// List is the public, unauthenticated lookup used to render a shelf's public
// link page - it intentionally performs no ownership check.
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

	if err := validateLinkURL(u.Link); err != nil {
		return nil, err
	}

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

	if err := validateLinkURL(linkRequest.Link); err != nil {
		return nil, err
	}

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
