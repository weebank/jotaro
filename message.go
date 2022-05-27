package rmq

import (
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Message protoreflect.ProtoMessage

type RoutedMessage struct {
	ID   string // ID for routing
	Body []byte // Message content
}

type handler struct {
	function func(msg Message)
	data     Message
}

type MessagingService struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	queues   []amqp.Queue
	handlers map[string]handler
}

func unwrap(bytes []byte) (any *anypb.Any) {
	proto.Unmarshal(bytes, any)
	return
}

func wrap(msg protoreflect.ProtoMessage) []byte {
	any, _ := anypb.New(msg)
	p, _ := proto.Marshal(any)
	return p
}
