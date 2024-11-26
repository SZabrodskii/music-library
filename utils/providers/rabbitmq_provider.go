package providers

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"os"
)

type RabbitMQProviderConfig struct {
	URL string
}

func NewRabbitMQProviderConfig() *RabbitMQProviderConfig {
	return &RabbitMQProviderConfig{
		URL: os.Getenv("RABBITMQ_URL"),
	}
}

type RabbitMQProvider struct {
	logger *zap.Logger
	conn   *amqp.Connection
	ch     *amqp.Channel
}

func NewRabbitMQProvider(logger *zap.Logger, config *RabbitMQProviderConfig) *RabbitMQProvider {
	conn, err := amqp.Dial(config.URL)
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	return &RabbitMQProvider{
		logger: logger,
		conn:   conn,
		ch:     ch,
	}
}

func (r *RabbitMQProvider) GetConnection() *amqp.Connection {
	return r.conn
}

func (r *RabbitMQProvider) GetChannel() *amqp.Channel {
	return r.ch
}

func (r *RabbitMQProvider) Publish(queueName string, body []byte) error {
	err := r.ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
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
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
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
