package queue

import (
	"github.com/streadway/amqp"
)

type ConnectionManager interface {
	GetChannel() *amqp.Channel
	DeclareQueue(queueName string) (amqp.Queue, error)
	Close() error
}

type RabbitMQConnectionManager struct {
	connection *amqp.Connection
	channel    *amqp.Channel
}

func NewRabbitMQConnectionManager(amqpURL string) (*RabbitMQConnectionManager, error) {
	conn, err := amqp.Dial(amqpURL)
	if err != nil {
		return nil, err
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitMQConnectionManager{
		connection: conn,
		channel:    ch,
	}, nil
}

func (rm *RabbitMQConnectionManager) GetChannel() *amqp.Channel {
	return rm.channel
}

func (rm *RabbitMQConnectionManager) DeclareQueue(queueName string) (amqp.Queue, error) {
	return rm.channel.QueueDeclare(
		queueName,
		true,  // durable
		false, // delete when unused
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
}

func (rm *RabbitMQConnectionManager) Close() error {
	if err := rm.channel.Close(); err != nil {
		return err
	}
	return rm.connection.Close()
}
