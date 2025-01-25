package pubsub

import (
	"fmt"

	amqp "github.com/rabbitmq/amqp091-go"
)

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType int) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}
	fmt.Printf("Declaring queue: %s\n", queueName)
	fmt.Printf("Exchange: %s, Key: %s\n", exchange, key)
	queue, err := channel.QueueDeclare(queueName, simpleQueueType == 1, simpleQueueType == 0, simpleQueueType == 0, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}
	fmt.Printf("Queue created: %+v\n", queue)
	err = channel.QueueBind(queueName, key, exchange, false, nil)
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	return channel, queue, nil
}