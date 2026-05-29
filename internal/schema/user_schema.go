package schema

type UserLoginPayloadReq struct {
	Email    string `json:"email" validate:"required,email,gt=0,lte=500"`
	Password string `json:"password" validate:"required,gt=8,lte=32"`
}

type UserRegisterPayloadReq struct {
	UserLoginPayloadReq `json:"user" validate:"required"`
}


type UserRegisterReqPayload struct {
	Email    string `validate:"required,email,gt=2,lte=30" json:"email"`
	Username string `validate:"required,gt=4,lte=30" json:"username"`
	Password string `validate:"required,gte=8,lte=30" json:"password"`
}

type UserRegisterReq struct {
	UserRegisterReqPayload `json:"user"`
}

