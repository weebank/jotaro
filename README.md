# rmq

This service allows microservices through the backend to implement RabbitMQ message exchanges.

## Example

```go
func main() {
    // Connect to RabbitMQ
    conn := rmq.NewConnection()

    // Declare exchange name
    exc := "capybara"
    conn.NewExchange(exc)

    // Declare queue
    q := conn.NewQueue()
    conn.Bind(q, exc)

    // Consume queue
    msgs, _ := conn.Consume(q)

    // Main queue loop
    forever := make(chan struct{})
    go func() {
        for d := range msgs {
            // Do something when messages arrive

            // Just as an example, this is how
            // you publish messages to another
            // exchange (in this case, "zebra")
            conn.Publish([]byte("hi"), "zebra")
        }
    }()
    <-forever
}
```

## Known issues/limitations

- Single threaded
- Single queue
- Still too low level (could forward less RabbitMQ abstractions)
- Exchanges are hardcoded
