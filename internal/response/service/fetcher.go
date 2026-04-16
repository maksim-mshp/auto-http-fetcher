package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"bytes"
	"context"
	"io"
	"log/slog"
	"math"
	"net/http"
	"time"
)

type Fetcher struct {
	repo   Repository
	client *http.Client
	logger *slog.Logger
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
	res, reqErr := f.doRequest(ctx, wh)
	if reqErr != nil {
		f.logger.Error("webhook processing failed",
			"webhook_id", wh.ID,
			"url", wh.URL.String(),
			"error", reqErr,
		)
	} else {
		f.logger.Info("webhook processing success",
			"webhook_id", wh.ID,
			"status", res.StatusCode,
		)
	}
	resp.Complete(res.StatusCode, res.Body, res.Headers, res.Duration)
	return f.repo.Save(ctx, resp)
}

func (f *Fetcher) doRequest(ctx context.Context, wh webhookDomain.Webhook) (*responseDomain.Response, error) {
	var lastErr error
	const maxRetries = 3
	for i := 0; i < maxRetries; i++ {
		attemptCtx, cancel := context.WithTimeout(ctx, wh.Timeout)
		req, err := http.NewRequestWithContext(attemptCtx, wh.Method, wh.URL.String(), bytes.NewReader(wh.Body))
		if err != nil {
			cancel()
			return &responseDomain.Response{}, err
		}

		req.Header = wh.Headers.Clone()

		start := time.Now()
		resp, err := f.client.Do(req)
		duration := time.Since(start)

		if err != nil {
			cancel()
			lastErr = err

			f.logger.Warn("request attempt failed",
				"webhook_id", wh.ID,
				"attempt", i+1,
				"error", err,
			)

			if i < maxRetries-1 {
				delay := time.Second * time.Duration(int(math.Pow(2, float64(i))))
				select {
				case <-ctx.Done():
					return &responseDomain.Response{}, ctx.Err()
				case <-time.After(delay):
					continue
				}
			}
			break
		}
		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		cancel()
		if err != nil {
			return &responseDomain.Response{}, err
		}
		return &responseDomain.Response{
			StatusCode: resp.StatusCode,
			Body:       body,
			Headers:    resp.Header,
			Duration:   duration,
		}, nil
	}
	return &responseDomain.Response{}, lastErr
}
