package providers

import (
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type RabbitMQProvider struct {
	Logger *zap.Logger
	Conn   *amqp.Connection
	Ch     *amqp.Channel
}

func NewRabbitMQProvider(logger *zap.Logger) *RabbitMQProvider {
	conn, err := amqp.Dial(config.GetEnv("RABBITMQ_URL"))
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	return &RabbitMQProvider{
		Logger: logger,
		Conn:   conn,
		Ch:     ch,
	}
}

func (r *RabbitMQProvider) GetConnection() *amqp.Connection {
	return r.Conn
}

func (r *RabbitMQProvider) GetChannel() *amqp.Channel {
	return r.Ch
}

func (r *RabbitMQProvider) Publish(queueName string, body []byte) error {
	err := r.Ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		r.Logger.Error("Failed to publish a message", zap.Error(err))
		return err
	}
	return nil
}

func (r *RabbitMQProvider) Consume(queueName string, consumer func(d amqp.Delivery)) error {
	msgs, err := r.Ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		r.Logger.Error("Failed to register a consumer", zap.Error(err))
		return err
	}

	go func() {
		for d := range msgs {
			consumer(d)
		}
	}()

	return nil
}
