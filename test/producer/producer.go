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
			// Generate Pokémon
			pokémon := initialPokémons[rand.Intn(len(initialPokémons))]

			// Build Payload Object
			pO, _ := msg.BuildPayloadObject(shared.Pokémon{Name: pokémon}, nil)

			// Build Message
			msg := msg.Message{
				Payload: map[string]msg.PayloadObject{
					consumer.EventEvolvePokémon: pO,
				},
			}

			// Publish message
			err := comm.Publish(msg, consumer.Service, consumer.EventEvolvePokémon)
			if err != nil {
				logger.WithError(err).Error("error sending pokémon")
			}

			logger.WithField("pokémon", pokémon).Info("sent")

			wg.Done()
		}()
	}
	wg.Wait()

	// Set handler
	comm.On(shared.EventReceivePokémon,
		func(m msg.Message) {
			// Receive message from "consumer"
			pokémon := new(shared.Pokémon)
			pO, _ := m.CurrentPayload()
			pO.Bind(pokémon)

			logger.WithField("pokémon", pokémon.Name).Info("received")
		},
	)

	// Start consuming messages
	comm.Consume()
}
