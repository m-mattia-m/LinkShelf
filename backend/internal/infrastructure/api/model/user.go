package model

type User struct {
	Id string `json:"id" bson:"id"`
	UserBase
	EmailVerified bool `json:"email_verified" bson:"email_verified"`
	// HasPassword is false for an admin-invited account that hasn't set a
	// password yet.
	HasPassword bool `json:"has_password" bson:"has_password"`
	// EmailDeliveryFailed is set only by CreateUser, when the account was
	// created but its verification/invite email could not be sent.
	EmailDeliveryFailed bool `json:"email_delivery_failed,omitempty" bson:"-"`
}

type UserBase struct {
	Email string `json:"email" bson:"email" required:"true"`
	// Username is part of a shelf's public URL when app.userBasedPaths is
	// enabled. Reserved and taken names are rejected by the domain layer.
	Username  string `json:"username" bson:"username" required:"true" minLength:"3" maxLength:"30" pattern:"^[a-z0-9]([a-z0-9-]*[a-z0-9])?$" patternDescription:"lowercase letters, numbers, and hyphens, not starting or ending with a hyphen"`
	FirstName string `json:"first_name" bson:"first_name" required:"true"`
	LastName  string `json:"last_name" bson:"last_name" required:"true"`
	Role      string `json:"role" bson:"role" doc:"The user's role, e.g. 'user' or 'admin'. Only an admin caller may set this - ignored otherwise." required:"false"`
}

type UserCreate struct {
	UserBase
	// Password is required for self-registration but optional when an admin
	// creates the account, which then gets an invite link instead.
	Password string `json:"password" bson:"password" required:"false"`
}

type UserRequestBody struct {
	// Authorization is optional; an admin caller may also set the new user's role.
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

type ResendVerificationRequest struct {
	Email string `json:"email" bson:"email" required:"true"`
}

type ResendVerificationRequestBody struct {
	Body ResendVerificationRequest `json:"body" bson:"body"`
}

type VerifyEmailRequest struct {
	Token string `json:"token" bson:"token" required:"true"`
}

type VerifyEmailRequestBody struct {
	Body VerifyEmailRequest `json:"body" bson:"body"`
}

type ForgotPasswordRequest struct {
	Email string `json:"email" bson:"email" required:"true"`
}

type ForgotPasswordRequestBody struct {
	Body ForgotPasswordRequest `json:"body" bson:"body"`
}

type ResetPasswordRequest struct {
	Token       string `json:"token" bson:"token" required:"true"`
	NewPassword string `json:"new_password" bson:"new_password" required:"true" minLength:"8"`
}

type ResetPasswordRequestBody struct {
	Body ResetPasswordRequest `json:"body" bson:"body"`
}

type SetPasswordRequest struct {
	Token       string `json:"token" bson:"token" required:"true"`
	NewPassword string `json:"new_password" bson:"new_password" required:"true"`
}

type SetPasswordRequestBody struct {
	Body SetPasswordRequest `json:"body" bson:"body"`
}
