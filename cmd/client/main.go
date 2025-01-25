package main

import (
	"fmt"
	"log"

	"github.com/bootdotdev/learn-pub-sub-starter/internal/gamelogic"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/pubsub"
	"github.com/bootdotdev/learn-pub-sub-starter/internal/routing"
	amqp "github.com/rabbitmq/amqp091-go"
)

func main() {
	fmt.Println("Starting Peril client...")
	connStr := "amqp://guest:guest@localhost:5672/"
	conn, err := amqp.Dial(connStr)
		if err != nil {
		log.Fatalf("Could not dial, err: %v", err.Error())
	}
	defer conn.Close()
	name, err := gamelogic.ClientWelcome()
	if err != nil {
		log.Fatalf("Could not get name from client, err: %v", err.Error())
	}
	puser := fmt.Sprintf("%s.%s", routing.PauseKey, name)
	_, _, err = pubsub.DeclareAndBind(conn, routing.ExchangePerilDirect, puser, routing.PauseKey, 0)
	if err != nil {
		log.Fatalf("could not declare and bind, err: %v", err.Error())
	}

  gameState := gamelogic.NewGameState(name)
  gamelogic.PrintClientHelp()
  for {
    words := gamelogic.GetInput()
    if len(words) == 0 {
      continue
    }

    if words[0] == "spawn" {
      err := gameState.CommandSpawn(words)
      if err != nil {
        log.Printf("Could not spawn: %v", err)
      }
      continue
    }

    if words[0] == "move" {
      move, err := gameState.CommandMove(words)
      if err != nil {
        log.Printf("Could not move: %v", err)
      } else {
        log.Printf("move worked: %v", move)
      }
    }

  }

	fmt.Println("Waiting for messages. To exit press CTRL+C")
	forever := make(chan bool)
	<-forever
}
