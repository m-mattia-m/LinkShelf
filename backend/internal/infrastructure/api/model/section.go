package model

type Section struct {
	Id string `json:"id" bson:"id"`
	SectionBase
}

type SectionBase struct {
	Title   string `json:"title" bson:"title"`
	ShelfId string `json:"shelfId" bson:"shelfId"`
}

type SectionRequestBody struct {
	Body SectionBase `json:"body" bson:"body"`
}

type SectionRequestFilter struct {
	SectionRequestShelfFilter
	SectionRequestSectionFilter
}

type SectionRequestShelfFilter struct {
	ShelfId string `query:"shelfId"`
}

type SectionRequestSectionFilter struct {
	SectionId string `path:"sectionId"`
}

type SectionFilterFilterAndBody struct {
	SectionRequestFilter
	Body SectionBase `json:"body" bson:"body"`
}

type SectionRequestSectionFilterAndBody struct {
	SectionRequestSectionFilter
	Body SectionBase `json:"body" bson:"body"`
}

type SectionResponse struct {
	Body Section `json:"body" bson:"body"`
}

type SectionResponseList struct {
	Body []Section `json:"body" bson:"body"`
}
