package main

import (
	"fmt"

	"github.com/JonathanVil/protocol-visualizer"
)

type HelloWorldActor struct {
	id  int
	sim *simulator.Simulator
}

func (actor *HelloWorldActor) ID() int {
	return actor.id
}

func (actor *HelloWorldActor) OnMessage(msg simulator.Message) {
	fmt.Printf("Node %d received from Node %d: %v\n", actor.id, msg.From, msg.Payload)

	if msg.Payload != "pong" {
		actor.sim.Send(actor.id, msg.From, "pong")
	}
}

func main() {
	sim := simulator.New()
	srv := simulator.NewServer(sim, ":8067")
	defer srv.Close()

	a := &HelloWorldActor{id: 0, sim: sim}
	b := &HelloWorldActor{id: 1, sim: sim}

	sim.Register(a)
	sim.Register(b)

	sim.Send(0, 1, "ping")

	srv.StartServer()
}
