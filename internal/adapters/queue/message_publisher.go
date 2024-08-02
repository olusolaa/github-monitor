package queue

import (
	"github.com/streadway/amqp"
	"log"
)

type MessagePublisher interface {
	PublishMessage(queueName string, body string) error
}

type RabbitMQPublisher struct {
	manager ConnectionManager
}

func NewRabbitMQPublisher(manager ConnectionManager) *RabbitMQPublisher {
	return &RabbitMQPublisher{manager: manager}
}

func (p *RabbitMQPublisher) PublishMessage(queueName string, body string) error {
	_, err := p.manager.DeclareQueue(queueName)
	if err != nil {
		return err
	}

	err = p.manager.GetChannel().Publish(
		"",        // exchange
		queueName, // routing key
		false,     // mandatory
		false,     // immediate
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(body),
		})
	if err != nil {
		return err
	}
	log.Printf(" [x] Sent %s", body)
	return nil
}
