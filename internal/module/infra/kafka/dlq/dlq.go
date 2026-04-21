package dlq

import (
	"auto-http-fetcher/internal/core/kafka"
	kafka2 "auto-http-fetcher/internal/module/infra/kafka"

	"context"
	"log"
	"log/slog"
	"sync"
	"time"
)

type Message struct {
	UserID     int
	MessageDTO kafka2.WebhookKafkaDTO
	Error      error
}

type DeadLetterQueue struct {
	mu     *sync.Mutex
	queue  []*Message
	kafka  *kafka.Producer
	logger *slog.Logger
	stopCh chan struct{}
	wg     *sync.WaitGroup
}

func NewDeadLetterQueue(logger *slog.Logger, kafka *kafka.Producer) *DeadLetterQueue {
	dlq := &DeadLetterQueue{
		mu:     new(sync.Mutex),
		queue:  []*Message{},
		kafka:  kafka,
		logger: logger,
		stopCh: make(chan struct{}),
		wg:     new(sync.WaitGroup),
	}

	dlq.wg.Add(1)
	go dlq.worker()
	return dlq
}

func (dlq *DeadLetterQueue) Push(userID int, msg kafka2.WebhookKafkaDTO, error error) {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()

	dlq.queue = append(dlq.queue, &Message{
		UserID:     userID,
		MessageDTO: msg,
		Error:      error,
	})
	dlq.logger.Warn("Pushing message to dead-letter queue", "webhook_id", msg.ID, "user_id", userID)
	log.Println(dlq.queue)
}

func (dlq *DeadLetterQueue) worker() {
	defer dlq.wg.Done()

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-dlq.stopCh:
			return
		case <-ticker.C:
			dlq.processMessage()
		}
	}
}

func (dlq *DeadLetterQueue) processMessage() {
	dlq.mu.Lock()

	if len(dlq.queue) == 0 {
		dlq.mu.Unlock()
		return
	}

	toProcess := make([]*Message, len(dlq.queue))
	copy(toProcess, dlq.queue)
	dlq.queue = make([]*Message, 0)
	dlq.mu.Unlock()

	var remaining []*Message

	for _, msg := range toProcess {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		if err := dlq.kafka.SendMessage(ctx, msg.UserID, msg.MessageDTO); err != nil {
			dlq.logger.Error("dead letter queue retry error", "user_id", msg.UserID,
				"webhook_id", msg.MessageDTO.ID,
				"error", err)
			remaining = append(remaining, msg)
		} else {
			dlq.logger.Info("dead letter queue retry success", "user_id", msg.UserID,
				"webhook_id", msg.MessageDTO.ID,
				"error", err)
		}
		cancel()
	}

	if len(remaining) > 0 {
		dlq.mu.Lock()
		dlq.queue = append(remaining, dlq.queue...)
		dlq.mu.Unlock()
	}
}

func (dlq *DeadLetterQueue) Stop() {
	close(dlq.stopCh)
	dlq.wg.Wait()
	dlq.logger.Info("DLQ stopped")
}

func (dlq *DeadLetterQueue) Size() int {
	dlq.mu.Lock()
	defer dlq.mu.Unlock()
	return len(dlq.queue)
}
