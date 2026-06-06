package main

import (
	"github.com/JonathanVil/protocol-visualizer"
)

func main() {
	sim := simulator.New()
	srv := simulator.NewServer(sim, ":8067")
	defer srv.Close()

	a := &PingActor{id: 0, sim: sim}
	b := &PingActor{id: 1, sim: sim}

	sim.Register(a)
	sim.Register(b)

	sim.Send(0, 1, "ping")

	srv.StartServer()
}
