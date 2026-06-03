package handler

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/mdobak/go-xerrors"
	myblogError "github.com/siahsang/blog/internal/errors"
	"github.com/siahsang/blog/internal/validator"
)

type Handler struct {
	logger    *slog.Logger
	validator *validator.Validator
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{
		logger:    logger,
		validator: validator.NewValidator(logger),
	}
}

func (h *Handler) HandleResponse(ctx *gin.Context, data any, err error) {

	if err == nil {
		ctx.JSON(http.StatusOK, data)
		return
	}

	var appError *myblogError.AppError
	if !errors.As(err, &appError) {
		h.logger.Error("http_handle HandleResponse fail", "error", err)
		h.errorResponse(ctx, nil, &myblogError.AppError{
			Code:         http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
			ErrorStack:   err,
		})
		return
	}

	h.errorResponse(ctx, nil, appError)
}

func (h *Handler) BindAndCheck(ctx *gin.Context, data any) bool {
	if err := ctx.ShouldBind(data); err != nil {
		h.logger.Error("http_handle BindAndCheck fail", "error", err)
		appError := &myblogError.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: "Invalid request payload",
			ErrorStack:   err,
		}
		h.HandleResponse(ctx, nil, appError)
		return true
	}

	// do validation
	errFields, err := h.validator.Check(data)
	if err != nil {
		appError := &myblogError.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: "Invalid request payload",
			ErrorStack:   err,
			ErrorDetails: errFields,
		}
		h.HandleResponse(ctx, nil, appError)
		return true
	}

	return false
}

func (h *Handler) errorResponse(ctx *gin.Context, headers http.Header, appError *myblogError.AppError) {
	errorDetails := map[string]any{}

	if appError.ErrorDetails != nil {
		errorDetails["errorDetails"] = appError.ErrorDetails
	}

	var attrs []slog.Attr
	attrs = append(attrs, slog.String("request_url", ctx.Request.URL.String()))
	attrs = append(attrs, slog.String("request_method", ctx.Request.Method))
	if appError.ErrorStack != nil {
		attrs = append(attrs, slog.String("stack", xerrors.Sprint(appError.ErrorStack)))
	}

	for _, valueData := range appError.ErrorDetails {
		attrs = append(attrs, slog.Any(valueData.ErrorField, valueData.ErrorMsg))
	}

	h.logger.LogAttrs(ctx, slog.LevelError, "Error in handling request", attrs...)
	for key, values := range headers {
		for _, value := range values {
			ctx.Writer.Header().Add(key, value)
		}
	}

	ctx.JSON(appError.Code, errorDetails)
}
