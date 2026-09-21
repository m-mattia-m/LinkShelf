//go:generate mockgen -source=email_verification.go -destination=mocks/email_verification_service.go -package=mocks

package domain

import (
	"backend/internal/config"
	"backend/internal/infrastructure/api/model"
	"backend/internal/infrastructure/mailer"
	"backend/internal/infrastructure/repository"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"time"
)

// resendCooldown throttles both the explicit resend endpoint and the
// automatic resend triggered by a login attempt, so a single account can't be
// used to spam an inbox (or a third party's) with repeated emails.
const resendCooldown = 60 * time.Second

type EmailVerificationService interface {
	// SendInitial sends the first email for a newly created user: an
	// invite/set-password link if it has no password yet, otherwise a plain
	// verify-email link.
	SendInitial(user *model.User) error
	// Resend re-sends whatever link is still pending for the given email,
	// rate-limited. It never reports whether the address exists, is already
	// verified, or was rate-limited - all of those look identical to the
	// caller, by design.
	Resend(email string) error
	// VerifyEmail completes plain verification for an account that already
	// has a password.
	VerifyEmail(rawToken string) error
	// SetPassword completes the admin-invite flow: sets the first password
	// and marks the account verified, consuming the token.
	SetPassword(rawToken, newPassword string) error
	// RequestPasswordReset emails a reset link for the given address. Like
	// Resend it never reports whether the address exists or the request was
	// rate-limited, so it can't be used to enumerate accounts. It only errors
	// when the feature is disabled or something internal fails.
	RequestPasswordReset(email string) error
	// ResetPassword completes the forgot-password flow: sets the new
	// password, marks the address verified (the link proved control of it) and
	// ends every session of that account.
	ResetPassword(rawToken, newPassword string) error
	// MarkVerified is the admin override - forces a user's email to
	// verified without any token at all.
	MarkVerified(userId string) error
}

type emailVerificationServiceImpl struct {
	Repository *repository.Repository
	Domain     *Service
	mailer     mailer.Mailer
}

func NewEmailVerificationService(repo *repository.Repository, domain *Service, m mailer.Mailer) EmailVerificationService {
	return &emailVerificationServiceImpl{
		Repository: repo,
		Domain:     domain,
		mailer:     m,
	}
}

func (s *emailVerificationServiceImpl) SendInitial(user *model.User) error {
	if user.HasPassword {
		return s.send(user.Id, user.Email, repository.EmailActionVerify)
	}
	return s.send(user.Id, user.Email, repository.EmailActionSetPassword)
}

func (s *emailVerificationServiceImpl) Resend(email string) error {
	record, err := s.Repository.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}
	if record == nil || record.EmailVerified {
		return nil
	}

	action := repository.EmailActionVerify
	if record.Password == "" {
		action = repository.EmailActionSetPassword
	}

	return s.send(record.Id, record.Email, action)
}

// send enforces the resend cooldown, then (re)issues and emails a fresh
// token, replacing any previous still-pending one for the same action.
func (s *emailVerificationServiceImpl) send(userId, email, action string) error {
	// Nil mailer means authentication.emailVerification.enabled is false -
	// callers are expected to check that first, but this is a safe no-op
	// rather than a nil-pointer panic if one doesn't.
	if s.mailer == nil {
		return nil
	}

	latest, err := s.Repository.EmailActionTokenRepository.GetLatestByUserIdAndAction(userId, action)
	if err != nil {
		return err
	}
	if latest != nil && time.Since(latest.CreatedAt) < resendCooldown {
		return nil
	}

	rawToken, hash, err := generateOpaqueToken()
	if err != nil {
		return err
	}

	expiry := tokenLifetime(action)
	if err := s.Repository.EmailActionTokenRepository.DeleteByUserIdAndAction(userId, action); err != nil {
		return err
	}
	if err := s.Repository.EmailActionTokenRepository.Create(userId, hash, action, time.Now().Add(expiry)); err != nil {
		return err
	}

	switch action {
	case repository.EmailActionSetPassword:
		return s.mailer.Send(setPasswordMessage(email, rawToken))
	case repository.EmailActionResetPassword:
		return s.mailer.Send(resetPasswordMessage(email, rawToken))
	default:
		return s.mailer.Send(verifyEmailMessage(email, rawToken))
	}
}

// tokenLifetime is how long a freshly issued link works. A reset link is
// short-lived on purpose: it is a way into the account.
func tokenLifetime(action string) time.Duration {
	if action == repository.EmailActionResetPassword {
		return time.Duration(config.Int("authentication.passwordReset.tokenExpiryMinutes")) * time.Minute
	}
	return time.Duration(config.Int("authentication.emailVerification.tokenExpiryHours")) * time.Hour
}

func (s *emailVerificationServiceImpl) RequestPasswordReset(email string) error {
	if !config.Bool("authentication.passwordReset.enabled") {
		return ErrPasswordResetDisabled
	}

	record, err := s.Repository.UserRepository.FindByEmail(email)
	if err != nil {
		return err
	}
	if record == nil {
		return nil
	}

	return s.send(record.Id, record.Email, repository.EmailActionResetPassword)
}

func (s *emailVerificationServiceImpl) ResetPassword(rawToken, newPassword string) error {
	if !config.Bool("authentication.passwordReset.enabled") {
		return ErrPasswordResetDisabled
	}

	token, err := s.consumeToken(rawToken, repository.EmailActionResetPassword)
	if err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}
	if err := s.Repository.UserRepository.SetPassword(token.UserId, hashedPassword); err != nil {
		return err
	}

	// A reset is often done because the old password leaked, so nobody who
	// was signed in with it may stay signed in.
	return s.Repository.RefreshTokenRepository.DeleteByUserId(token.UserId)
}

func (s *emailVerificationServiceImpl) VerifyEmail(rawToken string) error {
	token, err := s.consumeToken(rawToken, repository.EmailActionVerify)
	if err != nil {
		return err
	}

	return s.Repository.UserRepository.MarkVerified(token.UserId)
}

func (s *emailVerificationServiceImpl) SetPassword(rawToken, newPassword string) error {
	token, err := s.consumeToken(rawToken, repository.EmailActionSetPassword)
	if err != nil {
		return err
	}

	hashedPassword, err := hashPassword(newPassword)
	if err != nil {
		return err
	}

	return s.Repository.UserRepository.SetPassword(token.UserId, hashedPassword)
}

// consumeToken validates a raw token against the given expected action and
// deletes every pending token of that action for the user (single-use).
func (s *emailVerificationServiceImpl) consumeToken(rawToken, expectedAction string) (*repository.EmailActionToken, error) {
	hash := hashOpaqueToken(rawToken)

	token, err := s.Repository.EmailActionTokenRepository.GetByHash(hash)
	if err != nil {
		return nil, err
	}
	if token == nil || token.Action != expectedAction || token.ExpiresAt.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	if err := s.Repository.EmailActionTokenRepository.DeleteByUserIdAndAction(token.UserId, expectedAction); err != nil {
		return nil, err
	}

	return token, nil
}

func (s *emailVerificationServiceImpl) MarkVerified(userId string) error {
	return s.Repository.UserRepository.MarkVerified(userId)
}

// generateOpaqueToken mirrors generateRefreshToken (jwt.go): a random opaque
// token handed out to the client, and the SHA-256 hash that's actually
// persisted, so a DB leak doesn't hand out usable links directly.
func generateOpaqueToken() (rawToken, hash string, err error) {
	buf := make([]byte, 32)
	if _, err = rand.Read(buf); err != nil {
		return "", "", err
	}

	rawToken = base64.RawURLEncoding.EncodeToString(buf)
	hash = hashOpaqueToken(rawToken)
	return rawToken, hash, nil
}

func hashOpaqueToken(rawToken string) string {
	sum := sha256.Sum256([]byte(rawToken))
	return hex.EncodeToString(sum[:])
}

func verifyEmailMessage(to, rawToken string) mailer.Message {
	link := fmt.Sprintf("%s/auth/verify-email?token=%s", config.String("app.frontendUrl"), rawToken)
	return mailer.Message{
		To:      to,
		Subject: fmt.Sprintf("Verify your email for %s", config.String("app.name")),
		TextBody: fmt.Sprintf("Please verify your email address by opening this link:\n\n%s\n\n"+
			"This link expires in %d hours.", link, config.Int("authentication.emailVerification.tokenExpiryHours")),
		HTMLBody: fmt.Sprintf(`<p>Please verify your email address by clicking the link below.</p><p><a href="%s">Verify my email</a></p><p>This link expires in %d hours.</p>`,
			link, config.Int("authentication.emailVerification.tokenExpiryHours")),
	}
}

func setPasswordMessage(to, rawToken string) mailer.Message {
	link := fmt.Sprintf("%s/auth/set-password?token=%s", config.String("app.frontendUrl"), rawToken)
	return mailer.Message{
		To:      to,
		Subject: fmt.Sprintf("Finish setting up your %s account", config.String("app.name")),
		TextBody: fmt.Sprintf("An account was created for you. Set your password to finish registration:\n\n%s\n\n"+
			"This link expires in %d hours.", link, config.Int("authentication.emailVerification.tokenExpiryHours")),
		HTMLBody: fmt.Sprintf(`<p>An account was created for you. Set your password to finish registration.</p><p><a href="%s">Set my password</a></p><p>This link expires in %d hours.</p>`,
			link, config.Int("authentication.emailVerification.tokenExpiryHours")),
	}
}

func resetPasswordMessage(to, rawToken string) mailer.Message {
	link := fmt.Sprintf("%s/auth/reset-password?token=%s", config.String("app.frontendUrl"), rawToken)
	minutes := config.Int("authentication.passwordReset.tokenExpiryMinutes")
	return mailer.Message{
		To:      to,
		Subject: fmt.Sprintf("Reset your %s password", config.String("app.name")),
		TextBody: fmt.Sprintf("Someone asked to reset the password of this account. To choose a new one, open this link:\n\n%s\n\n"+
			"It expires in %d minutes. If this wasn't you, you can ignore this email and your password stays as it is.", link, minutes),
		HTMLBody: fmt.Sprintf(`<p>Someone asked to reset the password of this account. To choose a new one, click the link below.</p><p><a href="%s">Reset my password</a></p><p>It expires in %d minutes. If this wasn't you, you can ignore this email and your password stays as it is.</p>`,
			link, minutes),
	}
}
