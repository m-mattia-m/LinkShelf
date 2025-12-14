package model

type Link struct {
	Id string `json:"id" bson:"id"`
	LinkBase
}

type LinkBase struct {
	Title     string `json:"title" bson:"title"`
	Link      string `json:"link" bson:"link"`
	Icon      string `json:"icon" bson:"icon"`
	Color     string `json:"color" bson:"color"`
	SectionId string `json:"sectionId" bson:"sectionId"`
}

type LinkRequestBody struct {
	Body LinkBase `json:"body" bson:"body"`
}

type LinkRequestFilter struct {
	ShelfId string `path:"shelfId"`
	LinkId  string `path:"linkId"`
}

type LinkFilterFilterAndBody struct {
	LinkRequestFilter
	Body LinkBase `json:"body" bson:"body"`
}

type LinkResponse struct {
	Body Link `json:"body" bson:"body"`
}

type LinkResponseList struct {
	Body []Link `json:"body" bson:"body"`
}
