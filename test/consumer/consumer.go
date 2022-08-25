package consumer

import (
	"fmt"

	"github.com/weebank/jotaro/msg"
	"github.com/weebank/jotaro/test/consumer/shared"
	producerShared "github.com/weebank/jotaro/test/producer/shared"
)

var pokémonEvolutions = map[string]string{
	"bulbasaur":  "ivysaur",
	"charmander": "charmeleon",
	"squirtle":   "wartortle",
}

func Main() {
	// Initialize service
	comm := msg.NewService("consumer")
	defer comm.Close()

	// Set handler
	comm.On(shared.EventEvolvePokémon,
		func(m msg.Message) (any, error) {
			// Receive messages from "producer"
			pokémon := &producerShared.Pokémon{}
			m.Bind(pokémon)
			fmt.Println("pokémon received:", pokémon.Name)

			// Evolve pokémon
			pokémon.Name = pokémonEvolutions[pokémon.Name]

			// Return to producer
			return pokémon, nil
		},
	)

	// Consume messages
	fmt.Println("awaiting for pokémons...")
	comm.Consume()
}
