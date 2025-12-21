package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"strings"
)

type LinkService interface {
	List(shelfId string) ([]model.Link, error)
	Get(linkId string) (*model.Link, error)
	Create(u *model.Link) (*model.Link, error)
	Update(linkId string, linkRequest *model.Link) (*model.Link, error)
	Delete(linkId string) error
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

func (s *linkServiceImpl) Create(u *model.Link) (*model.Link, error) {
	linkId, err := s.Repository.LinkRepository.Create(u)
	if err != nil {
		return nil, err
	}

	link, err := s.Domain.LinkService.Get(linkId)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (s *linkServiceImpl) Update(linkId string, linkRequest *model.Link) (*model.Link, error) {
	linkRequest.Id = linkId
	err := s.Repository.LinkRepository.Update(linkRequest)
	if err != nil {
		return nil, err
	}

	links, err := s.Repository.LinkRepository.Get(linkId)
	if err != nil {
		return nil, err
	}

	return links, nil
}

func (s *linkServiceImpl) Delete(linkId string) error {
	return s.Repository.LinkRepository.Delete(&model.Link{Id: linkId})
}
