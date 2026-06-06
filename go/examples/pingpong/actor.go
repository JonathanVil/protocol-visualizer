package main

import (
	"fmt"

	simulator "github.com/JonathanVil/protocol-visualizer"
)

type PingActor struct {
	id  int
	sim *simulator.Simulator
}

func (actor *PingActor) ID() int {
	return actor.id
}

func (actor *PingActor) OnMessage(msg simulator.Message) {
	fmt.Printf("Node %d received from Node %d: %v\n", actor.id, msg.From, msg.Payload)

	if msg.Payload != "pong" {
		actor.sim.Send(actor.id, msg.From, "pong")
	}
}
