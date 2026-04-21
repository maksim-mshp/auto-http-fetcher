package kafka

type APIError struct {
	ErrorCode string         `json:"error"`
	Details   map[string]any `json:"details,omitempty"`
}

func (err APIError) Error() string {
	return err.ErrorCode
}

var (
	ErrInvalidBody = APIError{
		ErrorCode: "INVALID_BODY",
	}
)

func ErrNewWithDetails(err *APIError, key, value string) *APIError {
	err.Details = map[string]any{key: value}
	return err
}
