package http

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
)

func ParseJSONBody(logger *slog.Logger, r *http.Request, data any) *APIError {
	defer func() {
		if err := r.Body.Close(); err != nil {
			logger.Error("failed to close request body", "error", err.Error())
		}
	}()

	if ct := r.Header.Get("Content-Type"); ct != "" {
		mediaType, _, err := mime.ParseMediaType(ct)
		if err != nil || mediaType != "application/json" {
			return &APIError{
				StatusCode: http.StatusUnsupportedMediaType,
				ErrorCode:  "UNSUPPORTED_MEDIA_TYPE",
			}
		}
	}

	const maxBodyBytes = 1024 * 1024
	r.Body = http.MaxBytesReader(nil, r.Body, maxBodyBytes)

	dec := json.NewDecoder(r.Body)

	if err := dec.Decode(data); err != nil {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return &APIError{
				StatusCode: http.StatusRequestEntityTooLarge,
				ErrorCode:  "REQUEST_ENTITY_TOO_LARGE",
			}
		}
		return &ErrInvalidBody
	}

	if err := dec.Decode(&struct{}{}); err != nil && err != io.EOF {
		logger.Warn("failed to decode json", "error", err.Error())
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			return &APIError{
				StatusCode: http.StatusRequestEntityTooLarge,
				ErrorCode:  "REQUEST_ENTITY_TOO_LARGE",
			}
		}
		return &ErrInvalidBody
	}
	return nil
}

func SendJSON(logger *slog.Logger, w http.ResponseWriter, data interface{}, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		logger.Warn("failed to encode json", "error", err.Error())
	}
}
