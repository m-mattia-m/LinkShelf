//go:generate mockgen -source=user.go -destination=mocks/user_service.go -package=mocks

package domain

import (
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type UserService interface {
	List() ([]model.User, error)
	Get(id string) (*model.User, error)
	Create(u *model.UserCreate, callerIsAdmin bool) (*model.User, error)
	Update(userId string, userRequest *model.User, callerIsAdmin bool) (*model.User, error)
	PatchPassword(userId string, u *model.UserRequestBodyOnlyPassword) error
	Delete(u *model.User) error
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

func (s *userServiceImpl) List() ([]model.User, error) {
	return s.Repository.UserRepository.List()
}

func (s *userServiceImpl) Get(id string) (*model.User, error) {
	return s.Repository.UserRepository.Get(id)
}

// Create always registers a "user" role account unless the caller is an
// authenticated admin explicitly requesting a different one - self
// registration can never grant itself elevated access.
func (s *userServiceImpl) Create(u *model.UserCreate, callerIsAdmin bool) (*model.User, error) {
	role := model.RoleUser
	if callerIsAdmin && u.Role != "" {
		validated, err := validateRole(u.Role)
		if err != nil {
			return nil, err
		}
		role = validated
	}

	hashedPassword, err := hashPassword(u.Password)
	if err != nil {
		return nil, err
	}

	userId, err := s.Repository.UserRepository.Create(u.UserBase, hashedPassword, role)
	if err != nil {
		return nil, err
	}

	user, err := s.Get(userId)
	if err != nil {
		return nil, err
	}
	return user, nil
}

// Update keeps the target's existing role unless the caller is an
// authenticated admin explicitly changing it - a self profile-update can
// never change its own role.
func (s *userServiceImpl) Update(userId string, userRequest *model.User, callerIsAdmin bool) (*model.User, error) {
	existing, err := s.Repository.UserRepository.Get(userId)
	if err != nil {
		return nil, err
	}
	if existing == nil {
		return nil, nil
	}

	role := existing.Role
	if callerIsAdmin && userRequest.Role != "" {
		validated, err := validateRole(userRequest.Role)
		if err != nil {
			return nil, err
		}
		role = validated
	}

	userRequest.Id = userId
	userRequest.Role = role
	err = s.Repository.UserRepository.Update(userRequest)
	if err != nil {
		return nil, err
	}

	user, err := s.Get(userId)
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

	return s.Repository.UserRepository.PatchPassword(userId, newHashedPassword)
}

func (s *userServiceImpl) Delete(u *model.User) error {
	return s.Repository.UserRepository.Delete(u)
}

func validateRole(role string) (string, error) {
	if role != model.RoleUser && role != model.RoleAdmin {
		return "", fmt.Errorf("%w: %q (must be %q or %q)", ErrInvalidRole, role, model.RoleUser, model.RoleAdmin)
	}
	return role, nil
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
