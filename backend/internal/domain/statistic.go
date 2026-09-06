//go:generate mockgen -source=statistic.go -destination=mocks/statistic_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type StatisticService interface {
	Get(userId string) (*model.Statistic, error)
}

type statisticServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewStatisticService(repository *repository.Repository, domain *Service) StatisticService {
	return &statisticServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

func (s *statisticServiceImpl) Get(userId string) (*model.Statistic, error) {
	shelfNumber, err := s.Repository.StatisticRepository.GetShelfAmount(userId)
	if err != nil {
		return nil, err
	}

	sectionNumber, err := s.Repository.StatisticRepository.GetSectionAmount(userId)
	if err != nil {
		return nil, err
	}

	linkNumber, err := s.Repository.StatisticRepository.GetLinkAmount(userId)
	if err != nil {
		return nil, err
	}

	return &model.Statistic{
		ShelfNumber:   *shelfNumber,
		SectionNumber: *sectionNumber,
		LinkNumber:    *linkNumber,
	}, nil
}
