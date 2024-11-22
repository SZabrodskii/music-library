package queue

import (
	"github.com/SZabrodskii/music-library/utils/config"
	"github.com/streadway/amqp"
	"go.uber.org/zap"
)

type Queue struct {
	logger *zap.Logger
	conn   *amqp.Connection
}

func NewQueue(logger *zap.Logger) *Queue {
	conn, err := amqp.Dial(config.GetEnv("RABBITMQ_URL"))
	if err != nil {
		logger.Fatal("Failed to connect to RabbitMQ", zap.Error(err))
	}
	return &Queue{
		logger: logger,
		conn:   conn,
	}
}

func (q *Queue) GetConnection() *amqp.Connection {
	return q.conn
}
