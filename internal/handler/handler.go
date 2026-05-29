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

func HandleResponse(ctx *gin.Context, data any, err error) {

	if err == nil {
		ctx.JSON(http.StatusOK, data)
		return
	}

	var appError *myblogError.AppError
	if !errors.As(err, &appError) {
		slog.Error("http_handle HandleResponse fail", "error", err)
		errorResponse(ctx, nil, &myblogError.AppError{
			Code:         http.StatusInternalServerError,
			ErrorMessage: "Internal server error",
			ErrorStack:   err,
		})
		return
	}

	errorResponse(ctx, nil, appError)
}

func BindAndCheck(ctx *gin.Context, data any) bool {
	if err := ctx.ShouldBind(data); err != nil {
		slog.Error("http_handle BindAndCheck fail", "error", err)
		appError := myblogError.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: "Invalid request payload",
			ErrorStack:   err,
		}
		HandleResponse(ctx, nil, appError)
		return true
	}

	// do validation
	errFields, err := validator.GetValidator().Check(data)
	if err != nil {
		appError := myblogError.AppError{
			Code:         http.StatusBadRequest,
			ErrorMessage: "Invalid request payload",
			ErrorStack:   err,
		}
		HandleResponse(ctx, errFields, appError)
		return true
	}

	return false
}

func errorResponse(ctx *gin.Context, headers http.Header, appError *myblogError.AppError) {
	errorDetails := map[string]any{}

	if appError.ErrorMessage != "" {
		errorDetails["errorMessage"] = appError.ErrorMessage
	}

	if appError.ErrorDetails != nil {
		errorDetails["errorDetails"] = appError.ErrorDetails
	}

	var attrs []slog.Attr
	attrs = append(attrs, slog.String("request_url", ctx.Request.URL.String()))
	attrs = append(attrs, slog.String("request_method", ctx.Request.Method))
	if appError.ErrorStack != nil {
		attrs = append(attrs, slog.String("stack", xerrors.Sprint(appError.ErrorStack)))
	}

	for key, valueData := range appError.ErrorDetails {
		attrs = append(attrs, slog.Any(key, valueData))
	}

	slog.LogAttrs(ctx, slog.LevelError, "Error in handling request", attrs...)
	ctx.BindHeader(headers)
	ctx.JSON(appError.Code, errorDetails)
}
