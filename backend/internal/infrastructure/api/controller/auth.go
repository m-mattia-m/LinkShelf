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
			if errors.Is(err, domain.ErrLocalAuthDisabled) {
				return nil, huma.Error403Forbidden(err.Error())
			}
			if errors.Is(err, domain.ErrEmailVerificationPending) {
				return nil, huma.Error403Forbidden("please verify your email address - we've sent a new link", err)
			}
			return nil, huma.Error401Unauthorized("invalid email or password")
		}
		return &model.TokenResponse{Body: *tokens}, nil
	}
}

// ResendVerification never reveals whether the address exists.
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

// ConfirmEmailChange applies a pending email change.
func ConfirmEmailChange(svc *domain.Service) func(c context.Context, input *model.VerifyEmailRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.VerifyEmailRequestBody) (*struct{}, error) {
		if err := svc.EmailVerificationService.ConfirmEmailChange(input.Body.Token); err != nil {
			if errors.Is(err, domain.ErrInvalidToken) {
				return nil, huma.Error400BadRequest("invalid or expired confirmation link", err)
			}
			if errors.Is(err, domain.ErrConflict) {
				return nil, huma.Error409Conflict("this email address is already used by another account", err)
			}
			return nil, huma.Error400BadRequest("failed to confirm the new email", err)
		}
		return nil, nil
	}
}

// SetPassword completes an invite by setting the first password.
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

// ForgotPassword emails a reset link. It never reveals whether the address exists.
func ForgotPassword(svc *domain.Service) func(c context.Context, input *model.ForgotPasswordRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.ForgotPasswordRequestBody) (*struct{}, error) {
		err := svc.EmailVerificationService.RequestPasswordReset(input.Body.Email)
		if errors.Is(err, domain.ErrPasswordResetDisabled) {
			return nil, huma.Error403Forbidden(err.Error())
		}
		// Every other outcome, including a failed send, looks the same.
		return nil, nil
	}
}

// ResetPassword completes the forgot-password flow with the emailed token.
func ResetPassword(svc *domain.Service) func(c context.Context, input *model.ResetPasswordRequestBody) (*struct{}, error) {
	return func(c context.Context, input *model.ResetPasswordRequestBody) (*struct{}, error) {
		if err := svc.EmailVerificationService.ResetPassword(input.Body.Token, input.Body.NewPassword); err != nil {
			switch {
			case errors.Is(err, domain.ErrPasswordResetDisabled):
				return nil, huma.Error403Forbidden(err.Error())
			case errors.Is(err, domain.ErrInvalidToken):
				return nil, huma.Error400BadRequest("invalid or expired reset link", err)
			default:
				return nil, huma.Error400BadRequest("failed to reset password", err)
			}
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

// OidcCallback links to the current account when a Bearer token is present, otherwise logs in.
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
			if errors.Is(err, domain.ErrAccountNotLinkable) {
				return nil, huma.Error409Conflict("an account with this email already exists, but it can't be linked to this single sign-on login automatically - sign in to it with its password first, or ask an administrator", err)
			}
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
