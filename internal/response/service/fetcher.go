package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Fetcher struct {
	repo   Repository
	client *http.Client
	logger *slog.Logger
}

func NewFetcher(repo Repository, logger *slog.Logger) *Fetcher {
	return &Fetcher{
		repo:   repo,
		client: &http.Client{},
		logger: logger,
	}
}

func (f *Fetcher) Fetch(ctx context.Context, wh webhookDomain.Webhook, t responseDomain.ResponseType) (*responseDomain.Response, error) {
	resp, err := responseDomain.NewResponse(wh.ID, t)
	if err != nil {
		f.logger.Error("creating response error",
			"webhook_id", wh.ID,
			"error", err,
		)
		return nil, err
	}
	requestCtx := ctx
	cancel := func() {}
	if wh.Timeout > 0 {
		requestCtx, cancel = context.WithTimeout(ctx, wh.Timeout)
	}
	defer cancel()

	res, reqErr := f.doRequest(requestCtx, wh)
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
	if err := f.repo.Save(ctx, resp); err != nil {
		return resp, err
	}
	if reqErr != nil {
		return resp, reqErr
	}
	if resp.Status == responseDomain.FailedStatus {
		return resp, fmt.Errorf("webhook returned status code %d", resp.StatusCode)
	}
	return resp, nil
}

func (f *Fetcher) doRequest(ctx context.Context, wh webhookDomain.Webhook) (responseDomain.Response, error) {
	req, err := http.NewRequestWithContext(ctx, wh.Method, wh.URL.String(), bytes.NewReader(wh.Body))
	if err != nil {
		return responseDomain.Response{}, err
	}
	req.Header = wh.Headers.Clone()

	start := time.Now()
	resp, err := f.client.Do(req)
	duration := time.Since(start)

	if err != nil {
		return responseDomain.Response{}, err
	}

	body, err := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	if err != nil {
		return responseDomain.Response{}, err
	}
	return responseDomain.Response{
		StatusCode: resp.StatusCode,
		Body:       body,
		Headers:    resp.Header,
		Duration:   duration,
	}, nil
}
