package model

type User struct {
	Id string `json:"id" bson:"id"`
	UserBase
}

type UserBase struct {
	Email     string `json:"email" bson:"email"`
	FirstName string `json:"first_name" bson:"first_name"`
	LastName  string `json:"last_name" bson:"last_name"`
	Password  string `json:"password" bson:"password"`
}

type UserRequestBody struct {
	Body UserBase `json:"body" bson:"body"`
}

type UserPatchPasswordFilterAndBody struct {
	UserRequestFilter
	Body UserRequestBodyOnlyPassword `json:"body" bson:"body"`
}

type UserRequestBodyOnlyPassword struct {
	OldPassword string `json:"old_password" bson:"old_password"`
	NewPassword string `json:"new_password" bson:"new_password"`
}

type UserRequestFilter struct {
	UserId string `path:"userId" doc:"The identifier of the chosen form you want."`
}

type UserFilterFilterAndBody struct {
	UserRequestFilter
	Body UserBase `json:"body" bson:"body"`
}

type UserResponse struct {
	Body User `json:"body" bson:"body"`
}
