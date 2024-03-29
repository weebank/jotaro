package producer

import (
	"math/rand"
	"os"
	"sync"
	"time"

	apexLogger "github.com/apex/log"
	"github.com/apex/log/handlers/cli"
	consumer "github.com/weebank/jotaro/example/consumer/shared"
	"github.com/weebank/jotaro/example/producer/shared"
	"github.com/weebank/jotaro/msg"
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

	// Send messages to "consumer"
	for i := 0; i < int(count); i++ {
		wg.Add(1)

		go func() {
			// Generate Pokémon
			pokémon := initialPokémons[rand.Intn(len(initialPokémons))]

			// Build Payload Object
			pO, _ := msg.NewPayloadObject(shared.Pokémon{Name: pokémon})

			// Build Message
			msg := msg.Message{
				Payload: map[string]msg.PayloadObject{
					consumer.EventEvolvePokémon: pO,
				},
			}

			// Publish message
			if err := comm.Publish(msg, consumer.Service, consumer.EventEvolvePokémon); err != nil {
				logger.WithError(err).Error("error sending pokémon")
			}

			logger.WithField("pokémon", pokémon).Info("sent")

			wg.Done()
		}()
	}
	wg.Wait()

	// Set handler
	comm.On(msg.ResponseEvent(consumer.EventEvolvePokémon),
		func(m msg.Message) any {
			// Receive message from "consumer"
			pokémon := new(shared.Pokémon)
			pO, _ := m.CurrentPayload()
			pO.Bind(pokémon)

			logger.WithField("pokémon", pokémon.Name).Info("received")

			return nil
		},
	)

	// Start consuming messages
	comm.Consume()
}
