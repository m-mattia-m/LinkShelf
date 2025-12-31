package model

type Setting struct {
	Key          string `json:"key" bson:"key"`
	LanguageCode string `json:"language_code" bson:"language_code"`
	Value        string `json:"value" bson:"value"`
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
}

type SettingRequestFiler struct {
	LanguageCode string `json:"language_code" bson:"language_code" query:"language_code"`
}
