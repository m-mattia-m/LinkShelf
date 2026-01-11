package model

type Shelf struct {
	Id string `json:"id" bson:"id"`
	ShelfBase
}

type ShelfBase struct {
	Title       string `json:"title" bson:"title"`
	Path        string `json:"path" bson:"path"`
	Domain      string `json:"domain" bson:"domain"`
	Description string `json:"description" bson:"description"`
	Theme       string `json:"theme" bson:"theme"`
	Icon        string `json:"icon" bson:"icon"`
	UserId      string `json:"userId" bson:"userId"`
}

type ShelfRequestBody struct {
	Body ShelfBase `json:"body" bson:"body"`
}

type ShelfRequestFilter struct {
	ShelfId string `path:"shelfId"`
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
