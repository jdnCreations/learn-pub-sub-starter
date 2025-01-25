package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)
func main() {
	fmt.Println("Starting Peril server...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
	if err != nil {
		log.Fatalf("Could not dial, err: %v", err.Error())
	}
	defer conn.Close()

	channel, err := conn.Channel()
	if err != nil {
		log.Fatalf("error with channel on connection, err: %v", err.Error())
	}

	fmt.Println("Connection was successful")

	err = channel.ExchangeDeclare(
    routing.ExchangePerilDirect, // name of the exchange
    "direct",                    // type of exchange
    true,                        // durable
    false,                       // auto-deleted
    false,                       // internal
    false,                       // no-wait
    nil,                         // arguments
)
if err != nil {
    log.Fatalf("Failed to declare exchange: %v", err)
}

	gamelogic.PrintServerHelp()
	for {
		words := gamelogic.GetInput()
		if len(words) == 0 {
			continue
		}

		if words[0] == "pause" {
			log.Printf("sending a pause message")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: true})
			continue
		}

		if words[0] == "resume" {
			log.Printf("sending a resume message")
			pubsub.PublishJSON(channel, routing.ExchangePerilDirect, routing.PauseKey, routing.PlayingState{IsPaused: false})
			continue
		}
		if words[0] == "quit" {
			log.Printf("sending a quit message")
			break	
		}

		log.Printf("I don't understand the command")
	}


	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	<-signalChan
	fmt.Println("Program is shutting down")
}
