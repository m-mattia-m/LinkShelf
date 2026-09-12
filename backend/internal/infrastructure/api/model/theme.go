package model

const (
	ThemeScopeInstance = "instance"
	ThemeScopeUser     = "user"
)

type Theme struct {
	Id          string `json:"id" bson:"id"`
	Scope       string `json:"scope" bson:"scope"`
	OwnerUserId string `json:"ownerUserId,omitempty" bson:"ownerUserId"`
	Name        string `json:"name" bson:"name"`
	SourceFile  string `json:"sourceFile,omitempty" bson:"sourceFile"`
	Config      string `json:"config" bson:"config"`
}

type ThemeBase struct {
	Name   string `json:"name" bson:"name" required:"true" minLength:"1"`
	Config string `json:"config" bson:"config" required:"true"`
}

type ThemeRequestBody struct {
	Body ThemeBase `json:"body" bson:"body"`
}

type ThemeRequestFilter struct {
	ThemeId string `path:"themeId"`
}

type ThemeFilterAndBody struct {
	ThemeRequestFilter
	Body ThemeBase `json:"body" bson:"body"`
}

type ThemeResponse struct {
	Body Theme `json:"body" bson:"body"`
}

type ThemeListResponse struct {
	Body []Theme `json:"body" bson:"body"`
}

// ThemeGroupedResponseBody groups themes the way the shelf's theme picker
// displays them: instance-provided vs the caller's own.
type ThemeGroupedResponseBody struct {
	Instance []Theme `json:"instance" bson:"instance"`
	Mine     []Theme `json:"mine" bson:"mine"`
}

type ThemeGroupedResponse struct {
	Body ThemeGroupedResponseBody `json:"body" bson:"body"`
}
