package model

// PublicShelf is the subset of a shelf's fields that are safe to expose to
// anonymous visitors of the public link page. Theme is the resolved,
// validated set of CSS custom properties for the shelf's selected theme
// (nil if none is selected, or if the selected one no longer exists) - a
// theme's content is otherwise only ever visible to its owner, but once it's
// assigned to a public shelf its rendered values are inherently public
// anyway, so there's nothing left to protect by hiding this.
type PublicShelf struct {
	Id          string            `json:"id" bson:"id"`
	Title       string            `json:"title" bson:"title"`
	Description string            `json:"description" bson:"description"`
	Icon        string            `json:"icon" bson:"icon"`
	Path        string            `json:"path" bson:"path"`
	Theme       map[string]string `json:"theme" bson:"theme"`
	// NoIndex is true when the owner has opted this shelf's public page out
	// of search engine indexing. The public page uses it to render a
	// "noindex" robots meta tag - it never affects the instance's own
	// robots.txt, which stays permissive.
	NoIndex bool `json:"noIndex" bson:"noIndex"`
	// FooterEnabled is whether the shelf's public page shows a footer at all.
	// When true and FooterCustomText is empty, it shows the default "Powered
	// by LinkShelf" footer instead.
	FooterEnabled bool `json:"footerEnabled" bson:"footerEnabled"`
	// FooterCustomText, when set, replaces the default "Powered by LinkShelf"
	// footer. Rendered by the frontend as a restricted subset of Markdown
	// (bold, italic, links only).
	FooterCustomText string `json:"footerCustomText" bson:"footerCustomText"`
}

type Shelf struct {
	PublicShelf
	Domain string `json:"domain" bson:"domain"`
	UserId string `json:"userId" bson:"userId"`
	// Username is the owner's username, so a client can build the public URL
	// (/<username>/<path> with app.userBasedPaths) without a second request -
	// including for an admin looking at other people's shelves.
	Username string `json:"username" bson:"username"`
	// CreatedWithUserBasedPaths records whether app.userBasedPaths was on when
	// the shelf was created. It never changes afterwards. Compared with the
	// current setting it tells a client that the shelf's URL has changed shape
	// since it was created, and that links shared back then no longer work.
	CreatedWithUserBasedPaths bool `json:"createdWithUserBasedPaths" bson:"createdWithUserBasedPaths"`
	// ThemeId is the selected theme's id ("" if none selected).
	ThemeId string `json:"themeId" bson:"themeId"`
	// ThemeMissing is true when ThemeId is set but no longer resolves to an
	// existing theme (it was deleted, or an instance theme's file was
	// removed) - the edit page uses this to prompt picking a new one.
	ThemeMissing bool `json:"themeMissing" bson:"themeMissing"`
}

type ShelfBase struct {
	Title string `json:"title" bson:"title" required:"true" minLength:"1"`
	// A shelf is reached through a path or a domain, never both: exactly one
	// of Path and Domain has to be set. The API can't express that with tags,
	// so ShelfService enforces it.
	Path        string `json:"path" bson:"path" required:"false" pattern:"^[a-zA-Z0-9-]*$" patternDescription:"letters, numbers, and hyphens only"`
	Domain      string `json:"domain" bson:"domain" required:"false" doc:"A fully qualified domain name with an optional port, for example profile.example.com. Stored trimmed and lowercased, without a trailing dot or slash and without :80 or :443."`
	Description string `json:"description" bson:"description" required:"false"`
	ThemeId     string `json:"themeId" bson:"themeId" required:"false"`
	Icon        string `json:"icon" bson:"icon" required:"false"`
	// NoIndex opts this shelf's public page out of search engine indexing.
	// Defaults to false (crawlable) when omitted.
	NoIndex bool `json:"noIndex" bson:"noIndex" required:"false" doc:"When true, the shelf's public page asks search engines not to index it. The instance itself, and every other shelf, is unaffected."`
	// FooterEnabled defaults to false when omitted, which is only correct for
	// a client that always sends it explicitly (the web UI defaults new
	// shelves to true itself). Existing shelves keep the DB column's own
	// default of true regardless.
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
