package rmq

import (
	"github.com/streadway/amqp"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Message struct {
	ID   string // ID for routing
	Body []byte // Message content
}

type MessagingService struct {
	conn     *amqp.Connection
	ch       *amqp.Channel
	queues   []amqp.Queue
	handlers map[string]func(msg protoreflect.ProtoMessage)
}

func bytesToAny(bytes []byte) (any anypb.Any) {
	proto.Unmarshal(bytes, &any)
	return
}

func wrapProto(msg protoreflect.ProtoMessage) []byte {
	any, _ := anypb.New(msg)
	p, _ := proto.Marshal(any)
	return p
}
