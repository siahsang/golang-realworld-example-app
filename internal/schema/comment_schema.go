package schema

type CreateCommentPayload struct {
	Body string `json:"body" validate:"required,notblank"`
}

type CreateCommentRequest struct {
	Comment CreateCommentPayload `json:"comment" validate:"required"`
}
