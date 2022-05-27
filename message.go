package rmq

import (
	"hash/fnv"
	"os"
	"strconv"

	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type MessagingService struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	queues   []amqp.Queue
	handlers map[string]func(bytes []byte, index int)
}

func unwrap(bytes []byte) (*anypb.Any, error) {
	any := &anypb.Any{}
	if err := proto.Unmarshal(bytes, any); err != nil {
		return nil, err
	}

	return any, nil
}

func wrap(msg protoreflect.ProtoMessage) []byte {
	any, _ := anypb.New(msg)
	bytes, _ := proto.Marshal(any)

	return bytes
}

func queueIndex(id string) int {
	queueCount := 8
	if str, ok := os.LookupEnv("QUEUE_COUNT"); ok {
		if count, err := strconv.Atoi(str); err == nil {
			queueCount = count
		}
	}

	h := fnv.New32a()
	h.Write([]byte(id))
	h.Sum32()

	return int(h.Sum32()) % queueCount
}
