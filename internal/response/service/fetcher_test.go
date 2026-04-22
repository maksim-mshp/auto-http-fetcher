package service

import (
	"auto-http-fetcher/internal/core/mock"
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestFetch_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			assert.Equal(t, 1, r.WebhookID)
			assert.Equal(t, 200, r.StatusCode)
			assert.Equal(t, responseDomain.SuccessStatus, r.Status)
			return nil
		},
	}

	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	webhookURL, _ := url.Parse(server.URL)
	wh := webhookDomain.Webhook{
		ID:          1,
		ModuleID:    1,
		Description: "Test webhook",
		Interval:    5 * time.Second,
		Timeout:     30 * time.Second,
		URL:         *webhookURL,
		Method:      "GET",
		Headers:     http.Header{},
		Body:        nil,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, responseDomain.ManualType)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Equal(t, responseDomain.SuccessStatus, resp.Status)
}

func TestFetch_Non200StatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`not found`))
	}))
	defer server.Close()

	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			assert.Equal(t, 404, r.StatusCode)
			assert.Equal(t, responseDomain.FailedStatus, r.Status)
			return nil
		},
	}

	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	webhookURL, _ := url.Parse(server.URL)
	wh := webhookDomain.Webhook{
		ID:          1,
		ModuleID:    1,
		Description: "Test webhook",
		Interval:    5 * time.Second,
		Timeout:     30 * time.Second,
		URL:         *webhookURL,
		Method:      "GET",
		Headers:     http.Header{},
		Body:        nil,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, responseDomain.ManualType)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, 404, resp.StatusCode)
	assert.Contains(t, err.Error(), "status code 404")
}

func TestFetch_InvalidURL(t *testing.T) {
	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			return nil
		},
	}
	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	wh := webhookDomain.Webhook{
		ID:      1,
		URL:     url.URL{},
		Method:  "GET",
		Timeout: 30 * time.Second,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, responseDomain.ManualType)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, responseDomain.FailedStatus, resp.Status)
}
func TestFetch_InvalidResponseType(t *testing.T) {
	mockRepo := &mock.MockResponseRepository{}
	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	webhookURL, _ := url.Parse("https://example.com")
	wh := webhookDomain.Webhook{
		ID:      1,
		URL:     *webhookURL,
		Method:  "GET",
		Timeout: 30 * time.Second,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, "invalid_type")

	assert.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "invalid response type")
}

func TestFetch_RepositorySaveError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			return errors.New("database connection failed")
		},
	}

	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	webhookURL, _ := url.Parse(server.URL)
	wh := webhookDomain.Webhook{
		ID:      1,
		URL:     *webhookURL,
		Method:  "GET",
		Timeout: 30 * time.Second,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, responseDomain.ManualType)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Contains(t, err.Error(), "database connection failed")
}

func TestFetch_WithHeaders(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "application/json", r.Header.Get("Content-Type"))
		assert.Equal(t, "Bearer token123", r.Header.Get("Authorization"))
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			return nil
		},
	}

	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	headers := make(http.Header)
	headers.Set("Content-Type", "application/json")
	headers.Set("Authorization", "Bearer token123")

	webhookURL, _ := url.Parse(server.URL)
	wh := webhookDomain.Webhook{
		ID:      1,
		URL:     *webhookURL,
		Method:  "POST",
		Headers: headers,
		Body:    []byte(`{"key":"value"}`),
		Timeout: 30 * time.Second,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, responseDomain.ManualType)

	assert.NoError(t, err)
	assert.NotNil(t, resp)
}

func TestFetch_ContextCanceled(t *testing.T) {
	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			return nil
		},
	}

	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	webhookURL, _ := url.Parse("https://example.com")
	wh := webhookDomain.Webhook{
		ID:      1,
		URL:     *webhookURL,
		Method:  "GET",
		Timeout: 30 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	resp, err := fetcher.Fetch(ctx, wh, responseDomain.ManualType)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, responseDomain.FailedStatus, resp.Status)
}

func TestFetch_WithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	mockRepo := &mock.MockResponseRepository{
		SaveFunc: func(ctx context.Context, r *responseDomain.Response) error {
			return nil
		},
	}

	logger := slog.Default()
	fetcher := NewFetcher(mockRepo, logger)

	webhookURL, _ := url.Parse(server.URL)
	wh := webhookDomain.Webhook{
		ID:      1,
		URL:     *webhookURL,
		Method:  "GET",
		Timeout: 50 * time.Millisecond,
	}

	resp, err := fetcher.Fetch(context.Background(), wh, responseDomain.ManualType)

	assert.Error(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, responseDomain.FailedStatus, resp.Status)
}
