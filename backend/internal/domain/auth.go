//go:generate mockgen -source=auth.go -destination=mocks/auth_service.go -package=mocks

package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/oidcclient"
	"backend/internal/infrastructure/repository"
	"context"
	"errors"
	"time"
)

var (
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrEmailNotVerified   = errors.New("the identity provider did not confirm this email address is verified")
	ErrOidcNotConfigured  = errors.New("OIDC login is not enabled")
)

type AuthService interface {
	Login(email, password string) (*model.TokenPair, error)
	Refresh(rawRefreshToken string) (*model.TokenPair, error)
	Logout(rawRefreshToken string) error
	OidcAuthorizationURL() (*model.OidcLoginResponseBody, error)
	// OidcCallback completes an OIDC login. currentUserId, when non-nil, means
	// the caller was already authenticated and this is a link-to-my-account
	// request rather than a login/auto-provision one.
	OidcCallback(ctx context.Context, code, state string, currentUserId *string) (*model.TokenPair, error)
}

type authServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
	oidc       *oidcclient.Client
}

func NewAuthService(repository *repository.Repository, domain *Service, oidc *oidcclient.Client) AuthService {
	return &authServiceImpl{
		Repository: repository,
		Domain:     domain,
		oidc:       oidc,
	}
}

func (s *authServiceImpl) Login(email, password string) (*model.TokenPair, error) {
	record, err := s.Repository.UserRepository.FindByEmail(email)
	if err != nil {
		return nil, err
	}
	if record == nil || record.Password == "" {
		return nil, ErrInvalidCredentials
	}
	if err := checkPassword(record.Password, password); err != nil {
		return nil, ErrInvalidCredentials
	}

	return s.issueTokenPair(record.Id, record.Role)
}

func (s *authServiceImpl) Refresh(rawRefreshToken string) (*model.TokenPair, error) {
	hash := hashRefreshToken(rawRefreshToken)

	stored, err := s.Repository.RefreshTokenRepository.GetByHash(hash)
	if err != nil {
		return nil, err
	}
	if stored == nil || stored.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	// Single-use rotation: invalidate the token being used right away.
	if err := s.Repository.RefreshTokenRepository.DeleteByHash(hash); err != nil {
		return nil, err
	}

	user, err := s.Repository.UserRepository.Get(stored.UserId)
	if err != nil {
		return nil, err
	}
	if user == nil {
		return nil, ErrInvalidToken
	}

	return s.issueTokenPair(user.Id, user.Role)
}

func (s *authServiceImpl) Logout(rawRefreshToken string) error {
	return s.Repository.RefreshTokenRepository.DeleteByHash(hashRefreshToken(rawRefreshToken))
}

func (s *authServiceImpl) OidcAuthorizationURL() (*model.OidcLoginResponseBody, error) {
	if s.oidc == nil {
		return nil, ErrOidcNotConfigured
	}

	authURL, state, err := s.oidc.AuthorizationURL()
	if err != nil {
		return nil, err
	}

	return &model.OidcLoginResponseBody{AuthorizationUrl: authURL, State: state}, nil
}

func (s *authServiceImpl) OidcCallback(ctx context.Context, code, state string, currentUserId *string) (*model.TokenPair, error) {
	if s.oidc == nil {
		return nil, ErrOidcNotConfigured
	}

	identity, err := s.oidc.Exchange(ctx, code, state)
	if err != nil {
		return nil, err
	}

	// Link mode: attach this external identity to the already-authenticated user.
	if currentUserId != nil {
		if err := s.Repository.UserRepository.LinkProvider(*currentUserId, model.ProviderOIDC, identity.Subject); err != nil {
			return nil, err
		}
		user, err := s.Repository.UserRepository.Get(*currentUserId)
		if err != nil {
			return nil, err
		}
		return s.issueTokenPair(user.Id, user.Role)
	}

	// Returning user: already linked, matched by the stable provider subject.
	record, err := s.Repository.UserRepository.FindByProviderId(identity.Subject)
	if err != nil {
		return nil, err
	}
	if record != nil {
		return s.issueTokenPair(record.Id, record.Role)
	}

	// First-time external login: auto-link to an existing local user by
	// verified email, or auto-provision a brand-new one.
	if !identity.EmailVerified {
		return nil, ErrEmailNotVerified
	}

	existing, err := s.Repository.UserRepository.FindByEmail(identity.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		if err := s.Repository.UserRepository.LinkProvider(existing.Id, model.ProviderOIDC, identity.Subject); err != nil {
			return nil, err
		}
		return s.issueTokenPair(existing.Id, existing.Role)
	}

	userId, err := s.Repository.UserRepository.CreateExternal(identity.Email, identity.FirstName, identity.LastName, model.ProviderOIDC, identity.Subject)
	if err != nil {
		return nil, err
	}
	return s.issueTokenPair(userId, model.RoleUser)
}

func (s *authServiceImpl) issueTokenPair(userId, role string) (*model.TokenPair, error) {
	accessToken, err := issueAccessToken(userId, role)
	if err != nil {
		return nil, err
	}

	rawRefreshToken, hash, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	expiry := time.Duration(config.Int("authentication.refreshTokenExpiryMinutes")) * time.Minute
	if err := s.Repository.RefreshTokenRepository.Create(userId, hash, time.Now().Add(expiry)); err != nil {
		return nil, err
	}

	return &model.TokenPair{AccessToken: accessToken, RefreshToken: rawRefreshToken}, nil
}
