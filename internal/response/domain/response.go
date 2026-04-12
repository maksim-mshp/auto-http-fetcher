package domain

import (
	"errors"
	"net/http"
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
	ID         int
	WebhookID  int
	Type       ResponseType
	Status     ResponseStatus
	StatusCode int
	Body       []byte
	Headers    http.Header
	StartedAt  time.Time
	FinishedAt *time.Time
	Attempt    int
	Duration   time.Duration
}

func NewResponse(webhookId int, t ResponseType) (*Response, error) {
	if !t.IsValid() {
		return nil, errors.New("invalid response type")
	}
	return &Response{
		WebhookID: webhookId,
		Type:      t,
		Status:    PendingStatus,
		StartedAt: time.Now(),
		Attempt:   1,
	}, nil
}

func (r *Response) Complete(statusCode int, body []byte, headers http.Header, duration time.Duration) {
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
