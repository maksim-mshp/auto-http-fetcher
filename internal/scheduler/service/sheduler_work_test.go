package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"context"
	"errors"
	"io"
	"log/slog"
	"testing"
	"time"
)

type recordingFetcher struct {
	ids chan int
	err error
}

func newRecordingFetcher() *recordingFetcher {
	return &recordingFetcher{
		ids: make(chan int, 16),
	}
}

func (f *recordingFetcher) Fetch(wh *webhookDomain.Webhook) error {
	f.ids <- wh.ID
	return f.err
}

type recordingResponseSaver struct {
	responses chan *responseDomain.Response
}

func newRecordingResponseSaver() *recordingResponseSaver {
	return &recordingResponseSaver{
		responses: make(chan *responseDomain.Response, 1),
	}
}

func (s *recordingResponseSaver) Save(_ context.Context, response *responseDomain.Response) error {
	s.responses <- response
	return nil
}

func TestSchedulerWorkFetchesWebhookWhenMinimumTimeArrives(t *testing.T) {
	fetcher := newRecordingFetcher()
	scheduler := NewScheduler(testLogger(), fetcher)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scheduler.Work(ctx)
	scheduler.AddWebhook(testScheduledWebhook(1, 20*time.Millisecond))

	if got := waitFetchedID(t, fetcher.ids); got != 1 {
		t.Fatalf("fetched webhook id = %d, want 1", got)
	}
}

func TestSchedulerWorkResetsTimerWhenMinimumChanges(t *testing.T) {
	fetcher := newRecordingFetcher()
	scheduler := NewScheduler(testLogger(), fetcher)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scheduler.Work(ctx)
	scheduler.AddWebhook(testScheduledWebhook(1, 300*time.Millisecond))
	time.Sleep(20 * time.Millisecond)
	scheduler.AddWebhook(testScheduledWebhook(2, 20*time.Millisecond))

	if got := waitFetchedID(t, fetcher.ids); got != 2 {
		t.Fatalf("first fetched webhook id = %d, want 2", got)
	}
}

func TestSchedulerWorkResetsTimerWhenUpdatedWebhookBecomesMinimum(t *testing.T) {
	fetcher := newRecordingFetcher()
	scheduler := NewScheduler(testLogger(), fetcher)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go scheduler.Work(ctx)
	scheduler.AddWebhook(testScheduledWebhook(1, 300*time.Millisecond))
	scheduler.AddWebhook(testScheduledWebhook(2, 400*time.Millisecond))
	time.Sleep(20 * time.Millisecond)
	scheduler.UpdateWebhook(2, testScheduledWebhook(2, 20*time.Millisecond))

	if got := waitFetchedID(t, fetcher.ids); got != 2 {
		t.Fatalf("first fetched webhook id = %d, want 2", got)
	}
}

func TestSchedulerRetrySavesFailedResponseWhenMaxAttemptsExceeded(t *testing.T) {
	saver := newRecordingResponseSaver()
	scheduler := NewScheduler(testLogger())
	scheduler.SetResponseSaver(saver)
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))

	scheduler.mu.Lock()
	item := scheduler.webhooks[1]
	item.Attempt = MaxAttempts
	scheduler.webhooks[1] = item
	scheduler.mu.Unlock()

	scheduler.Retry(1)

	response := waitSavedResponse(t, saver.responses)
	if response.WebhookID != 1 {
		t.Fatalf("saved WebhookID = %d, want 1", response.WebhookID)
	}
	if response.Type != responseDomain.ScheduledType {
		t.Fatalf("saved Type = %s, want %s", response.Type, responseDomain.ScheduledType)
	}
	if response.Status != responseDomain.FailedStatus {
		t.Fatalf("saved Status = %s, want %s", response.Status, responseDomain.FailedStatus)
	}
	if string(response.Body) != maxAttemptsExceededResponse {
		t.Fatalf("saved Body = %q, want %q", string(response.Body), maxAttemptsExceededResponse)
	}
	if response.Attempt != MaxAttempts {
		t.Fatalf("saved Attempt = %d, want %d", response.Attempt, MaxAttempts)
	}
	if response.FinishedAt == nil {
		t.Fatal("saved FinishedAt = nil, want non-nil")
	}
}

func TestSchedulerFetchRetriesFailedFetch(t *testing.T) {
	fetcher := newRecordingFetcher()
	fetcher.err = errors.New("fetch failed")
	saver := newRecordingResponseSaver()
	scheduler := NewScheduler(testLogger(), fetcher)
	scheduler.SetResponseSaver(saver)
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))

	scheduler.mu.Lock()
	item := scheduler.webhooks[1]
	item.Attempt = MaxAttempts
	scheduler.webhooks[1] = item
	scheduler.mu.Unlock()

	scheduler.fetch(testScheduledWebhook(1, time.Hour))

	response := waitSavedResponse(t, saver.responses)
	if response.Status != responseDomain.FailedStatus {
		t.Fatalf("saved Status = %s, want %s", response.Status, responseDomain.FailedStatus)
	}
}

func waitFetchedID(t *testing.T, ids <-chan int) int {
	t.Helper()

	select {
	case id := <-ids:
		return id
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for fetched webhook")
		return 0
	}
}

func waitSavedResponse(t *testing.T, responses <-chan *responseDomain.Response) *responseDomain.Response {
	t.Helper()

	select {
	case response := <-responses:
		return response
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for saved response")
		return nil
	}
}

func testScheduledWebhook(id int, interval time.Duration) *webhookDomain.Webhook {
	return &webhookDomain.Webhook{
		ID:       id,
		Interval: interval,
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
