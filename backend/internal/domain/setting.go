//go:generate mockgen -source=setting.go -destination=mocks/setting_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
)

type SettingService interface {
	List() ([]model.Setting, error)
	Update(setting model.Setting) error
}

type settingServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewSettingService(repository *repository.Repository, domain *Service) SettingService {
	return &settingServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

func (s *settingServiceImpl) List() ([]model.Setting, error) {
	return s.Repository.SettingRepository.List()
}

func (s *settingServiceImpl) Update(setting model.Setting) error {
	return s.Repository.SettingRepository.Upsert(setting.Key, setting.LanguageCode, setting.Value)
}
