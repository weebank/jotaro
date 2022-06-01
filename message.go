package rmq

import (
	"hash/fnv"
	"os"
	"strconv"

	"github.com/streadway/amqp"
	"github.com/weebank/rmq/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type MessagingService struct {
	name      string
	conn      *amqp.Connection
	ch        *amqp.Channel
	queues    []amqp.Queue
	handlers  map[string]func(id string, bytes []byte, index int)
	callbacks map[string]func(bytes []byte, index int)
}

func unwrap(bytes []byte) (id string, any *anypb.Any, err error) {
	wrapping := &pb.Wrapping{}
	if err := proto.Unmarshal(bytes, wrapping); err != nil {
		return "", nil, err
	}

	any = &anypb.Any{}
	if err := proto.Unmarshal(bytes, any); err != nil {
		return wrapping.GetId(), nil, err
	}

	return wrapping.GetId(), any, nil
}

func wrap(id string, msg protoreflect.ProtoMessage) []byte {
	any, _ := anypb.New(msg)
	body, _ := proto.Marshal(any)
	wrapping := &pb.Wrapping{Id: id, Body: body}
	bytes, _ := proto.Marshal(wrapping)

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

	return int(h.Sum32()) % queueCount
}
