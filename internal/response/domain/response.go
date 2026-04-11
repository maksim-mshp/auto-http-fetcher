package domain

import (
	"auto-http-fetcher/pkg/uuid"
	"errors"
	"time"
)

type ResponseType string
type ResponseStatus string

const (
	ManualType    ResponseType = "Manual"
	ScheduledType ResponseType = "Scheduled"
)

const (
	PendingStatus ResponseStatus = "Pending"
	SuccessStatus ResponseStatus = "Success"
	FailedStatus  ResponseStatus = "Failed"
)

func (t ResponseType) IsValid() bool {
	return t == ManualType || t == ScheduledType
}

type Response struct {
	ID         string
	WebhookID  string
	Type       ResponseType
	Status     ResponseStatus
	StatusCode int
	Body       []byte
	Headers    map[string]string
	StartedAt  time.Time
	FinishedAt *time.Time
	Attempt    int
	Duration   time.Duration
}

func NewResponse(webhookId string, t ResponseType) (*Response, error) {
	if !t.IsValid() {
		return nil, errors.New("invalid response type")
	}
	return &Response{
		ID:        uuid.Generate(),
		WebhookID: webhookId,
		Type:      t,
		Status:    PendingStatus,
		StartedAt: time.Now(),
		Attempt:   1,
	}, nil
}

func (r *Response) Complete(statusCode int, body []byte, headers map[string]string, duration time.Duration) {
	now := time.Now()
	r.FinishedAt = &now
	r.StatusCode = statusCode
	r.Body = body
	r.Headers = headers
	r.Duration = duration
	if statusCode >= 200 && statusCode < 300 {
		r.Status = SuccessStatus
	} else {
		r.Status = FailedStatus
	}
}

func (r *Response) Retry() {
	r.Attempt++
	r.Status = PendingStatus
	r.FinishedAt = nil
}

// func (r *Response) IsRetryable() bool {
// 	// return r.Attempt < MaxAttempt
// }
