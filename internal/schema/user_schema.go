package schema

type UserLoginPayload struct {
	Email    string `validate:"required,email,gt=2,lte=30" json:"email"`
	Password string `validate:"required,gte=8,lte=30" json:"password"`
}

type UserLoginPayloadReq struct {
	UserLoginPayload `json:"user" validate:"required"`
}

type UserRegisterReqPayload struct {
	Email    string `validate:"required,email,gt=2,lte=30" json:"email"`
	Username string `validate:"required,gt=4,lte=30" json:"username"`
	Password string `validate:"required,gte=8,lte=30" json:"password"`
}

type UserRegisterReq struct {
	UserRegisterReqPayload `json:"user"`
}

type ProfileResponse struct {
	Profile interface{} `json:"profile"`
}
