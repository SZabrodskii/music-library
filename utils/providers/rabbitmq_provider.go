package providers

import (
	"fmt"
	"github.com/SZabrodskii/music-library/utils"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQProviderConfig struct {
	URL string
}

func NewRabbitMQProviderConfig() *RabbitMQProviderConfig {
	return &RabbitMQProviderConfig{
		URL: utils.GetEnv("RABBITMQ_URL", "amqp://guest:guest@rabbitmq:5672/"),
	}
}

type RabbitMQProvider struct {
	logger *zap.Logger
	conn   *amqp.Connection
	ch     *amqp.Channel
}

func NewRabbitMQProvider(logger *zap.Logger, config *RabbitMQProviderConfig) (*RabbitMQProvider, error) {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}
	return &RabbitMQProvider{
		logger: logger,
		conn:   conn,
		ch:     ch,
	}, nil
}

func (r *RabbitMQProvider) GetConnection() *amqp.Connection {
	return r.conn
}

func (r *RabbitMQProvider) GetChannel() *amqp.Channel {
	return r.ch
}

func (r *RabbitMQProvider) Publish(queueName string, body []byte) error {
	err := r.ch.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		r.logger.Error("Failed to publish a message", zap.Error(err))
		return err
	}
	return nil
}

func (r *RabbitMQProvider) Consume(queueName string, consumer func(d amqp.Delivery)) error {
	msgs, err := r.ch.Consume(
		queueName,
		"",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		r.logger.Error("Failed to register a consumer", zap.Error(err))
		return err
	}

	go func() {
		for d := range msgs {
			consumer(d)
		}
	}()

	return nil
}
