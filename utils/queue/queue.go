package queue

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"music-library/utils/config"
)

var logger *zap.Logger

func NewQueue() *Queue {
	logger, _ = zap.NewProduction()
	conn, err := amqp.Dial(config.GetEnv("RABBITMQ_URL"))
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	return &Queue{
		logger: logger,
		conn:   conn,
	}
}

type Queue struct {
	logger *zap.Logger
	conn   *amqp.Connection
}

func (q *Queue) GetConnection() *amqp.Connection {
	return q.conn
}
