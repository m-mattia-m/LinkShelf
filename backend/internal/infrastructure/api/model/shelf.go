package model

// PublicShelf is the subset of a shelf's fields that are safe to expose to
// anonymous visitors of the public link page.
type PublicShelf struct {
	Id          string `json:"id" bson:"id"`
	Title       string `json:"title" bson:"title"`
	Description string `json:"description" bson:"description"`
	Icon        string `json:"icon" bson:"icon"`
	Path        string `json:"path" bson:"path"`
}

type Shelf struct {
	PublicShelf
	Domain string `json:"domain" bson:"domain"`
	Theme  string `json:"theme" bson:"theme"`
	UserId string `json:"userId" bson:"userId"`
}

type ShelfBase struct {
	Title       string `json:"title" bson:"title" required:"true" minLength:"1"`
	Path        string `json:"path" bson:"path" required:"true" pattern:"^[a-zA-Z0-9-]*$" patternDescription:"letters, numbers, and hyphens only"`
	Domain      string `json:"domain" bson:"domain" required:"false"`
	Description string `json:"description" bson:"description" required:"false"`
	Theme       string `json:"theme" bson:"theme" required:"false"`
	Icon        string `json:"icon" bson:"icon" required:"false"`
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
