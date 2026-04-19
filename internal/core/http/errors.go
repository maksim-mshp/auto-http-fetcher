package http

import (
	"log/slog"
	"net/http"
)

type APIError struct {
	StatusCode int            `json:"-"`
	ErrorCode  string         `json:"error"`
	Details    map[string]any `json:"details,omitempty"`
}

func (e APIError) Error() string {
	return e.ErrorCode
}

var (
	ErrInternal = APIError{
		StatusCode: http.StatusInternalServerError,
		ErrorCode:  "INTERNAL_ERROR",
	}

	ErrInvalidBody = APIError{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  "INVALID_BODY",
	}

	ErrUserNotFound = APIError{
		StatusCode: http.StatusNotFound,
		ErrorCode:  "USER_NOT_FOUND",
	}

	ErrUserAlreadyExists = APIError{
		StatusCode: http.StatusConflict,
		ErrorCode:  "USER_ALREADY_EXISTS",
	}

	ErrInvalidEmail = APIError{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  "INVALID_EMAIL",
	}

	ErrInvalidUserID = APIError{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  "INVALID_USER_ID",
	}

	ErrVerificationFailed = APIError{
		StatusCode: http.StatusForbidden,
		ErrorCode:  "VERIFICATION_FAILED",
	}
)

func NewValidationError(field string, message string) *APIError {
	return &APIError{
		StatusCode: http.StatusBadRequest,
		ErrorCode:  "VALIDATION_ERROR",
		Details: map[string]any{
			"field":   field,
			"message": message,
		},
	}
}

func SendErrorJSON(logger *slog.Logger, w http.ResponseWriter, apiErr *APIError) {
	SendJSON(logger, w, apiErr, apiErr.StatusCode)
}
