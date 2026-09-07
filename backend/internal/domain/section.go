//go:generate mockgen -source=section.go -destination=mocks/section_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type SectionService interface {
	List(shelfId string) ([]model.Section, error)
	Get(sectionId string) (*model.Section, error)
	Create(callerUserId string, isAdmin bool, u *model.Section) (*model.Section, error)
	Update(sectionId, callerUserId string, isAdmin bool, u *model.Section) (*model.Section, error)
	Delete(sectionId, callerUserId string, isAdmin bool) error
}

type sectionServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewSectionService(repository *repository.Repository, domain *Service) SectionService {
	return &sectionServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

// List is the public, unauthenticated lookup used to render a shelf's public
// link page - it intentionally performs no ownership check.
func (s *sectionServiceImpl) List(shelfId string) ([]model.Section, error) {
	return s.Repository.SectionRepository.ListByShelfId(shelfId)
}

func (s *sectionServiceImpl) Get(sectionId string) (*model.Section, error) {
	return s.Repository.SectionRepository.Get(sectionId)
}

// shelfOwner resolves the user_id of the shelf a section belongs (or would
// belong) to.
func (s *sectionServiceImpl) shelfOwner(shelfId string) (*model.Shelf, error) {
	return s.Repository.ShelfRepository.Get(shelfId)
}

func (s *sectionServiceImpl) Create(callerUserId string, isAdmin bool, sectionRequest *model.Section) (*model.Section, error) {
	shelf, err := s.shelfOwner(sectionRequest.ShelfId)
	if err != nil {
		return nil, err
	}
	if shelf == nil {
		return nil, ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return nil, ErrForbidden
	}

	sectionId, err := s.Repository.SectionRepository.Create(sectionRequest)
	if err != nil {
		return nil, err
	}
	section, err := s.Repository.SectionRepository.Get(sectionId)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (s *sectionServiceImpl) Update(sectionId, callerUserId string, isAdmin bool, u *model.Section) (*model.Section, error) {
	existing, err := s.Repository.SectionRepository.Get(sectionId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, ErrNotFound
	}

	shelf, err := s.shelfOwner(existing.ShelfId)
	if err != nil {
		return nil, err
	}
	if shelf == nil {
		return nil, ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return nil, ErrForbidden
	}

	u.Id = sectionId
	err = s.Repository.SectionRepository.Update(u)
	if err != nil {
		return nil, err
	}

	section, err := s.Repository.SectionRepository.Get(sectionId)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (s *sectionServiceImpl) Delete(sectionId, callerUserId string, isAdmin bool) error {
	existing, err := s.Repository.SectionRepository.Get(sectionId)
	if err != nil {
		return err
	}
	if existing == nil {
		return ErrNotFound
	}

	shelf, err := s.shelfOwner(existing.ShelfId)
	if err != nil {
		return err
	}
	if shelf == nil {
		return ErrNotFound
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return ErrForbidden
	}

	return s.Repository.SectionRepository.Delete(&model.Section{Id: sectionId})
}
