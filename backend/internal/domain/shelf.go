package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type ShelfService interface {
	GetShelfById(id string) (*model.Shelf, error)
	CreateShelf(u *model.Shelf) (string, error)
	UpdateShelf(shelfId string, shelfRequest *model.Shelf) (*model.Shelf, error)
	DeleteShelf(u *model.Shelf) error
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

func (s *shelfServiceImpl) GetShelfById(id string) (*model.Shelf, error) {
	return s.Repository.ShelfRepository.Get(id)
}

func (s *shelfServiceImpl) CreateShelf(shelfRequest *model.Shelf) (string, error) {
	return s.Repository.ShelfRepository.Create(shelfRequest)
}

func (s *shelfServiceImpl) UpdateShelf(shelfId string, shelfRequest *model.Shelf) (*model.Shelf, error) {
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

func (s *shelfServiceImpl) DeleteShelf(shelfRequest *model.Shelf) error {
	return s.Repository.ShelfRepository.Delete(shelfRequest)
}
