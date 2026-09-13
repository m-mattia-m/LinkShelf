package model

type Link struct {
	Id    string `json:"id" bson:"id"`
	Order int    `json:"order" bson:"order"`
	LinkBase
}

type LinkBase struct {
	Title     string `json:"title" bson:"title" required:"true" minLength:"1"`
	Link      string `json:"link" bson:"link" required:"true"`
	Icon      string `json:"icon" bson:"icon" required:"false"`
	Color     string `json:"color" bson:"color" required:"false"`
	SectionId string `json:"sectionId" bson:"sectionId" required:"true"`
}

// LinkOrderItem is one entry of a batch reorder request - the link's id and
// the new position it should be moved to within its section.
type LinkOrderItem struct {
	Id    string `json:"id" bson:"id" required:"true"`
	Order int    `json:"order" bson:"order" required:"true"`
}

type LinkOrderRequestBody struct {
	Links []LinkOrderItem `json:"links" bson:"links" required:"true"`
}

type LinkOrderRequest struct {
	Body LinkOrderRequestBody `json:"body" bson:"body"`
}

// LinkOrderFailure reports why one item of a batch reorder was not saved.
type LinkOrderFailure struct {
	Id     string `json:"id" bson:"id"`
	Reason string `json:"reason" bson:"reason"`
}

type LinkOrderResponseBody struct {
	Failures []LinkOrderFailure `json:"failures" bson:"failures"`
}

type LinkOrderResponse struct {
	Body LinkOrderResponseBody `json:"body" bson:"body"`
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
