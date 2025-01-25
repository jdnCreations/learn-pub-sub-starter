package pubsub

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"fmt"
	"log"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Acktype int
const (
  Ack Acktype = iota
  NackRequeue
  NackDiscard
)

type SimpleQueueType int
const (
	SimpleQueueDurable SimpleQueueType = iota
	SimpleQueueTransient
)

func SubscribeJSON[T any] (
	conn *amqp.Connection, 
	exchange, 
	queueName, 
	key string, 
	simpleQueueType SimpleQueueType, 
	handler func(T) Acktype,
	) error {
  channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
  if err != nil {
    return err
  }
  // channel.Qos(0, 10, false)
  deliveryCh, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
  if err != nil {
    return err
  }

  go func () {
		defer channel.Close()
    for delivery := range deliveryCh {
      var message T
      err := json.Unmarshal(delivery.Body, &message)
      if err != nil {
        log.Fatal(err)
      }
      ackType := handler(message)
      // removes message from queue
      switch ackType {
      case Ack:
        err = delivery.Ack(false)
        log.Printf("ACK: %v", message)
      case NackRequeue:
        err = delivery.Nack(false, true)
        log.Printf("NACK REQUEUE: %v", message)
      case NackDiscard:
        err = delivery.Nack(false, false)
        log.Printf("NACK DISCARD: %v", message)
      }
      if err != nil {
        log.Fatal(err)
      }
    }
  }()
  return nil
}

func SubscribeGob[T any](
  conn *amqp.Connection,
  exchange,
  queueName,
  key string,
  simpleQueueType SimpleQueueType,
  handler func(T) Acktype,
) error {
   channel, queue, err := DeclareAndBind(conn, exchange, queueName, key, simpleQueueType)
  if err != nil {
    return err
  }
  err = channel.Qos(10, 0, false)
  if err != nil {
    return fmt.Errorf("failed to set QoS: %w", err)
  }
  deliveryCh, err := channel.Consume(queue.Name, "", false, false, false, false, nil)
  if err != nil {
    return err
  }

  go func () {
		defer channel.Close()
    for delivery := range deliveryCh {
      var message T
      buffer := bytes.NewBuffer(delivery.Body)
      dec := gob.NewDecoder(buffer)
      err := dec.Decode(&message) 
      if err != nil {
        log.Printf("failed to decode message: %v", err)
        continue
      }
      ackType := handler(message)
      // removes message from queue
      switch ackType {
      case Ack:
        err = delivery.Ack(false)
        log.Printf("ACK: %v", message)
      case NackRequeue:
        err = delivery.Nack(false, true)
        log.Printf("NACK REQUEUE: %v", message)
      case NackDiscard:
        err = delivery.Nack(false, false)
        log.Printf("NACK DISCARD: %v", message)
      }
      if err != nil {
        log.Println(err)
      }
    }
  }()
  return nil
}

func DeclareAndBind(conn *amqp.Connection, exchange, queueName, key string, simpleQueueType SimpleQueueType) (*amqp.Channel, amqp.Queue, error) {
	channel, err := conn.Channel()
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}

	queue, err := channel.QueueDeclare(
    queueName, 
    simpleQueueType == SimpleQueueDurable, 
    simpleQueueType != SimpleQueueDurable, 
    simpleQueueType != SimpleQueueDurable, 
    false, 
    amqp.Table{
      "x-dead-letter-exchange": "peril_dlx",
    })
	if err != nil {
		return &amqp.Channel{}, amqp.Queue{}, err
	}
	fmt.Printf("Queue created: %+v\n", queue)
	err = channel.QueueBind(
    queue.Name,
    key, 
    exchange, 
    false, 
    nil,
  )
	if err != nil {
		return nil, amqp.Queue{}, fmt.Errorf("could not bind queue: %v", err) 
	}

	return channel, queue, nil
}