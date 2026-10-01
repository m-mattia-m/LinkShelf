package model

// PublicShelf is the subset of a shelf visible to anonymous visitors.
type PublicShelf struct {
	Id          string            `json:"id" bson:"id"`
	Title       string            `json:"title" bson:"title"`
	Description string            `json:"description" bson:"description"`
	Icon        string            `json:"icon" bson:"icon"`
	Path        string            `json:"path" bson:"path"`
	Theme       map[string]string `json:"theme" bson:"theme"`
	// NoIndex renders a "noindex" robots meta tag on the public page.
	NoIndex          bool   `json:"noIndex" bson:"noIndex"`
	FooterEnabled    bool   `json:"footerEnabled" bson:"footerEnabled"`
	FooterCustomText string `json:"footerCustomText" bson:"footerCustomText"`
}

type Shelf struct {
	PublicShelf
	Domain string `json:"domain" bson:"domain"`
	UserId string `json:"userId" bson:"userId"`
	// Username is used to build /<username>/<path> URLs.
	Username string `json:"username" bson:"username"`
	// CreatedWithUserBasedPaths records app.userBasedPaths at creation, to detect URL changes.
	CreatedWithUserBasedPaths bool `json:"createdWithUserBasedPaths" bson:"createdWithUserBasedPaths"`
	// ThemeId is the selected theme's id ("" if none selected).
	ThemeId string `json:"themeId" bson:"themeId"`
	// ThemeMissing is true when ThemeId is set but the theme no longer exists.
	ThemeMissing bool `json:"themeMissing" bson:"themeMissing"`
}

type ShelfBase struct {
	Title string `json:"title" bson:"title" required:"true" minLength:"1"`
	// Exactly one of Path and Domain must be set; ShelfService enforces it.
	Path             string `json:"path" bson:"path" required:"false" pattern:"^[a-zA-Z0-9-]*$" patternDescription:"letters, numbers, and hyphens only"`
	Domain           string `json:"domain" bson:"domain" required:"false" doc:"A fully qualified domain name with an optional port, for example profile.example.com. Stored trimmed and lowercased, without a trailing dot or slash and without :80 or :443."`
	Description      string `json:"description" bson:"description" required:"false"`
	ThemeId          string `json:"themeId" bson:"themeId" required:"false"`
	Icon             string `json:"icon" bson:"icon" required:"false"`
	NoIndex          bool   `json:"noIndex" bson:"noIndex" required:"false" doc:"When true, the shelf's public page asks search engines not to index it. The instance itself, and every other shelf, is unaffected."`
	FooterEnabled    bool   `json:"footerEnabled" bson:"footerEnabled" required:"false" doc:"Whether the shelf's public page shows a footer at all. Defaults to true - the shelf's public page shows the default \"Powered by LinkShelf\" footer, or FooterCustomText when set."`
	FooterCustomText string `json:"footerCustomText" bson:"footerCustomText" required:"false" maxLength:"500" doc:"Optional custom text shown in the footer instead of \"Powered by LinkShelf\", when FooterEnabled is true. Rendered as a restricted subset of Markdown: bold, italic, and links only."`
}

type ShelfRequestBody struct {
	Body ShelfBase `json:"body" bson:"body"`
}

type ShelfRequestFilter struct {
	ShelfId string `path:"shelfId"`
}

type ShelfPathFilter struct {
	Path string `path:"path"`
}

type ShelfDomainFilter struct {
	Domain string `path:"domain"`
}

type ShelfUsernamePathFilter struct {
	Username string `path:"username"`
	Path     string `path:"path"`
}

type ShelfFilterFilterAndBody struct {
	ShelfRequestFilter
	Body ShelfBase `json:"body" bson:"body"`
}

type ShelfResponse struct {
	Body Shelf `json:"body" bson:"body"`
}

type ShelfListResponse struct {
	Body []Shelf `json:"body" bson:"body"`
}

type PublicShelfResponse struct {
	Body PublicShelf `json:"body" bson:"body"`
}
