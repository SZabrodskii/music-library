package queue

import (
	"github.com/streadway/amqp"
	"go.uber.org/zap"
	"music-library/config"
)

var logger *zap.Logger

func init() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
}
func InitQueue() (*amqp.Connection, error) {
	conn, err := amqp.Dial(config.GetEnv("RABBITMQ_URL"))
	if err != nil {
		logger.Error("Failed to connect to RabbitMQ", zap.Error(err))
		return nil, err
	}

	return conn, nil
}
