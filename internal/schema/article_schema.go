package schema

type ArticlePayloadReq struct {
	Title       string    `json:"title" validate:"required,notblank"`
	Description string    `json:"description" validate:"required,notblank"`
	Body        string    `json:"body" validate:"required,notblank"`
	TagList     *[]string `json:"tagList"`
}

type ArticlePayload struct {
	ArticlePayloadReq `json:"article" validate:"required"`
}

type ArticleUpdatePayload struct {
	Title       *string `json:"title"`
	Description *string `json:"description"`
	Body        *string `json:"body"`
}

type ArticleUpdatePayloadReq struct {
	ArticleUpdatePayload `json:"article" validate:"required"`
}
