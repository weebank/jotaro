package consumer

import (
	"os"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/weebank/jotaro/msg"
	"github.com/weebank/jotaro/test/consumer/shared"
	producerShared "github.com/weebank/jotaro/test/producer/shared"
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
		func(m msg.Message) (any, error) {
			// Receive messages from "producer"
			pokémon := &producerShared.Pokémon{}
			m.Bind(pokémon)
			logger.WithField("pokémon", pokémon.Name).Info("received")

			// Evolve pokémon
			pokémon.Name = pokémonEvolutions[pokémon.Name]

			// Return to producer
			logger.WithField("pokémon", pokémon.Name).Info("sent")
			return pokémon, nil
		},
	)

	// Consume messages
	logger.Info("awaiting for pokémons")
	comm.Consume()
}
