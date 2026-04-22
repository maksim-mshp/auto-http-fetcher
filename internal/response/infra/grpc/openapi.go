package grpc

type FetchHeaderValues struct {
	Values []string `json:"values" example:"application/json"`
}

type FetchRequest struct {
	ID          int64                        `json:"id" example:"1"`
	Description string                       `json:"description" example:"Health check"`
	IntervalMs  int64                        `json:"intervalMs" example:"60000"`
	TimeoutMs   int64                        `json:"timeoutMs" example:"5000"`
	URL         string                       `json:"url" example:"https://example.com/health"`
	Method      string                       `json:"method" enums:"GET,HEAD,POST,PUT,PATCH,DELETE,CONNECT,OPTIONS,TRACE" example:"GET"`
	Headers     map[string]FetchHeaderValues `json:"headers"`
	Body        string                       `json:"body" format:"byte" example:""`
	Type        string                       `json:"type" enums:"Manual,Scheduled" example:"Manual"`
}

type FetchResponse struct {
	Attempt int64 `json:"attempt" example:"1"`
}

type GatewayError struct {
	Code    int    `json:"code" example:"3"`
	Message string `json:"message" example:"invalid request body"`
}
