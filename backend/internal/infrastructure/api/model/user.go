package model

type User struct {
	Id string `json:"id" bson:"id"`
	UserBase
}

type UserBase struct {
	Email     string `json:"email" bson:"email" required:"true"`
	FirstName string `json:"first_name" bson:"first_name" required:"true"`
	LastName  string `json:"last_name" bson:"last_name" required:"true"`
	// Role is only ever applied when the caller is an authenticated admin -
	// silently ignored (self-registration, self profile-update) otherwise.
	// Enforced in the domain layer, not by this schema.
	Role string `json:"role" bson:"role" doc:"The user's role, e.g. 'user' or 'admin'. Only an admin caller may set this - ignored otherwise." required:"false"`
}

type UserCreate struct {
	UserBase
	Password string `json:"password" bson:"password" required:"true"`
}

type UserRequestBody struct {
	// Optional: when present and valid for an admin, the create request may
	// also set the new user's role. Absent (self-registration), the role is
	// always "user".
	Authorization string     `header:"Authorization"`
	Body          UserCreate `json:"body" bson:"body"`
}

type UserPatchPasswordFilterAndBody struct {
	UserRequestFilter
	Body UserRequestBodyOnlyPassword `json:"body" bson:"body"`
}

type UserRequestBodyOnlyPassword struct {
	OldPassword string `json:"old_password" bson:"old_password" required:"true"`
	NewPassword string `json:"new_password" bson:"new_password" required:"true"`
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

type UserListResponse struct {
	Body []User `json:"body" bson:"body"`
}
