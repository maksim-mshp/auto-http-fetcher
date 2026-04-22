package grpc

type FetchHeaderValues struct {
	Values []string `json:"values"`
}

type FetchRequest struct {
	ID          int64                        `json:"id"`
	Description string                       `json:"description"`
	IntervalMs  int64                        `json:"intervalMs"`
	TimeoutMs   int64                        `json:"timeoutMs"`
	URL         string                       `json:"url"`
	Method      string                       `json:"method" enums:"GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE"`
	Headers     map[string]FetchHeaderValues `json:"headers"`
	Body        string                       `json:"body" format:"byte"`
	Type        string                       `json:"type" enums:"Manual,Scheduled"`
}

type FetchResponse struct {
	Attempt int64 `json:"attempt"`
}

type GatewayError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}
