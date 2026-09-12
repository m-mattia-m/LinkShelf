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
	// RegistrationEnabled and EmailVerificationEnabled mirror the
	// authentication.registrationEnabled / authentication.emailVerification.enabled
	// config so the frontend can adapt (hide sign-up, show a "check your
	// email" step) without guessing from a failed request.
	RegistrationEnabled      bool `json:"registration_enabled" bson:"registration_enabled"`
	EmailVerificationEnabled bool `json:"email_verification_enabled" bson:"email_verification_enabled"`
}

// EmailDeliveryInfo is deliberately minimal (host + from address only, no
// port/credentials/TLS mode) and only ever served to an authenticated admin -
// SMTP itself is configured exclusively via config, never through the UI.
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

// SettingUpdateFailure reports why one item of a batch update was not saved -
// either it failed validation (never reached the database) or the upsert
// itself errored.
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
