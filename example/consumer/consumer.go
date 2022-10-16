package consumer

import (
	"os"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/weebank/jotaro/example/consumer/shared"
	producer "github.com/weebank/jotaro/example/producer/shared"
	"github.com/weebank/jotaro/msg"
)

var logger = apexLogger.Logger{
	Handler: cli.New(os.Stdout),
}

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
		func(m msg.Message) any {
			// Receive messages from "producer"
			pokémon := new(producer.Pokémon)
			pO, _ := m.CurrentPayload()
			pO.Bind(pokémon)

			logger.WithField("pokémon", pokémon.Name).Info("received")

			// Evolve pokémon
			pokémon.Name = pokémonEvolutions[pokémon.Name]

			logger.WithField("pokémon", pokémon.Name).Info("sent")

			return pokémon
		},
	)

	// Consume messages
	logger.Info("awaiting for pokémons")
	comm.Consume()
}
