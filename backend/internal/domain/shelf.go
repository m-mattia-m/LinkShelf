package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type ShelfService interface {
	Get(id string) (*model.Shelf, error)
	Create(u *model.Shelf) (string, error)
	Update(shelfId string, shelfRequest *model.Shelf) (*model.Shelf, error)
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

func (s *shelfServiceImpl) Get(id string) (*model.Shelf, error) {
	return s.Repository.ShelfRepository.Get(id)
}

func (s *shelfServiceImpl) Create(shelfRequest *model.Shelf) (string, error) {
	return s.Repository.ShelfRepository.Create(shelfRequest)
}

func (s *shelfServiceImpl) Update(shelfId string, shelfRequest *model.Shelf) (*model.Shelf, error) {
	shelfRequest.Id = shelfId
	err := s.Repository.ShelfRepository.Update(shelfRequest)
	if err != nil {
		return nil, err
	}

	shelf, err := s.Repository.ShelfRepository.Get(shelfId)
	if err != nil {
		return nil, err
	}
	return shelf, nil
}

func (s *shelfServiceImpl) Delete(shelfRequest *model.Shelf) error {
	return s.Repository.ShelfRepository.Delete(shelfRequest)
}
