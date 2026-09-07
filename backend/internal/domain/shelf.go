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

// Get returns a shelf, enforcing that only its owner or an admin may see it.
func (s *shelfServiceImpl) Get(id, callerUserId string, isAdmin bool) (*model.Shelf, error) {
	shelf, err := s.Repository.ShelfRepository.Get(id)
	if err != nil || shelf == nil {
		return shelf, err
	}
	if !isAdmin && shelf.UserId != callerUserId {
		return nil, ErrForbidden
	}
	return shelf, nil
}

// GetByPath is the public, unauthenticated lookup used to render a shelf's
// public link page - it intentionally performs no ownership check.
func (s *shelfServiceImpl) GetByPath(path string) (*model.Shelf, error) {
	return s.Repository.ShelfRepository.GetByPath(path)
}

// List returns every shelf for an admin, or only the caller's own shelves otherwise.
func (s *shelfServiceImpl) List(callerUserId string, isAdmin bool) ([]model.Shelf, error) {
	if isAdmin {
		return s.Repository.ShelfRepository.List()
	}
	return s.Repository.ShelfRepository.ListByUserId(callerUserId)
}

// Create always assigns ownership to the caller - a client can never create a
// shelf on someone else's behalf.
func (s *shelfServiceImpl) Create(callerUserId string, shelfRequest *model.Shelf) (string, error) {
	shelfRequest.UserId = callerUserId
	return s.Repository.ShelfRepository.Create(shelfRequest)
}

func (s *shelfServiceImpl) Update(shelfId, callerUserId string, isAdmin bool, shelfRequest *model.Shelf) (*model.Shelf, error) {
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

	shelfRequest.Id = shelfId
	err = s.Repository.ShelfRepository.Update(shelfRequest)
	if err != nil {
		return nil, err
	}

	shelf, err := s.Repository.ShelfRepository.Get(shelfId)
	if err != nil {
		return nil, err
	}
	return shelf, nil
}

// Delete relies on the caller having already fetched the shelf through Get,
// which enforces ownership.
func (s *shelfServiceImpl) Delete(shelfRequest *model.Shelf) error {
	return s.Repository.ShelfRepository.Delete(shelfRequest)
}
