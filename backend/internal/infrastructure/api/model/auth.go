package model

const (
	RoleAdmin = "admin"
	RoleUser  = "user"

	ProviderLocal = "LOCAL"
	ProviderOIDC  = "OIDC"
)

type LoginRequest struct {
	Email    string `json:"email" bson:"email" required:"true"`
	Password string `json:"password" bson:"password" required:"true"`
}

type LoginRequestBody struct {
	Body LoginRequest `json:"body" bson:"body"`
}

type TokenPair struct {
	AccessToken  string `json:"access_token" bson:"access_token"`
	RefreshToken string `json:"refresh_token" bson:"refresh_token"`
}

type TokenResponse struct {
	Body TokenPair `json:"body" bson:"body"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token" bson:"refresh_token" required:"true"`
}

type RefreshRequestBody struct {
	Body RefreshRequest `json:"body" bson:"body"`
}

type LogoutRequestBody struct {
	Body RefreshRequest `json:"body" bson:"body"`
}

type OidcLoginResponseBody struct {
	AuthorizationUrl string `json:"authorization_url" bson:"authorization_url"`
	State            string `json:"state" bson:"state"`
}

type OidcLoginResponse struct {
	Body OidcLoginResponseBody `json:"body" bson:"body"`
}

type OidcCallbackRequest struct {
	Code  string `json:"code" bson:"code" required:"true"`
	State string `json:"state" bson:"state" required:"true"`
}

type OidcCallbackRequestBody struct {
	Authorization string              `header:"Authorization" doc:"Optional bearer token. When present and valid, the external identity is linked to the already-authenticated user instead of logging in as a new/matched user."`
	Body          OidcCallbackRequest `json:"body" bson:"body"`
}
