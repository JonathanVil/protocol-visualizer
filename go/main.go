package main

import (
	"fmt"
	"time"

	"github.com/JonathanVil/protocol-visualizer/simulator"
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

	a := &HelloWorldActor{id: 0, sim: sim}
	b := &HelloWorldActor{id: 1, sim: sim}

	sim.Register(a)
	sim.Register(b)

	sim.Send(0, 1, "ping")

	go sim.Start()
	time.Sleep(10 * time.Second)
	sim.Stop()
}
