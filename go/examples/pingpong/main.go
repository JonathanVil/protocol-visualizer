package main

import (
	"reflect"

	"github.com/JonathanVil/protocol-visualizer"
)

func main() {
	sim := simulator.New()
	sim.TransitTicks = 5
	srv := simulator.NewServer(sim, ":8067")
	defer srv.Close()

	sim.RegisterActorType(reflect.TypeOf(&PingActor{}), "ping")

	srv.StartServer()
}
