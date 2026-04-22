package grpc

import (
	"auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	fetcherpb "auto-http-fetcher/proto/fetcher/v1"
	"context"
	"errors"
	"net/http"
	"net/url"
	"time"
)

func (h *Handler) Fetch(ctx context.Context, req *fetcherpb.FetchRequest) (*fetcherpb.FetchResponse, error) {
	wh := webhookDomain.Webhook{
		ID:          int(req.Id),
		Description: req.Description,
		Interval:    time.Duration(req.IntervalMs) * time.Millisecond,
		Timeout:     time.Duration(req.TimeoutMs) * time.Millisecond,
		Method:      req.Method,
		Body:        req.Body,
	}

	headers := make(http.Header)
	for key, values := range req.Headers {
		for _, value := range values.Values {
			headers.Add(key, value)
		}
	}

	wh.Headers = headers

	url, err := url.Parse(req.Url)
	if err != nil {
		return nil, err
	}
	wh.URL = *url

	whType := domain.ResponseType(req.Type)
	if !whType.IsValid() {
		return nil, errors.New("invalid webhook type")
	}
	res, err := h.fetcher.Fetch(ctx, wh, whType)
	if err != nil {
		return nil, err
	}

	resp := fetcherpb.FetchResponse{
		Attempt: int64(res.Attempt),
	}
	return &resp, nil
}
