package controller

import (
	"backend/internal/domain"
	"backend/internal/infrastructure/api/model"
	"context"
	"errors"
	"strings"

	"github.com/danielgtaylor/huma/v2"
)

func Login(svc *domain.Service) func(c context.Context, input *model.LoginRequestBody) (*model.TokenResponse, error) {
	return func(c context.Context, input *model.LoginRequestBody) (*model.TokenResponse, error) {
		tokens, err := svc.AuthService.Login(input.Body.Email, input.Body.Password)
		if err != nil {
			if errors.Is(err, domain.ErrEmailVerificationPending) {
				return nil, huma.Error403Forbidden("please verify your email address - we've sent a new link", err)
			}
			return nil, huma.Error401Unauthorized("invalid email or password")
		}
		return &model.TokenResponse{Body: *tokens}, nil
	}
}

// ResendVerification is intentionally silent about whether the address
// exists, is already verified, or was rate-limited - all look identical to
// the caller, so it can never be used to enumerate accounts.
func ResendVerification(svc *domain.Service) func(c context.Context, input *model.ResendVerificationRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.ResendVerificationRequestBody) (*struct{}, error) {
		_ = svc.EmailVerificationService.Resend(input.Body.Email)
		return nil, nil
	}
}

func VerifyEmail(svc *domain.Service) func(c context.Context, input *model.VerifyEmailRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.VerifyEmailRequestBody) (*struct{}, error) {
		if err := svc.EmailVerificationService.VerifyEmail(input.Body.Token); err != nil {
			if errors.Is(err, domain.ErrInvalidToken) {
				return nil, huma.Error400BadRequest("invalid or expired verification link", err)
			}
			return nil, huma.Error400BadRequest("failed to verify email", err)
		}
		return nil, nil
	}
}

// SetPassword completes the admin-invite flow: consumes a set-password token
// and sets the account's first password, marking it verified in the process.
func SetPassword(svc *domain.Service) func(c context.Context, input *model.SetPasswordRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.SetPasswordRequestBody) (*struct{}, error) {
		if err := svc.EmailVerificationService.SetPassword(input.Body.Token, input.Body.NewPassword); err != nil {
			if errors.Is(err, domain.ErrInvalidToken) {
				return nil, huma.Error400BadRequest("invalid or expired invite link", err)
			}
			return nil, huma.Error400BadRequest("failed to set password", err)
		}
		return nil, nil
	}
}

func Refresh(svc *domain.Service) func(c context.Context, input *model.RefreshRequestBody) (*model.TokenResponse, error) {
	return func(c context.Context, input *model.RefreshRequestBody) (*model.TokenResponse, error) {
		tokens, err := svc.AuthService.Refresh(input.Body.RefreshToken)
		if err != nil {
			return nil, huma.Error401Unauthorized("invalid or expired refresh token")
		}
		return &model.TokenResponse{Body: *tokens}, nil
	}
}

func Logout(svc *domain.Service) func(c context.Context, input *model.LogoutRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.LogoutRequestBody) (*struct{}, error) {
		if err := svc.AuthService.Logout(input.Body.RefreshToken); err != nil {
			return nil, huma.Error400BadRequest("failed to logout", err)
		}
		return nil, nil
	}
}

func OidcLogin(svc *domain.Service) func(c context.Context, input *struct{}) (*model.OidcLoginResponse, error) {
	return func(c context.Context, input *struct{}) (*model.OidcLoginResponse, error) {
		body, err := svc.AuthService.OidcAuthorizationURL()
		if err != nil {
			return nil, huma.Error400BadRequest("failed to start OIDC login", err)
		}
		return &model.OidcLoginResponse{Body: *body}, nil
	}
}

// OidcCallback behaves as a link-to-my-account request when called with a
// valid Bearer token already present, and as a login-or-auto-provision
// request otherwise - so it deliberately carries no Security requirement of
// its own.
func OidcCallback(svc *domain.Service) func(c context.Context, input *model.OidcCallbackRequestBody) (*model.TokenResponse, error) {
	return func(c context.Context, input *model.OidcCallbackRequestBody) (*model.TokenResponse, error) {
		var currentUserId *string
		if bearer := strings.TrimSpace(strings.TrimPrefix(input.Authorization, "Bearer ")); bearer != "" {
			if claims, err := domain.ValidateAccessToken(bearer); err == nil {
				currentUserId = &claims.Subject
			}
		}

		tokens, err := svc.AuthService.OidcCallback(c, input.Body.Code, input.Body.State, currentUserId)
		if err != nil {
			if errors.Is(err, domain.ErrEmailNotVerifiedForLinking) {
				return nil, huma.Error409Conflict("an account with this email already exists and could not be auto-linked because the identity provider did not confirm this email address is verified", err)
			}
			if errors.Is(err, domain.ErrEmailNotVerified) {
				return nil, huma.Error409Conflict("a new account could not be created because the identity provider did not confirm this email address is verified", err)
			}
			return nil, huma.Error400BadRequest("failed to complete OIDC login", err)
		}
		return &model.TokenResponse{Body: *tokens}, nil
	}
}
