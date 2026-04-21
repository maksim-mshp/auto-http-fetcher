package kafka

type WebhookKafkaDTO struct {
	Action      string `json:"action"`
	ID          int    `json:"id"`
	Description string `json:"description"`

	Interval int `json:"interval"`
	Timeout  int `json:"timeout"`

	URL     string              `json:"url"`
	Method  string              `json:"method"`
	Headers map[string][]string `json:"headers"`
	Body    []byte              `json:"body"`
}
