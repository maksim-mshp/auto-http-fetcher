package grpc

import (
	"context"

	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	fetcher "auto-http-fetcher/proto/fetcher/v1"
)

type Fetcher struct {
	client fetcher.FetcherServiceClient
}

func NewFetcher(client fetcher.FetcherServiceClient) *Fetcher {
	return &Fetcher{
		client: client,
	}
}

func (f *Fetcher) Fetch(wh *webhookDomain.Webhook) error {
	req := &fetcher.FetchRequest{
		Id:          int64(wh.ID),
		Description: wh.Description,
		IntervalMs:  wh.Interval.Milliseconds(),
		TimeoutMs:   wh.Timeout.Milliseconds(),
		Url:         wh.URL.String(),
		Method:      wh.Method,
		Headers:     convertHeaders(wh.Headers),
		Body:        wh.Body,
	}

	_, err := f.client.Fetch(context.Background(), req)
	return err
}

func convertHeaders(headers map[string][]string) map[string]*fetcher.HeaderValues {
	if len(headers) == 0 {
		return nil
	}

	result := make(map[string]*fetcher.HeaderValues, len(headers))
	for key, values := range headers {
		result[key] = &fetcher.HeaderValues{
			Values: append([]string(nil), values...),
		}
	}

	return result
}
