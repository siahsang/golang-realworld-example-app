package schema

type UserLoginPayloadReq struct {
	Email    string `json:"email" validate:"required,email,gt=0,lte=500"`
	Password string `json:"password" validate:"required,gt=8,lte=32"`
}

type UserRegisterPayloadReq struct {
	UserLoginPayloadReq `json:"user" validate:"required"`
}
