package queue

import (
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Queue struct {
	logger *zap.Logger
	conn   *amqp.Connection
	ch     *amqp.Channel
}

func NewQueue(logger *zap.Logger) *Queue {
	conn, err := amqp.Dial(config.GetEnv("RABBITMQ_URL"))
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	ch, err := conn.Channel()
	if err != nil {
		logger.Fatal("Failed to open a channel", zap.Error(err))
	}
	return &Queue{
		logger: logger,
		conn:   conn,
		ch:     ch,
	}
}

func (q *Queue) GetConnection() *amqp.Connection {
	return q.conn
}

func (q *Queue) GetChannel() *amqp.Channel {
	return q.ch
}

func (q *Queue) Publish(queueName string, body []byte) error {
	err := q.ch.Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		})
	if err != nil {
		q.logger.Error("Failed to publish a message", zap.Error(err))
		return err
	}
	return nil
}

func (q *Queue) Consume(queueName string, consumer func(d amqp.Delivery)) error {
	msgs, err := q.ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)
	if err != nil {
		q.logger.Error("Failed to register a consumer", zap.Error(err))
		return err
	}

	go func() {
		for d := range msgs {
			consumer(d)
		}
	}()

	return nil
}
