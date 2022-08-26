package consumer

import (
	"os"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/weebank/jotaro/msg"
	"github.com/weebank/jotaro/test/consumer/shared"
	producer "github.com/weebank/jotaro/test/producer/shared"
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
		func(m *msg.Message) {
			// Receive messages from "producer"
			pokémon := new(producer.Pokémon)
			m.BindLatest(pokémon)

			logger.WithField("pokémon", pokémon.Name).Info("received")

			// Evolve pokémon
			pokémon.Name = pokémonEvolutions[pokémon.Name]

			logger.WithField("pokémon", pokémon.Name).Info("sent")

			// Prepare message to send it back to producer
			m.Prepare(producer.EventReceivePokémon, producer.Service)
		},
	)

	// Consume messages
	logger.Info("awaiting for pokémons")
	comm.Consume()
}
