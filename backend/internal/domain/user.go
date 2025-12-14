package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	GetUserById(id string) (*model.User, error)
	CreateUser(u *model.User) (*model.User, error)
	UpdateUser(userId string, userRequest *model.User) (*model.User, error)
	PatchPassword(userId string, u *model.UserRequestBodyOnlyPassword) error
	DeleteUser(u *model.User) error
}

type userServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
}

func NewUserService(repository *repository.Repository, domain *Service) UserService {
	return &userServiceImpl{
		Repository: repository,
		Domain:     domain,
	}
}

func (s *userServiceImpl) GetUserById(id string) (*model.User, error) {
	return s.Repository.UserRepository.Get(id)
}

func (s *userServiceImpl) CreateUser(u *model.User) (*model.User, error) {

	var err error
	u.Password, err = hashPassword(u.Password)
	if err != nil {
		return nil, err
	}

	userId, err := s.Repository.UserRepository.Create(u)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUserById(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userServiceImpl) UpdateUser(userId string, userRequest *model.User) (*model.User, error) {
	userRequest.Id = userId
	err := s.Repository.UserRepository.Update(userRequest)
	if err != nil {
		return nil, err
	}

	user, err := s.GetUserById(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (s *userServiceImpl) PatchPassword(userId string, u *model.UserRequestBodyOnlyPassword) error {

	safedPasswordHash, err := s.Repository.UserRepository.GetPassword(userId)
	if err != nil {
		return err
	}

	err = checkPassword(safedPasswordHash, u.OldPassword)
	if err != nil {
		return err
	}

	newHashedPassword, err := hashPassword(u.NewPassword)
	if err != nil {
		return err
	}

	return s.Repository.UserRepository.PatchPassword(&model.User{
		Id: userId,
		UserBase: model.UserBase{
			Password: newHashedPassword,
		},
	})
}

func (s *userServiceImpl) DeleteUser(u *model.User) error {
	return s.Repository.UserRepository.Delete(u)
}

// hashPassword hashes a plaintext password using bcrypt.
func hashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hashedBytes), nil
}

// checkPassword compares a bcrypt hashed password with its plaintext version.
func checkPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword(
		[]byte(hashedPassword),
		[]byte(password),
	)
}
