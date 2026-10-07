package schema

type ArticlePayloadReq struct {
	Title       string    `json:"title" validate:"required"`
	Description string    `json:"description"`
	Body        string    `json:"body"`
	TagList     *[]string `json:"tagList"`
}

type ArticlePayload struct {
	ArticlePayloadReq `json:"article"`
}
