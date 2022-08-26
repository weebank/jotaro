package producer

import (
	"math/rand"
	"os"
	"sync"
	"time"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	"github.com/weebank/jotaro/msg"
	consumer "github.com/weebank/jotaro/test/consumer/shared"
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

func Main(count uint) {
	// Set random seed
	rand.Seed(time.Now().Unix())

	// Initialize service
	comm := msg.NewService("producer")
	defer comm.Close()

	// Set a Wait Group so the program starts to receive
	// pokémons only after they were sent
	wg := sync.WaitGroup{}

	// Send messages to "pokémons"
	for i := 0; i < int(count); i++ {
		wg.Add(1)

		go func() {
			// Build message
			pokémon := shared.Pokémon{
				Name: initialPokémons[rand.Intn(len(initialPokémons))],
			}
			err := comm.Publish(msg.Message{}, consumer.Service, consumer.EventEvolvePokémon, pokémon, nil)
			if err != nil {
				logger.WithError(err).Error("error sending pokémon")
			}
			logger.WithField("pokémon", pokémon.Name).Info("sent")

			wg.Done()
		}()
	}
	wg.Wait()

	// Set handler
	comm.On(shared.EventReceivePokémon,
		func(m msg.Message) {
			// Receive message from "consumer"
			pokémon := new(shared.Pokémon)
			m.Bind(pokémon)

			logger.WithField("pokémon", pokémon.Name).Info("received")
		},
	)

	// Start consuming messages
	comm.Consume()
}
