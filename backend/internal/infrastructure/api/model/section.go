package model

type Section struct {
	Id    string `json:"id" bson:"id"`
	Order int    `json:"order" bson:"order"`
	SectionBase
}

type SectionBase struct {
	Title   string `json:"title" bson:"title" required:"true" minLength:"1"`
	ShelfId string `json:"shelfId" bson:"shelfId" required:"true"`
}

// SectionOrderItem is one entry of a batch reorder request - the section's id
// and the new position it should be moved to within its shelf.
type SectionOrderItem struct {
	Id    string `json:"id" bson:"id" required:"true"`
	Order int    `json:"order" bson:"order" required:"true"`
}

type SectionOrderRequestBody struct {
	Sections []SectionOrderItem `json:"sections" bson:"sections" required:"true"`
}

type SectionOrderRequest struct {
	Body SectionOrderRequestBody `json:"body" bson:"body"`
}

// SectionOrderFailure reports why one item of a batch reorder was not saved.
type SectionOrderFailure struct {
	Id     string `json:"id" bson:"id"`
	Reason string `json:"reason" bson:"reason"`
}

type SectionOrderResponseBody struct {
	Failures []SectionOrderFailure `json:"failures" bson:"failures"`
}

type SectionOrderResponse struct {
	Body SectionOrderResponseBody `json:"body" bson:"body"`
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
