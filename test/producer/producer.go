package producer

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/weebank/jotaro/msg"
	consumerShared "github.com/weebank/jotaro/test/consumer/shared"
	"github.com/weebank/jotaro/test/producer/shared"
)

var initialPokémons = []string{
	"bulbasaur",
	"charmander",
	"squirtle",
}

func Main() {
	// Set random seed
	rand.Seed(time.Now().Unix())

	// Initialize service
	comm := msg.NewService("producer")
	defer comm.Close()

	// Send messages to "pokémons"
	for i := 0; i < 500000; i++ {
		// Build message
		pokémon := shared.Pokémon{
			Name: initialPokémons[rand.Intn(len(initialPokémons))],
		}
		err := comm.Publish(consumerShared.Service, consumerShared.EventEvolvePokémon, pokémon)
		if err != nil {
			fmt.Println("error sending pokémon:", err)
		}
		fmt.Println("pokémon sent:", pokémon.Name)
	}

	// Set handler
	comm.On(consumerShared.EventEvolvePokémon,
		func(m msg.Message) (any, error) {
			// Receive message from "consumer"
			pokémon := &shared.Pokémon{}
			m.Bind(pokémon)
			fmt.Println("pokémon received:", pokémon.Name)

			return nil, nil
		},
	)

	// Start consuming messages
	comm.Consume()
}
