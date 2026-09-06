package model

type Link struct {
	Id string `json:"id" bson:"id"`
	LinkBase
}

type LinkBase struct {
	Title     string `json:"title" bson:"title" required:"true"`
	Link      string `json:"link" bson:"link" required:"true"`
	Icon      string `json:"icon" bson:"icon" required:"false"`
	Color     string `json:"color" bson:"color" required:"false"`
	SectionId string `json:"sectionId" bson:"sectionId" required:"true"`
}

type LinkRequestBody struct {
	Body LinkBase `json:"body" bson:"body"`
}

type LinkRequestFilter struct {
	LinkRequestShelfFilter
	LinkId string `path:"linkId"`
}

type LinkRequestShelfFilter struct {
	ShelfId string `query:"shelfId"`
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
