package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"bytes"
	"context"
	"io"
	"math"
	"net/http"
	"time"
)

type Fetcher struct {
	repo   Repository
	client *http.Client
}

func NewFetcher(repo Repository) *Fetcher {
	return &Fetcher{
		repo:   repo,
		client: &http.Client{},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, wh webhookDomain.Webhook, t responseDomain.ResponseType) error {
	resp, err := responseDomain.NewResponse(wh.ID, t)
	if err != nil {
		return err
	}
	statusCode, body, headers, duration, _ := f.doRequest(ctx, wh)
	resp.Complete(statusCode, body, headers, duration)
	return f.repo.Save(ctx, resp)
}

func (f *Fetcher) doRequest(ctx context.Context, wh webhookDomain.Webhook) (int, []byte, http.Header, time.Duration, error) {
	var lastErr error
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		attemptCtx, cancel := context.WithTimeout(ctx, wh.Timeout)
		req, err := http.NewRequestWithContext(attemptCtx, wh.Method, wh.URL.String(), bytes.NewReader(wh.Body))
		if err != nil {
			cancel()
			return 0, nil, nil, 0, err
		}
		start := time.Now()
		resp, err := f.client.Do(req)
		duration := time.Since(start)
		if err != nil {
			cancel()
			lastErr = err
			if i < maxRetries-1 {
				delay := time.Second * time.Duration(int(math.Pow(2, float64(i))))
				select {
				case <-ctx.Done():
					return 0, nil, nil, 0, ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
			break
		}
		body, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		cancel()
		if err != nil {
			return 0, nil, nil, 0, err
		}
		return resp.StatusCode, body, resp.Header, duration, nil
	}
	return 0, nil, nil, 0, lastErr
}
