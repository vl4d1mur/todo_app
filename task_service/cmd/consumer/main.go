package main

import (
	"fmt"

	"github.com/nats-io/nats.go"
)

func main() {
	nc, _ := nats.Connect("nats://localhost:4222")
	defer nc.Drain()

	nc.Subscribe("task-events", func(msg *nats.Msg) {
		fmt.Printf("Event received: %s\n", string(msg.Data))
	})

	fmt.Println("Listening for task-events...")
	select {}
}
