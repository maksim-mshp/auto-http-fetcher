package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

type Fetcher struct {
	repo        Repository
	client      *http.Client
	maxAttempts int
}

func NewFetcher(repo Repository, maxAttempts int) *Fetcher {
	return &Fetcher{
		repo:        repo,
		maxAttempts: maxAttempts,
		client:      &http.Client{},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, wh webhookDomain.Webhook, t responseDomain.ResponseType) error {
	resp, err := responseDomain.NewResponse(wh.ID, t)
	if err != nil {
		return err
	}
	for {
		statusCode, body, headers, duration, err := f.doRequest(ctx, wh)
		if err != nil {
			if !resp.IsRetryable(f.maxAttempts) {
				if err := f.repo.Save(ctx, resp); err != nil {
					return err
				}
				return err
			}
			resp.Retry()
		} else {
			resp.Complete(statusCode, body, headers, duration)
			if !resp.IsRetryable(f.maxAttempts) {
				break
			}
			resp.Retry()
		}
	}
	return f.repo.Save(ctx, resp)
}

func (f *Fetcher) doRequest(ctx context.Context, wh webhookDomain.Webhook) (int, []byte, http.Header, time.Duration, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, time.Duration(wh.Timeout))
	defer cancelCtx()
	req, err := http.NewRequestWithContext(ctx, wh.Method, wh.URL.String(), bytes.NewReader(wh.Body))
	if err != nil {
		return 0, nil, nil, 0, err
	}
	start := time.Now()
	resp, err := f.client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return 0, nil, nil, 0, err
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, 0, err
	}
	return resp.StatusCode, body, resp.Header, duration, nil
}
