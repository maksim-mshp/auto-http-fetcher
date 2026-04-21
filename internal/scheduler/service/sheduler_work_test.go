package service

import (
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"context"
	"io"
	"log/slog"
	"testing"
	"time"
)

type recordingFetcher struct {
	ids chan int
}

func newRecordingFetcher() *recordingFetcher {
	return &recordingFetcher{
		ids: make(chan int, 16),
	}
}

func (f *recordingFetcher) Fetch(wh *webhookDomain.Webhook) error {
	f.ids <- wh.ID
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

func testScheduledWebhook(id int, interval time.Duration) *webhookDomain.Webhook {
	return &webhookDomain.Webhook{
		ID:       id,
		Interval: interval,
	}
}

func testLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}
