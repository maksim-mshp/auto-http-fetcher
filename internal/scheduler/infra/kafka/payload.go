package kafka

type webhookPayload struct {
	ID int `json:"id"`

	Interval string `json:"interval"`
	Timeout  string `json:"timeout"`

	URL     string              `json:"url"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}

type WebhookCreatedPayload webhookPayload
type WebhookUpdatedPayload webhookPayload
type WebhookFetchPayload webhookPayload

type WebhookRetryPayload struct {
	ID      int `json:"id"`
	Attempt int `json:"attempt"`
}
