package model

type Setting struct {
	Key          string `json:"key" bson:"key" required:"true"`
	LanguageCode string `json:"language_code" bson:"language_code" required:"true"`
	Value        string `json:"value" bson:"value" required:"true"`
}

type SettingRequest struct {
	Body Setting `json:"body" bson:"body"`
}

type SettingPageResponse struct {
	Body SettingPageBody `json:"body" bson:"body"`
}

type SettingPageBody struct {
	AboutShow           bool   `json:"about_show" bson:"about_show"`
	About               string `json:"about" bson:"about"`
	ContactShow         bool   `json:"contact_show" bson:"contact_show"`
	Contact             string `json:"contact" bson:"contact"`
	ImprintShow         bool   `json:"imprint_show" bson:"imprint_show"`
	Imprint             string `json:"imprint" bson:"imprint"`
	TermsOfUseShow      bool   `json:"terms_of_use_show" bson:"terms_of_use_show"`
	TermsOfUse          string `json:"terms_of_use" bson:"terms_of_use"`
	PrivacyPolicyShow   bool   `json:"privacy_policy_show" bson:"privacy_policy_show"`
	PrivacyPolicy       string `json:"privacy_policy" bson:"privacy_policy"`
	RedirectToDashboard bool   `json:"redirect_to_dashboard" bson:"redirect_to_dashboard"`
	OidcEnabled         bool   `json:"oidc_enabled" bson:"oidc_enabled"`
	// RegistrationEnabled and EmailVerificationEnabled mirror the auth config.
	RegistrationEnabled      bool `json:"registration_enabled" bson:"registration_enabled"`
	EmailVerificationEnabled bool `json:"email_verification_enabled" bson:"email_verification_enabled"`
	// UserBasedPaths mirrors app.userBasedPaths.
	UserBasedPaths bool `json:"user_based_paths" bson:"user_based_paths"`
	// PasswordResetEnabled mirrors authentication.passwordReset.enabled.
	PasswordResetEnabled bool `json:"password_reset_enabled" bson:"password_reset_enabled"`
	// LocalAuthEnabled mirrors authentication.localAuthEnabled.
	LocalAuthEnabled bool `json:"local_auth_enabled" bson:"local_auth_enabled"`
	// UpgradeUrl mirrors limits.upgradeUrl, empty if unset.
	UpgradeUrl string `json:"upgrade_url" bson:"upgrade_url"`
}

// EmailDeliveryInfo is the admin-only SMTP summary (no credentials).
type EmailDeliveryInfo struct {
	Enabled bool   `json:"enabled" bson:"enabled"`
	Host    string `json:"host" bson:"host"`
	From    string `json:"from" bson:"from"`
}

type EmailDeliveryInfoResponse struct {
	Body EmailDeliveryInfo `json:"body" bson:"body"`
}

type SettingRequestFiler struct {
	LanguageCode string `json:"language_code" bson:"language_code" query:"language_code"`
}

type SettingBatchRequestBody struct {
	Settings []Setting `json:"settings" bson:"settings" required:"true"`
}

type SettingBatchRequest struct {
	Body SettingBatchRequestBody `json:"body" bson:"body"`
}

// SettingUpdateFailure reports why a batch item was not saved.
type SettingUpdateFailure struct {
	Key          string `json:"key" bson:"key"`
	LanguageCode string `json:"language_code" bson:"language_code"`
	Reason       string `json:"reason" bson:"reason"`
}

type SettingBatchResponseBody struct {
	Settings SettingPageBody        `json:"settings" bson:"settings"`
	Failures []SettingUpdateFailure `json:"failures" bson:"failures"`
}

type SettingBatchResponse struct {
	Body SettingBatchResponseBody `json:"body" bson:"body"`
}
