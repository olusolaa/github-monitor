package queue

import (
	"log"
)

type MessageConsumer interface {
	ConsumeMessages(queueName string, handleMessage func(msg []byte) error) error
}

type RabbitMQConsumer struct {
	manager ConnectionManager
}

func NewRabbitMQConsumer(manager ConnectionManager) *RabbitMQConsumer {
	return &RabbitMQConsumer{manager: manager}
}

func (c *RabbitMQConsumer) ConsumeMessages(queueName string, handleMessage func(msg []byte) error) error {
	_, err := c.manager.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	msgs, err := c.manager.GetChannel().Consume(
		queueName,
		"",    // consumers
		true,  // auto-ack
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return err
	}

	go func() {
		for d := range msgs {
			if err := handleMessage(d.Body); err != nil {
				log.Printf("Error handling message: %s", err)
			}
		}
	}()
	return nil
}
