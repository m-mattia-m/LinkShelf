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

// resendCooldown rate-limits verification and reset emails per account.
const resendCooldown = 60 * time.Second

type EmailVerificationService interface {
	// SendInitial sends an invite link to passwordless users, otherwise a verify link.
	SendInitial(user *model.User) error
	// Resend re-sends the pending link. It never reveals whether the address exists.
	Resend(email string) error
	// VerifyEmail verifies an account that already has a password.
	VerifyEmail(rawToken string) error
	// SetPassword completes an invite: sets the first password and verifies the account.
	SetPassword(rawToken, newPassword string) error
	// RequestPasswordReset emails a reset link. It never reveals whether the address exists.
	RequestPasswordReset(email string) error
	// ResetPassword sets the new password, marks the email verified and ends all sessions.
	ResetPassword(rawToken, newPassword string) error
	// MarkVerified is the admin override without a token.
	MarkVerified(userId string) error
	// RequestEmailChange stores newEmail as pending and emails a confirm link to it.
	// delivered is false if the link could not be sent.
	RequestEmailChange(userId, currentEmail, newEmail string) (delivered bool, err error)
	// ConfirmEmailChange applies a pending email change. ErrConflict if the address was taken meanwhile.
	ConfirmEmailChange(rawToken string) error
	// NotifyProviderLinked emails the owner that an SSO identity was linked. Best-effort.
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

// send enforces the cooldown, then issues and emails a fresh token.
func (s *emailVerificationServiceImpl) send(userId, email, action string) error {
	// Email verification is disabled.
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

	// Logged here because several callers ignore the error.
	if sendErr != nil {
		zap.L().Error("failed to send account email", zap.String("action", action), zap.Error(sendErr))
	}

	return sendErr
}

// tokenLifetime is how long a link works; reset links are short-lived.
func tokenLifetime(action string) time.Duration {
	if action == repository.EmailActionResetPassword {
		return time.Duration(config.Int("authentication.passwordReset.tokenExpiryMinutes")) * time.Minute
	}
	return time.Duration(config.Int("authentication.emailVerification.tokenExpiryHours")) * time.Hour
}

// passwordResetEnabled is false without local auth.
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

	// End all sessions; the old password may have leaked.
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

// consumeToken validates a token for action and deletes the user's pending tokens of that action.
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

	// Delete older links first so none can confirm a different pending email.
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

	// The address may have been registered since the request.
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

// generateOpaqueToken returns a random token and its SHA-256 hash for storage.
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

// emailChangeRequestedMessage notifies the old address of a requested change.
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
