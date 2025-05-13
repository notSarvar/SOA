package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"promoservice/event-service/internal/config"

	"github.com/Shopify/sarama"
	"github.com/sirupsen/logrus"
)

type Producer struct {
	producer sarama.SyncProducer
	logger   *logrus.Logger
	config   *config.ProducerConfig
	dlq      *DeadLetterQueue
}

type DeadLetterQueue struct {
	producer sarama.SyncProducer
	topic    string
	logger   *logrus.Logger
	config   *config.DeadLetterConfig
}

func NewProducer(brokers []string, cfg *config.ProducerConfig, dlqCfg *config.DeadLetterConfig, logger *logrus.Logger) (*Producer, error) {
	config := sarama.NewConfig()
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Retry.Max = cfg.Retries
	config.Producer.Retry.Backoff = cfg.RetryBackoff
	config.Producer.Return.Successes = true
	config.Net.MaxOpenRequests = cfg.MaxInFlight

	producer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create producer: %w", err)
	}

	// Инициализация Dead Letter Queue
	dlqProducer, err := sarama.NewSyncProducer(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create DLQ producer: %w", err)
	}

	dlq := &DeadLetterQueue{
		producer: dlqProducer,
		topic:    dlqCfg.Topic,
		logger:   logger,
		config:   dlqCfg,
	}

	return &Producer{
		producer: producer,
		logger:   logger,
		config:   cfg,
		dlq:      dlq,
	}, nil
}

func (p *Producer) SendMessage(ctx context.Context, topic string, key string, value interface{}) error {
	jsonValue, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonValue),
	}

	var lastErr error
	for i := 0; i < p.config.Retries; i++ {
		partition, offset, err := p.producer.SendMessage(msg)
		if err == nil {
			p.logger.WithFields(logrus.Fields{
				"topic":     topic,
				"partition": partition,
				"offset":    offset,
			}).Debug("Message sent successfully")
			return nil
		}

		lastErr = err
		p.logger.WithError(err).WithFields(logrus.Fields{
			"topic": topic,
			"retry": i + 1,
		}).Warn("Failed to send message, retrying...")

		time.Sleep(p.config.RetryBackoff)
	}

	// Если все попытки не удались, отправляем в DLQ
	if err := p.sendToDLQ(ctx, topic, key, value, lastErr); err != nil {
		p.logger.WithError(err).Error("Failed to send message to DLQ")
	}

	return fmt.Errorf("failed to send message after %d retries: %w", p.config.Retries, lastErr)
}

func (p *Producer) sendToDLQ(ctx context.Context, originalTopic, key string, value interface{}, err error) error {
	dlqMsg := struct {
		OriginalTopic string      `json:"original_topic"`
		Key           string      `json:"key"`
		Value         interface{} `json:"value"`
		Error         string      `json:"error"`
		Timestamp     time.Time   `json:"timestamp"`
	}{
		OriginalTopic: originalTopic,
		Key:           key,
		Value:         value,
		Error:         err.Error(),
		Timestamp:     time.Now(),
	}

	jsonValue, err := json.Marshal(dlqMsg)
	if err != nil {
		return fmt.Errorf("failed to marshal DLQ message: %w", err)
	}

	msg := &sarama.ProducerMessage{
		Topic: p.dlq.topic,
		Key:   sarama.StringEncoder(key),
		Value: sarama.ByteEncoder(jsonValue),
	}

	_, _, err = p.dlq.producer.SendMessage(msg)
	if err != nil {
		return fmt.Errorf("failed to send message to DLQ: %w", err)
	}

	p.logger.WithFields(logrus.Fields{
		"original_topic": originalTopic,
		"dlq_topic":      p.dlq.topic,
	}).Info("Message sent to DLQ")

	return nil
}

func (p *Producer) Close() error {
	if err := p.producer.Close(); err != nil {
		return fmt.Errorf("failed to close producer: %w", err)
	}
	if err := p.dlq.producer.Close(); err != nil {
		return fmt.Errorf("failed to close DLQ producer: %w", err)
	}
	return nil
}
