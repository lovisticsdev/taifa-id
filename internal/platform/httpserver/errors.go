package httpserver

import (
	"errors"
	"net/http"
)

type ErrorCode string

const (
	ErrorCodeValidation    ErrorCode = "VALIDATION_ERROR"
	ErrorCodeNotFound      ErrorCode = "NOT_FOUND"
	ErrorCodeConflict      ErrorCode = "CONFLICT"
	ErrorCodeUnauthorized  ErrorCode = "UNAUTHORIZED"
	ErrorCodeForbidden     ErrorCode = "FORBIDDEN"
	ErrorCodeInternal      ErrorCode = "INTERNAL_ERROR"
	ErrorCodeInvalidJSON   ErrorCode = "INVALID_JSON"
	ErrorCodeMethodNotAllowed ErrorCode = "METHOD_NOT_ALLOWED"
)

type AppError struct {
	Code       ErrorCode
	Message    string
	StatusCode int
	Cause      error
}

func (e *AppError) Error() string {
	if e == nil {
		return ""
	}

	if e.Cause != nil {
		return e.Message + ": " + e.Cause.Error()
	}

	return e.Message
}

func (e *AppError) Unwrap() error {
	if e == nil {
		return nil
	}

	return e.Cause
}

func NewAppError(statusCode int, code ErrorCode, message string, cause error) *AppError {
	return &AppError{
		Code:       code,
		Message:    message,
		StatusCode: statusCode,
		Cause:      cause,
	}
}

func WriteError(w http.ResponseWriter, r *http.Request, statusCode int, code ErrorCode, message string) {
	WriteJSON(w, r, statusCode, map[string]any{
		"error": map[string]any{
			"code":           string(code),
			"message":        message,
			"correlation_id": CorrelationIDFromContext(r.Context()),
		},
	})
}

func WriteAppError(w http.ResponseWriter, r *http.Request, err error) {
	var appErr *AppError
	if errors.As(err, &appErr) {
		WriteError(w, r, appErr.StatusCode, appErr.Code, appErr.Message)
		return
	}

	WriteError(
		w,
		r,
		http.StatusInternalServerError,
		ErrorCodeInternal,
		"An internal error occurred.",
	)
}