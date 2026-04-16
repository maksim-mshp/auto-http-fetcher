package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"bytes"
	"context"
	"fmt"
	"io"
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
	statusCode, body, headers, duration, reqErr := f.doRequest(ctx, wh)
	resp.Complete(statusCode, body, headers, duration)
	if saveErr := f.repo.Save(ctx, resp); saveErr != nil {
		if reqErr != nil {
			return fmt.Errorf("doRequest error: %w, Save error: %w", reqErr, saveErr)
		}
		return saveErr
	}
	return nil
}

func (f *Fetcher) doRequest(ctx context.Context, wh webhookDomain.Webhook) (int, []byte, http.Header, time.Duration, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, wh.Timeout)
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
