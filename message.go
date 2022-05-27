package rmq

import (
	"github.com/golang/protobuf/ptypes/any"
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type RoutedMessage struct {
	ID   string // ID for routing
	Body []byte // Message content
}

type MessagingService struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	queues   []amqp.Queue
	handlers map[string]func(bytes []byte)
}

func unwrap(bytes []byte) *anypb.Any {
	any := &any.Any{}
	proto.Unmarshal(bytes, any)
	return any
}

func wrap(msg protoreflect.ProtoMessage) []byte {
	any, _ := anypb.New(msg)
	p, _ := proto.Marshal(any)
	return p
}
