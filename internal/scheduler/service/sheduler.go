package service

import (
	"auto-http-fetcher/internal/scheduler/domain"
	webhookDomain "auto-http-fetcher/internal/webhook/domain"
	"context"
	"log/slog"
	"sync"
	"time"
)

const (
	MaxAttempts    = 3
	RetryBaseDelay = 2 * time.Second
)

type WebhookFetcher interface {
	Fetch(wh *webhookDomain.Webhook) error
}

type Scheduler struct {
	pq *PriorityQueue
	mu sync.Mutex

	webhooks          map[int]domain.SchedulerItem
	fetcher           WebhookFetcher
	updateMinNotifier chan struct{}

	logger *slog.Logger
}

func NewScheduler(logger *slog.Logger, fetchers ...WebhookFetcher) *Scheduler {
	if logger == nil {
		logger = slog.Default()
	}

	var fetcher WebhookFetcher
	if len(fetchers) > 0 {
		fetcher = fetchers[0]
	}

	return &Scheduler{
		pq:                NewPriorityQueue(),
		webhooks:          make(map[int]domain.SchedulerItem),
		fetcher:           fetcher,
		updateMinNotifier: make(chan struct{}, 1),
		logger:            logger,
	}
}

func (s *Scheduler) SetFetcher(fetcher WebhookFetcher) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.fetcher = fetcher
}

func (s *Scheduler) AddWebhook(wh *webhookDomain.Webhook) {
	s.mu.Lock()
	if _, ok := s.webhooks[wh.ID]; ok {
		s.mu.Unlock()
		return
	}

	nextRun := time.Now().Add(wh.Interval)

	err := s.pq.Add(wh.ID, nextRun, wh)
	if err == nil {
		s.webhooks[wh.ID] = domain.SchedulerItem{
			Attempt:       0,
			Webhook:       wh,
			ScheduledTime: nextRun,
		}
	}
	s.mu.Unlock()

	if err != nil {
		s.logger.Error("add webhook failed", "err", err)
		return
	}

	s.notifyMinChanged()
	s.logger.Info("added new webhook", "id", wh.ID)
}

func (s *Scheduler) UpdateWebhook(id int, newWh *webhookDomain.Webhook) {
	s.mu.Lock()
	if _, ok := s.webhooks[id]; !ok {
		s.mu.Unlock()
		return
	}

	queueItem, err := s.pq.Get(id)
	if err == nil {
		scheduledTime := queueItem.NextFetch
		if queueItem.Webhook.Interval != newWh.Interval {
			scheduledTime = time.Now().Add(newWh.Interval)
			err = s.pq.UpdateNextFetch(id, scheduledTime)
		}
		queueItem.Webhook = newWh
		s.webhooks[id] = domain.SchedulerItem{
			Attempt:       s.webhooks[id].Attempt,
			Webhook:       newWh,
			ScheduledTime: scheduledTime,
		}
	}
	s.mu.Unlock()

	if err != nil {
		s.logger.Error("update webhook failed", "id", id, "err", err)
		return
	}

	s.notifyMinChanged()
	s.logger.Info("webhook updated", "id", id)
}

func (s *Scheduler) DeleteWebhook(id int) {
	s.mu.Lock()
	if _, ok := s.webhooks[id]; !ok {
		s.mu.Unlock()
		return
	}

	_, err := s.pq.Remove(id)
	if err != nil {
		s.mu.Unlock()
		s.logger.Error("delete webhook failed", "id", id, "err", err)
		return
	}

	delete(s.webhooks, id)
	s.mu.Unlock()

	s.notifyMinChanged()

	s.logger.Info("webhook deleted", "id", id)
}

func (s *Scheduler) Retry(webhookID int) {
	s.mu.Lock()
	item, ok := s.webhooks[webhookID]
	if !ok {
		s.mu.Unlock()
		return
	}

	if item.Attempt == MaxAttempts {
		s.mu.Unlock()
		// TODO: save error to postgres
		return
	}

	delay := RetryBaseDelay * time.Duration(1<<(item.Attempt))
	nextRun := item.ScheduledTime.Add(delay)

	err := s.pq.UpdateNextFetch(webhookID, nextRun)
	if err != nil {
		s.mu.Unlock()
		s.logger.Error("retry failed", "id", webhookID, "err", err)
		return
	}

	s.webhooks[webhookID] = domain.SchedulerItem{
		Attempt:       item.Attempt + 1,
		ScheduledTime: nextRun,
		Webhook:       item.Webhook,
	}

	s.mu.Unlock()
	s.notifyMinChanged()

	s.logger.Info("webhook retried", "id", webhookID)
}

func (s *Scheduler) Work(ctx context.Context) {
	fetchCh := make(chan *webhookDomain.Webhook)

	go s.consumeFetches(ctx, fetchCh)

	var timer *time.Timer
	var timerCh <-chan time.Time

	stopTimer := func() {
		if timer == nil {
			return
		}
		if !timer.Stop() {
			select {
			case <-timer.C:
			default:
			}
		}
		timerCh = nil
	}

	resetTimer := func() {
		stopTimer()

		nextFetch, ok := s.minNextFetch()
		if !ok {
			return
		}

		delay := time.Until(nextFetch)
		if delay < 0 {
			delay = 0
		}

		timer.Reset(delay)
		timerCh = timer.C
	}

	timer = time.NewTimer(time.Hour)
	stopTimer()
	resetTimer()
	defer stopTimer()

	for {
		select {
		case <-ctx.Done():
			return
		case <-s.updateMinNotifier:
			resetTimer()
		case <-timerCh:
			wh, ok := s.scheduleDueWebhook()
			if !ok {
				resetTimer()
				continue
			}

			select {
			case fetchCh <- wh:
			case <-ctx.Done():
				return
			}

			resetTimer()
		}
	}
}

func (s *Scheduler) notifyMinChanged() {
	if s.updateMinNotifier == nil {
		return
	}

	select {
	case s.updateMinNotifier <- struct{}{}:
	default:
	}
}

func (s *Scheduler) minNextFetch() (time.Time, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	queueItem, ok := s.pq.Min()
	if !ok {
		return time.Time{}, false
	}

	return queueItem.NextFetch, true
}

func (s *Scheduler) scheduleDueWebhook() (*webhookDomain.Webhook, bool) {
	now := time.Now()

	s.mu.Lock()
	defer s.mu.Unlock()

	queueItem, ok := s.pq.Min()
	if !ok || queueItem.NextFetch.After(now) {
		return nil, false
	}

	wh := queueItem.Webhook
	nextRun := now.Add(wh.Interval)

	err := s.pq.UpdateNextFetch(queueItem.ID, nextRun)
	if err != nil {
		s.logger.Error("schedule next fetch failed", "id", queueItem.ID, "err", err)
		return nil, false
	}

	s.webhooks[queueItem.ID] = domain.SchedulerItem{
		Attempt:       0,
		ScheduledTime: nextRun,
		Webhook:       wh,
	}

	return wh, true
}

func (s *Scheduler) consumeFetches(ctx context.Context, fetchCh <-chan *webhookDomain.Webhook) {
	for {
		select {
		case <-ctx.Done():
			return
		case wh := <-fetchCh:
			go s.fetch(wh)
		}
	}
}

func (s *Scheduler) fetch(wh *webhookDomain.Webhook) {
	s.mu.Lock()
	fetcher := s.fetcher
	s.mu.Unlock()

	if fetcher == nil {
		s.logger.Error("fetcher is not configured", "id", wh.ID)
		return
	}

	if err := fetcher.Fetch(wh); err != nil {
		s.logger.Error("fetch failed", "id", wh.ID, "err", err)
	}
}
