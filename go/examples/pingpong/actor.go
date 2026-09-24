package main

import (
	"fmt"

	simulator "github.com/JonathanVil/protocol-visualizer"
)

// PingActor answers every "ping" with a "pong".
//
// Exported fields are shown (and editable) in the visualizer, and exported
// methods can be invoked from it.
type PingActor struct {
	id  int
	sim *simulator.Simulator

	PingsReceived int
	PongsReceived int
}

func (actor *PingActor) ID() int {
	return actor.id
}

func (actor *PingActor) OnMessage(msg simulator.Message) {
	fmt.Printf("Node %d received from Node %d: %v\n", actor.id, msg.From, msg.Payload)

	if msg.Payload == "pong" {
		actor.PongsReceived++
		return
	}
	actor.PingsReceived++
	actor.sim.Send(actor.id, msg.From, "pong")
}

func (actor *PingActor) Init(id int, sim *simulator.Simulator) {
	actor.id = id
	actor.sim = sim
}

// SendPing sends a "ping" to another actor.
func (actor *PingActor) SendPing(to int) error {
	return actor.sim.Send(actor.id, to, "ping")
}
