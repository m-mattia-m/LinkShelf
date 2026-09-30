//go:generate mockgen -source=user.go -destination=mocks/user_service.go -package=mocks

package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/repository"
	"fmt"
	"strings"

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

// Create registers a "user" account unless an admin caller requests another
// role. Self-registration requires a password and
// authentication.registrationEnabled. An admin may create a passwordless
// "invited" account when email verification is enabled.
func (s *userServiceImpl) Create(u *model.UserCreate, callerIsAdmin bool) (*model.User, error) {
	if !callerIsAdmin && (!config.Bool("authentication.registrationEnabled") || !config.LocalAuthEnabled()) {
		return nil, ErrRegistrationDisabled
	}
	u.Email = normalizeEmail(u.Email)

	role := model.RoleUser
	if callerIsAdmin && u.Role != "" {
		validated, err := validateRole(u.Role)
		if err != nil {
			return nil, err
		}
		role = validated
	}

	if err := validateUsername(u.Username); err != nil {
		return nil, err
	}
	if err := checkUsernameAvailable(s.Repository, u.Username, ""); err != nil {
		return nil, err
	}

	verificationEnabled := config.Bool("authentication.emailVerification.enabled")
	passwordRequired := !callerIsAdmin || !verificationEnabled
	if passwordRequired && strings.TrimSpace(u.Password) == "" {
		return nil, fmt.Errorf("%w: password is required", ErrInvalidInput)
	}

	var hashedPassword string
	if u.Password != "" {
		var err error
		hashedPassword, err = hashPassword(u.Password)
		if err != nil {
			return nil, err
		}
	}

	userId, err := s.Repository.UserRepository.Create(u.UserBase, hashedPassword, role)
	if err != nil {
		return nil, err
	}

	user, err := s.Get(userId)
	if err != nil {
		return nil, err
	}

	if verificationEnabled {
		// Best-effort: a failed send doesn't undo the account ("resend"
		// recovers), but the caller is told so it doesn't wait for an email.
		if err := s.Domain.EmailVerificationService.SendInitial(user); err != nil {
			user.EmailDeliveryFailed = true
		}
	}

	return user, nil
}

// Update keeps the target's existing role unless the caller is an
// authenticated admin explicitly changing it - a self profile-update can
// never change its own role.
//
// A changed email is never written directly, whoever the caller is: with
// email verification on it becomes a pending change that the new address has
// to confirm (the account keeps its current email until then), and without
// it the new email is applied but marked unverified. Otherwise an account
// could claim an address it doesn't own while keeping the old address's
// verified status - which OIDC then trusts to link a first SSO login to it.
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

	// Only a changed username is checked. Renaming moves every shelf URL,
	// since they are built from the current username.
	if userRequest.Username != existing.Username {
		if err := validateUsername(userRequest.Username); err != nil {
			return nil, err
		}
		if err := checkUsernameAvailable(s.Repository, userRequest.Username, userId); err != nil {
			return nil, err
		}
	}

	newEmail := normalizeEmail(userRequest.Email)
	emailChanged := newEmail != "" && newEmail != normalizeEmail(existing.Email)
	if emailChanged {
		if err := checkEmailAvailable(s.Repository, newEmail, userId); err != nil {
			return nil, err
		}
	}

	verificationEnabled := config.Bool("authentication.emailVerification.enabled")
	deliveryFailed := false
	// Requested before the profile is saved, so a refused (rate-limited)
	// change doesn't leave the other fields half-applied.
	if emailChanged && verificationEnabled {
		delivered, err := s.Domain.EmailVerificationService.RequestEmailChange(userId, existing.Email, newEmail)
		if err != nil {
			return nil, err
		}
		deliveryFailed = !delivered
	}

	userRequest.Id = userId
	userRequest.Role = role
	userRequest.Email = existing.Email
	err = s.Repository.UserRepository.Update(userRequest)
	if err != nil {
		return nil, err
	}

	if emailChanged && !verificationEnabled {
		if err := s.Repository.UserRepository.ChangeEmail(userId, newEmail, false); err != nil {
			return nil, err
		}
	}

	user, err := s.Get(userId)
	if err != nil {
		return nil, err
	}
	if user != nil {
		user.EmailDeliveryFailed = deliveryFailed
	}
	return user, nil
}

// checkEmailAvailable fails with ErrConflict when another account (not
// exceptUserId) already uses this email. Lookups ignore case.
func checkEmailAvailable(repo *repository.Repository, email, exceptUserId string) error {
	other, err := repo.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}
	if other != nil && other.Id != exceptUserId {
		return fmt.Errorf("%w: this email address is already used by another account", ErrConflict)
	}
	return nil
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
