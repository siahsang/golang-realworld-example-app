package schema

type UserLoginPayloadReq struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserRegisterPayloadReq struct {
	UserLoginPayloadReq `json:"user"`
}
