package domain

import (
	"net/http"
	"net/url"
	"time"
)

type Webhook struct {
	ID          int
	Description string

	Interval time.Duration
	Timeout  time.Duration

	URL     url.URL
	Method  string
	Headers http.Header
	Body    []byte
}
