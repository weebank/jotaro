package producer

import (
	"math/rand"
	"os"
	"time"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/weebank/jotaro/msg"
	consumerShared "github.com/weebank/jotaro/test/consumer/shared"
	"github.com/weebank/jotaro/test/producer/shared"
)

var logger = apexLogger.Logger{
	Handler: cli.New(os.Stdout),
}

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
	for i := 0; i < 10000; i++ {
		// Build message
		pokémon := shared.Pokémon{
			Name: initialPokémons[rand.Intn(len(initialPokémons))],
		}
		err := comm.Publish(consumerShared.Service, consumerShared.EventEvolvePokémon, pokémon)
		if err != nil {
			logger.WithError(err).Error("error sending pokémon")
		}
		logger.WithField("pokémon", pokémon.Name).Info("sent")
	}

	// Set handler
	comm.On(consumerShared.EventEvolvePokémon,
		func(m msg.Message) (any, error) {
			// Receive message from "consumer"
			pokémon := &shared.Pokémon{}
			m.Bind(pokémon)
			logger.WithField("pokémon", pokémon.Name).Info("received")

			return nil, nil
		},
	)

	// Start consuming messages
	comm.Consume()
}
