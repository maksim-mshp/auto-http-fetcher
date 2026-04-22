package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/IBM/sarama"
)

type ProducerInterface interface {
	SendMessage(ctx context.Context, userID int, message any) error
}

type Producer struct {
	producer sarama.SyncProducer
	topic    string
}

func NewProducer(brokers []string, topic string) (*Producer, error) {
	deviceName, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	configSarama := sarama.NewConfig()
	configSarama.Producer.RequiredAcks = sarama.WaitForAll
	configSarama.Producer.Retry.Max = 10
	configSarama.Producer.Retry.Backoff = 100 * time.Millisecond
	configSarama.Producer.Flush.Bytes = 16384
	configSarama.Consumer.MaxProcessingTime = 1 * time.Second
	configSarama.Producer.Flush.Frequency = 5 * time.Millisecond
	configSarama.Producer.Return.Successes = true
	configSarama.ClientID = fmt.Sprintf("kafka-producer-%s", deviceName)
	configSarama.Version = sarama.V2_6_0_0

	producer, err := sarama.NewSyncProducer(brokers, configSarama)
	if err != nil {
		return nil, err
	}

	return &Producer{
		producer: producer,
		topic:    topic,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, userID int, message any) error {
	data, err := json.Marshal(message)
	if err != nil {
		return ErrNewWithDetails(&ErrInvalidBody, "error", err.Error())
	}

	msg := &sarama.ProducerMessage{
		Topic:     p.topic,
		Key:       sarama.StringEncoder(strconv.Itoa(userID)),
		Value:     sarama.ByteEncoder(data),
		Timestamp: time.Now(),
		Headers: []sarama.RecordHeader{
			{Key: []byte("timestamp"), Value: []byte(time.Now().String())},
		},
	}

	done := make(chan error, 1)
	go func() {
		_, _, err = p.producer.SendMessage(msg)
		if err != nil {
			done <- err
			return
		}
		done <- nil
		close(done)
	}()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case err := <-done:
		return err
	}
}

func (p *Producer) Close() error {
	if p.producer != nil {
		return p.producer.Close()
	}
	return nil
}
