package service

import (
	responseDomain "auto-http-fetcher/internal/response/domain"
	schedulerDomain "auto-http-fetcher/internal/scheduler/domain"
	"context"
	"errors"
	"testing"
	"time"
)

type failingResponseSaver struct {
	responses chan *responseDomain.Response
	err       error
}

func newFailingResponseSaver(err error) *failingResponseSaver {
	return &failingResponseSaver{
		responses: make(chan *responseDomain.Response, 1),
		err:       err,
	}
}

func (s *failingResponseSaver) Save(_ context.Context, response *responseDomain.Response) error {
	s.responses <- response
	return s.err
}

func TestNewSchedulerUsesDefaultLoggerWhenNil(t *testing.T) {
	scheduler := NewScheduler(nil)

	if scheduler.logger == nil {
		t.Fatal("logger = nil, want default logger")
	}
}

func TestSchedulerSetFetcherIsUsedByFetch(t *testing.T) {
	fetcher := newRecordingFetcher()
	scheduler := NewScheduler(testLogger())
	scheduler.SetFetcher(fetcher)
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))

	scheduler.mu.Lock()
	item := scheduler.webhooks[1]
	item.Attempt = 2
	scheduler.webhooks[1] = item
	scheduler.mu.Unlock()

	scheduler.fetch(testScheduledWebhook(1, time.Hour))

	if got := waitFetchedID(t, fetcher.ids); got != 1 {
		t.Fatalf("fetched webhook id = %d, want 1", got)
	}

	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	if got := scheduler.webhooks[1].Attempt; got != 0 {
		t.Fatalf("attempt after successful fetch = %d, want 0", got)
	}
}

func TestSchedulerFetchWithoutFetcherReturns(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	scheduler.fetch(testScheduledWebhook(1, time.Hour))
}

func TestSchedulerUpsertAddsNewWebhook(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	wh := testScheduledWebhook(1, time.Hour)
	before := time.Now()

	scheduler.UpsertWebhook(wh)

	scheduler.mu.Lock()
	item, ok := scheduler.webhooks[1]
	scheduler.mu.Unlock()
	if !ok {
		t.Fatal("webhook was not added")
	}
	if item.Webhook != wh {
		t.Fatalf("stored webhook = %p, want %p", item.Webhook, wh)
	}
	assertTimeNear(t, item.ScheduledTime, before.Add(time.Hour), time.Second)
	assertNotifierHasSignal(t, scheduler)
}

func TestSchedulerUpsertUpdatesExistingWebhook(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))
	drainNotifier(scheduler)

	scheduler.mu.Lock()
	item := scheduler.webhooks[1]
	item.Attempt = 2
	scheduler.webhooks[1] = item
	scheduler.mu.Unlock()

	newWh := testScheduledWebhook(1, 2*time.Hour)
	before := time.Now()
	scheduler.UpsertWebhook(newWh)

	scheduler.mu.Lock()
	updated := scheduler.webhooks[1]
	scheduler.mu.Unlock()
	if updated.Webhook != newWh {
		t.Fatalf("stored webhook = %p, want %p", updated.Webhook, newWh)
	}
	if updated.Attempt != 2 {
		t.Fatalf("attempt = %d, want 2", updated.Attempt)
	}
	assertTimeNear(t, updated.ScheduledTime, before.Add(2*time.Hour), time.Second)
	assertNotifierHasSignal(t, scheduler)
}

func TestSchedulerAddWebhookIgnoresDuplicateID(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	first := testScheduledWebhook(1, time.Hour)
	second := testScheduledWebhook(1, time.Minute)

	scheduler.AddWebhook(first)
	drainNotifier(scheduler)
	scheduler.AddWebhook(second)

	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	if got := len(scheduler.webhooks); got != 1 {
		t.Fatalf("webhooks len = %d, want 1", got)
	}
	if got := scheduler.webhooks[1].Webhook; got != first {
		t.Fatalf("stored webhook = %p, want first webhook %p", got, first)
	}
	if got := scheduler.pq.Len(); got != 1 {
		t.Fatalf("queue len = %d, want 1", got)
	}
}

func TestSchedulerAddWebhookLogsQueueError(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	wh := testScheduledWebhook(1, time.Hour)
	if err := scheduler.pq.Add(1, time.Now().Add(time.Hour), wh); err != nil {
		t.Fatalf("preload queue error = %v", err)
	}

	scheduler.AddWebhook(wh)

	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	if _, ok := scheduler.webhooks[1]; ok {
		t.Fatal("webhook was stored after queue add error")
	}
}

func TestSchedulerUpdateWebhookIgnoresUnknownID(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	scheduler.UpdateWebhook(1, testScheduledWebhook(1, time.Hour))

	scheduler.mu.Lock()
	defer scheduler.mu.Unlock()
	if got := len(scheduler.webhooks); got != 0 {
		t.Fatalf("webhooks len = %d, want 0", got)
	}
}

func TestSchedulerUpdateWebhookKeepsScheduleWhenIntervalIsUnchanged(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))
	drainNotifier(scheduler)

	queueItem, err := scheduler.pq.Get(1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	originalSchedule := queueItem.NextFetch
	newWh := testScheduledWebhook(1, time.Hour)

	scheduler.UpdateWebhook(1, newWh)

	scheduler.mu.Lock()
	updated := scheduler.webhooks[1]
	scheduler.mu.Unlock()
	if updated.Webhook != newWh {
		t.Fatalf("stored webhook = %p, want %p", updated.Webhook, newWh)
	}
	if !updated.ScheduledTime.Equal(originalSchedule) {
		t.Fatalf("scheduled time = %v, want %v", updated.ScheduledTime, originalSchedule)
	}
	assertNotifierHasSignal(t, scheduler)
}

func TestSchedulerUpdateWebhookLogsQueueError(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	wh := testScheduledWebhook(1, time.Hour)
	scheduler.webhooks[1] = schedulerDomain.SchedulerItem{Webhook: wh}

	scheduler.UpdateWebhook(1, wh)
}

func TestSchedulerDeleteWebhookRemovesExistingWebhook(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))
	drainNotifier(scheduler)

	scheduler.DeleteWebhook(1)

	scheduler.mu.Lock()
	_, exists := scheduler.webhooks[1]
	scheduler.mu.Unlock()
	if exists {
		t.Fatal("webhook still exists after delete")
	}
	if got := scheduler.pq.Len(); got != 0 {
		t.Fatalf("queue len = %d, want 0", got)
	}
	assertNotifierHasSignal(t, scheduler)
}

func TestSchedulerDeleteWebhookIgnoresUnknownID(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	scheduler.DeleteWebhook(1)

	if got := scheduler.pq.Len(); got != 0 {
		t.Fatalf("queue len = %d, want 0", got)
	}
}

func TestSchedulerDeleteWebhookLogsQueueError(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.webhooks[1] = schedulerDomain.SchedulerItem{Webhook: testScheduledWebhook(1, time.Hour)}

	scheduler.DeleteWebhook(1)
}

func TestSchedulerRetryIgnoresUnknownWebhook(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	scheduler.Retry(1)
}

func TestSchedulerRetrySchedulesExponentialDelay(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))
	drainNotifier(scheduler)
	before := time.Now()

	scheduler.Retry(1)

	scheduler.mu.Lock()
	item := scheduler.webhooks[1]
	scheduler.mu.Unlock()
	if item.Attempt != 1 {
		t.Fatalf("attempt = %d, want 1", item.Attempt)
	}
	assertTimeNear(t, item.ScheduledTime, before.Add(RetryBaseDelay), time.Second)

	queueItem, err := scheduler.pq.Get(1)
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}
	if !queueItem.NextFetch.Equal(item.ScheduledTime) {
		t.Fatalf("queue next fetch = %v, want %v", queueItem.NextFetch, item.ScheduledTime)
	}
	assertNotifierHasSignal(t, scheduler)
}

func TestSchedulerRetryLogsQueueError(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.webhooks[1] = schedulerDomain.SchedulerItem{
		Attempt: 1,
		Webhook: testScheduledWebhook(1, time.Hour),
	}

	scheduler.Retry(1)
}

func TestSchedulerRetryMaxAttemptsLogsQueueError(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.webhooks[1] = schedulerDomain.SchedulerItem{
		Attempt: MaxAttempts,
		Webhook: testScheduledWebhook(1, time.Hour),
	}

	scheduler.Retry(1)
}

func TestSchedulerMinNextFetchReturnsFalseOnEmptyQueue(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	nextFetch, ok := scheduler.minNextFetch()
	if ok {
		t.Fatalf("ok = true, want false with next fetch %v", nextFetch)
	}
}

func TestSchedulerScheduleDueWebhookReturnsFalseOnEmptyQueue(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	wh, ok := scheduler.scheduleDueWebhook()
	if ok || wh != nil {
		t.Fatalf("scheduleDueWebhook() = (%v, %v), want (nil, false)", wh, ok)
	}
}

func TestSchedulerScheduleDueWebhookReturnsFalseForFutureWebhook(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.AddWebhook(testScheduledWebhook(1, time.Hour))

	wh, ok := scheduler.scheduleDueWebhook()
	if ok || wh != nil {
		t.Fatalf("scheduleDueWebhook() = (%v, %v), want (nil, false)", wh, ok)
	}
}

func TestSchedulerScheduleDueWebhookReturnsFalseWhenMapItemIsMissing(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	wh := testScheduledWebhook(1, time.Hour)
	if err := scheduler.pq.Add(1, time.Now().Add(-time.Second), wh); err != nil {
		t.Fatalf("Add() error = %v", err)
	}

	got, ok := scheduler.scheduleDueWebhook()
	if ok || got != nil {
		t.Fatalf("scheduleDueWebhook() = (%v, %v), want (nil, false)", got, ok)
	}
}

func TestSchedulerScheduleDueWebhookUpdatesNextRun(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	wh := testScheduledWebhook(1, time.Hour)
	scheduler.AddWebhook(wh)
	if err := scheduler.pq.UpdateNextFetch(1, time.Now().Add(-time.Second)); err != nil {
		t.Fatalf("UpdateNextFetch() error = %v", err)
	}

	before := time.Now()
	got, ok := scheduler.scheduleDueWebhook()

	if !ok {
		t.Fatal("ok = false, want true")
	}
	if got != wh {
		t.Fatalf("webhook = %p, want %p", got, wh)
	}

	scheduler.mu.Lock()
	item := scheduler.webhooks[1]
	scheduler.mu.Unlock()
	assertTimeNear(t, item.ScheduledTime, before.Add(time.Hour), time.Second)
}

func TestSchedulerMarkFetchSucceededIgnoresUnknownWebhook(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	scheduler.markFetchSucceeded(1)
}

func TestSchedulerMarkFetchSucceededLogsQueueError(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.webhooks[1] = schedulerDomain.SchedulerItem{
		Attempt: 1,
		Webhook: testScheduledWebhook(1, time.Hour),
	}

	scheduler.markFetchSucceeded(1)
}

func TestSchedulerSaveMaxAttemptsExceededWithoutSaverReturns(t *testing.T) {
	scheduler := NewScheduler(testLogger())

	scheduler.saveMaxAttemptsExceeded(1, MaxAttempts)
}

func TestSchedulerSaveMaxAttemptsExceededLogsSaveError(t *testing.T) {
	saveErr := errors.New("save failed")
	saver := newFailingResponseSaver(saveErr)
	scheduler := NewScheduler(testLogger())
	scheduler.SetResponseSaver(saver)

	scheduler.saveMaxAttemptsExceeded(1, MaxAttempts)

	response := waitSavedResponse(t, saver.responses)
	if response.WebhookID != 1 {
		t.Fatalf("saved WebhookID = %d, want 1", response.WebhookID)
	}
}

func TestSchedulerNotifyMinChangedReturnsWhenChannelIsNil(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.updateMinNotifier = nil

	scheduler.notifyMinChanged()
}

func TestSchedulerNotifyMinChangedDoesNotBlockWhenChannelIsFull(t *testing.T) {
	scheduler := NewScheduler(testLogger())
	scheduler.notifyMinChanged()
	scheduler.notifyMinChanged()

	assertNotifierHasSignal(t, scheduler)
}

func assertNotifierHasSignal(t *testing.T, scheduler *Scheduler) {
	t.Helper()

	select {
	case <-scheduler.updateMinNotifier:
	default:
		t.Fatal("expected scheduler notification")
	}
}

func drainNotifier(scheduler *Scheduler) {
	select {
	case <-scheduler.updateMinNotifier:
	default:
	}
}

func assertTimeNear(t *testing.T, got time.Time, want time.Time, tolerance time.Duration) {
	t.Helper()

	diff := got.Sub(want)
	if diff < 0 {
		diff = -diff
	}
	if diff > tolerance {
		t.Fatalf("time = %v, want within %v of %v", got, tolerance, want)
	}
}
