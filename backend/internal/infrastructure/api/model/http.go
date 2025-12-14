package model

type SuccessResponse struct {
	Body HttpResponseBody `json:"body" bson:"body"`
}

type ErrorResponse struct {
	Body HttpResponseBody `json:"body" bson:"body"`
}

type HttpResponseBody struct {
	Message string `json:"message" bson:"message"`
}
