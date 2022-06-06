package rmq

import (
	"hash/fnv"

	"github.com/streadway/amqp"
	"github.com/weebank/rmq/pb"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/anypb"
)

type Response struct {
	Route   string
	Message protoreflect.ProtoMessage
}

type MessagingService struct {
	name      string
	conn      *amqp.Connection
	ch        *amqp.Channel
	handlers  map[string]func(bytes []byte) *Response
	callbacks map[string]func(bytes []byte)
}

func unwrap(bytes []byte) (id string, isCallback bool, blockCallback bool, any *anypb.Any, err error) {
	wrapping := &pb.Wrapping{}
	if err := proto.Unmarshal(bytes, wrapping); err != nil {
		return "", false, false, nil, err
	}

	any = &anypb.Any{}
	if err := proto.Unmarshal(wrapping.Body, any); err != nil {
		return wrapping.GetId(), wrapping.GetIsCallback(), wrapping.GetBlockCallback(), nil, err
	}

	return wrapping.GetId(), wrapping.GetIsCallback(), wrapping.GetBlockCallback(), any, nil
}

func wrap(id string, isCallback bool, blockCallback bool, msg protoreflect.ProtoMessage) []byte {
	any, _ := anypb.New(msg)
	body, _ := proto.Marshal(any)
	wrapping := &pb.Wrapping{Id: id, IsCallback: isCallback, BlockCallback: blockCallback, Body: body}
	bytes, _ := proto.Marshal(wrapping)

	return bytes
}

func queueIndex(id string, queues uint) uint {
	h := fnv.New32a()
	h.Write([]byte(id))

	return uint(h.Sum32()) % queues
}
