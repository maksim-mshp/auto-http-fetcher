package service

import (
	"auto-http-fetcher/internal/response/domain"
	"bytes"
	"context"
	"io"
	"net/http"
	"time"
)

// потом поменяю на домен, который Влад написал
type WebhookDTO struct {
	ID      int
	URL     string
	Method  string
	Body    []byte
	Headers http.Header
	Timeout int
}

type Fetcher struct {
	repo        domain.Repository
	client      *http.Client
	maxAttempts int
}

func NewFetcher(repo domain.Repository, maxAttempts int) *Fetcher {
	return &Fetcher{
		repo:        repo,
		maxAttempts: maxAttempts,
		client:      &http.Client{},
	}
}

func (f *Fetcher) Fetch(ctx context.Context, wh WebhookDTO, t domain.ResponseType) error {
	resp, err := domain.NewResponse(wh.ID, t)
	if err != nil {
		return err
	}
	for {
		statusCode, body, headers, duration, err := f.doRequest(ctx, wh)
		if err != nil {
			if !resp.IsRetryable(f.maxAttempts) {
				f.repo.Save(ctx, resp)
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

func (f *Fetcher) doRequest(ctx context.Context, wh WebhookDTO) (int, []byte, http.Header, time.Duration, error) {
	ctx, cancelCtx := context.WithTimeout(ctx, time.Duration(wh.Timeout))
	defer cancelCtx()
	req, err := http.NewRequest(wh.Method, wh.URL, bytes.NewReader(wh.Body))
	if err != nil {
		return 0, nil, nil, 0, err
	}
	start := time.Now()
	resp, err := f.client.Do(req)
	duration := time.Since(start)
	if err != nil {
		return 0, nil, nil, 0, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, nil, nil, 0, err
	}
	return resp.StatusCode, body, resp.Header, duration, nil
}
