# rmq

This service allows microservices through the backend to implement RabbitMQ message exchanges.

## Example

Messages are defined in Protobuf files. In this example, let's use four `.proto` files. They are:

```protobuf
// ./pb/banana.proto
syntax="proto3";

package main;
option go_package = "./pb";

message Banana {
    float yellowness = 1;
}
```

```protobuf
// ./pb/apple.proto
syntax="proto3";

package main;
option go_package = "./pb";

message Apple {
    float redness = 1;
}
```

```protobuf
// ./pb/spider.proto
syntax="proto3";

package main;
option go_package = "./pb";

message Spider {
    string poison = 1;
}
```

```protobuf
// ./pb/cow.proto
syntax="proto3";

package main;
option go_package = "./pb";

message Cow {
    bool milk = 1;
}
```

> Remember to compile Protobuf files with `protoc --go_out=. pb/*.proto`

Now, on the main Go file, we can write a service like this:

```go
// ./main.go

func main() {
    // Create service and define name and exchanges to listen
    service := rmq.NewService("insectsAndFruits", []string{"insects", "fruits"})
    defer service.Close()

    // Bind handler for "Banana"
    service.Bind(&pb.Banana{},
        func(bytes []byte) *rmq.Response {
            banana := &pb.Banana{}
            proto.Unmarshal(bytes, banana)

            fmt.Println("banana received", "yellowness:", banana.GetYellowness())

            // When receive a "Banana", return an "Apple" as a callback
            return &rmq.Response{Message: &pb.Apple{Redness: 1.0}}
        },
    )

    // Bind handler for "Apple"
    service.Bind(&pb.Apple{},
        func(bytes []byte) *rmq.Response {
            apple := &pb.Apple{}
            proto.Unmarshal(bytes, apple)

            fmt.Println("apple received", "redness:", apple.GetRedness())

            // When receive an "Apple", publish a "Cow" to "mammals" exchange
            service.Publish("mammals", &pb.Cow{Milk: true})

            return nil
        },
    )

    // Bind handler for "Spider"
    service.Bind(&pb.Spider{},
        func(bytes []byte) *rmq.Response {
            spider := &pb.Spider{}
            proto.Unmarshal(bytes, spider)

            fmt.Println("spider received", "poison:", spider.GetPoison())

            // When receive a "Spider", also publish a "Cow" to "mammals" exchange,
            // specifying a callback function to be called by another service
            service.PublishEvent("mammals", &pb.Cow{Milk: true},
                func(bytes []byte) { fmt.Println("received cow callback") })

            return nil
        },
    )

    // Consume messages
    service.Consume()
}
```

## Known issues/limitations

- All microservices must have the same amount of queues (`QUEUE_COUNT`)
