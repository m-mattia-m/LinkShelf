//go:generate mockgen -source=section.go -destination=mocks/section_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type SectionService interface {
	List(shelfId string) ([]model.Section, error)
	Get(sectionId string) (*model.Section, error)
	Create(u *model.Section) (*model.Section, error)
	Update(sectionId string, u *model.Section) (*model.Section, error)
	Delete(sectionId string) error
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

func (s *sectionServiceImpl) List(shelfId string) ([]model.Section, error) {
	return s.Repository.SectionRepository.ListByShelfId(shelfId)
}

func (s *sectionServiceImpl) Get(sectionId string) (*model.Section, error) {
	return s.Repository.SectionRepository.Get(sectionId)
}

func (s *sectionServiceImpl) Create(sectionRequest *model.Section) (*model.Section, error) {
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

func (s *sectionServiceImpl) Update(sectionId string, u *model.Section) (*model.Section, error) {
	u.Id = sectionId
	err := s.Repository.SectionRepository.Update(u)
	if err != nil {
		return nil, err
	}

	section, err := s.Repository.SectionRepository.Get(sectionId)
	if err != nil {
		return nil, err
	}
	return section, nil
}

func (s *sectionServiceImpl) Delete(sectionId string) error {
	return s.Repository.SectionRepository.Delete(&model.Section{Id: sectionId})
}
