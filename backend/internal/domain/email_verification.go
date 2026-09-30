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
	"html"
	"time"

	"go.uber.org/zap"
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
	// RequestEmailChange stores newEmail as the account's pending email,
	// emails a confirm link to it and a heads-up to currentEmail. The account
	// keeps currentEmail until the link is used. A repeat request within
	// resendCooldown is refused with ErrTooManyRequests. delivered is false
	// when the confirm link could not be sent (the request is still stored,
	// so asking again later recovers).
	RequestEmailChange(userId, currentEmail, newEmail string) (delivered bool, err error)
	// ConfirmEmailChange completes an email change with the emailed token. It
	// fails with ErrConflict if another account took the address meanwhile.
	ConfirmEmailChange(rawToken string) error
	// NotifyProviderLinked tells the owner of an account that a single
	// sign-on identity was linked to it. Best-effort: a failure is only
	// logged, and it is a no-op without a mailer.
	NotifyProviderLinked(email string)
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

	var sendErr error
	switch action {
	case repository.EmailActionSetPassword:
		sendErr = s.mailer.Send(setPasswordMessage(email, rawToken))
	case repository.EmailActionResetPassword:
		sendErr = s.mailer.Send(resetPasswordMessage(email, rawToken))
	default:
		sendErr = s.mailer.Send(verifyEmailMessage(email, rawToken))
	}

	// Log here because every issuance flow passes through and several callers
	// discard the error as best-effort.
	if sendErr != nil {
		zap.L().Error("failed to send account email", zap.String("action", action), zap.Error(sendErr))
	}

	return sendErr
}

// tokenLifetime is how long a freshly issued link works. A reset link is
// short-lived on purpose: it is a way into the account.
func tokenLifetime(action string) time.Duration {
	if action == repository.EmailActionResetPassword {
		return time.Duration(config.Int("authentication.passwordReset.tokenExpiryMinutes")) * time.Minute
	}
	return time.Duration(config.Int("authentication.emailVerification.tokenExpiryHours")) * time.Hour
}

// passwordResetEnabled is false as well when local (password) auth is off:
// a reset would set a password nobody can sign in with.
func passwordResetEnabled() bool {
	return config.Bool("authentication.passwordReset.enabled") && config.LocalAuthEnabled()
}

func (s *emailVerificationServiceImpl) RequestPasswordReset(email string) error {
	if !passwordResetEnabled() {
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
	if !passwordResetEnabled() {
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

func (s *emailVerificationServiceImpl) RequestEmailChange(userId, currentEmail, newEmail string) (bool, error) {
	if s.mailer == nil {
		return false, nil
	}

	latest, err := s.Repository.EmailActionTokenRepository.GetLatestByUserIdAndAction(userId, repository.EmailActionChangeEmail)
	if err != nil {
		return false, err
	}
	if latest != nil && time.Since(latest.CreatedAt) < resendCooldown {
		return false, ErrTooManyRequests
	}

	rawToken, hash, err := generateOpaqueToken()
	if err != nil {
		return false, err
	}

	// The order matters: a confirm token applies whatever pending_email holds
	// when it is used, so every older link has to be gone before
	// pending_email changes. Otherwise a link sent to one address could
	// confirm a different, never-verified one.
	if err := s.Repository.EmailActionTokenRepository.DeleteByUserIdAndAction(userId, repository.EmailActionChangeEmail); err != nil {
		return false, err
	}
	if err := s.Repository.UserRepository.SetPendingEmail(userId, newEmail); err != nil {
		return false, err
	}
	expiresAt := time.Now().Add(tokenLifetime(repository.EmailActionChangeEmail))
	if err := s.Repository.EmailActionTokenRepository.Create(userId, hash, repository.EmailActionChangeEmail, expiresAt); err != nil {
		return false, err
	}

	sendErr := s.mailer.Send(confirmEmailChangeMessage(newEmail, rawToken))
	if sendErr != nil {
		zap.L().Error("failed to send account email", zap.String("action", repository.EmailActionChangeEmail), zap.Error(sendErr))
	}
	if err := s.mailer.Send(emailChangeRequestedMessage(currentEmail, newEmail)); err != nil {
		zap.L().Error("failed to send email change notice to the current address", zap.Error(err))
	}

	return sendErr == nil, nil
}

func (s *emailVerificationServiceImpl) ConfirmEmailChange(rawToken string) error {
	token, err := s.consumeToken(rawToken, repository.EmailActionChangeEmail)
	if err != nil {
		return err
	}

	user, err := s.Repository.UserRepository.Get(token.UserId)
	if err != nil {
		return err
	}
	if user == nil || user.PendingEmail == "" {
		return ErrInvalidToken
	}

	// The address was free when the change was requested, but someone may
	// have registered it since.
	other, err := s.Repository.UserRepository.FindByEmail(user.PendingEmail)
	if err != nil {
		return err
	}
	if other != nil && other.Id != user.Id {
		if err := s.Repository.UserRepository.SetPendingEmail(user.Id, ""); err != nil {
			return err
		}
		return fmt.Errorf("%w: this email address is already used by another account", ErrConflict)
	}

	return s.Repository.UserRepository.ChangeEmail(user.Id, user.PendingEmail, true)
}

func (s *emailVerificationServiceImpl) NotifyProviderLinked(email string) {
	if s.mailer == nil {
		return
	}
	if err := s.mailer.Send(providerLinkedMessage(email)); err != nil {
		zap.L().Error("failed to send single sign-on linked notice", zap.Error(err))
	}
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

func confirmEmailChangeMessage(to, rawToken string) mailer.Message {
	link := fmt.Sprintf("%s/auth/confirm-email?token=%s", config.String("app.frontendUrl"), rawToken)
	hours := config.Int("authentication.emailVerification.tokenExpiryHours")
	return mailer.Message{
		To:      to,
		Subject: fmt.Sprintf("Confirm your new email for %s", config.String("app.name")),
		TextBody: fmt.Sprintf("Someone asked to use this address for their %s account. To confirm the change, open this link:\n\n%s\n\n"+
			"It expires in %d hours. Until then the account keeps its current email. If this wasn't you, you can ignore this email.",
			config.String("app.name"), link, hours),
		HTMLBody: fmt.Sprintf(`<p>Someone asked to use this address for their %s account. To confirm the change, click the link below.</p><p><a href="%s">Confirm my new email</a></p><p>It expires in %d hours. Until then the account keeps its current email. If this wasn't you, you can ignore this email.</p>`,
			html.EscapeString(config.String("app.name")), link, hours),
	}
}

// emailChangeRequestedMessage goes to the address being replaced, so a
// hijacked session can't quietly move the account to another inbox.
func emailChangeRequestedMessage(to, newEmail string) mailer.Message {
	return mailer.Message{
		To:      to,
		Subject: fmt.Sprintf("An email change was requested for your %s account", config.String("app.name")),
		TextBody: fmt.Sprintf("Someone asked to change the email of your %s account to %s. The change only happens once the new address is confirmed.\n\n"+
			"If this wasn't you, change your password and contact your administrator.", config.String("app.name"), newEmail),
		HTMLBody: fmt.Sprintf(`<p>Someone asked to change the email of your %s account to <strong>%s</strong>. The change only happens once the new address is confirmed.</p><p>If this wasn't you, change your password and contact your administrator.</p>`,
			html.EscapeString(config.String("app.name")), html.EscapeString(newEmail)),
	}
}

func providerLinkedMessage(to string) mailer.Message {
	return mailer.Message{
		To:      to,
		Subject: fmt.Sprintf("Single sign-on was linked to your %s account", config.String("app.name")),
		TextBody: fmt.Sprintf("A single sign-on login was just linked to your %s account, so it can now be used to sign in.\n\n"+
			"If this wasn't you, contact your administrator.", config.String("app.name")),
		HTMLBody: fmt.Sprintf(`<p>A single sign-on login was just linked to your %s account, so it can now be used to sign in.</p><p>If this wasn't you, contact your administrator.</p>`,
			html.EscapeString(config.String("app.name"))),
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
