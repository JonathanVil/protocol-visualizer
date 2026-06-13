package main

import (
	"reflect"

	simulator "github.com/JonathanVil/protocol-visualizer"
	"github.com/JonathanVil/protocol-visualizer/transport"
)

func main() {
	sim := simulator.New()
	sim.TransitTicks = 5
	srv := transport.NewServer(sim, ":8067")
	defer srv.Close()

	sim.RegisterActorType(reflect.TypeOf(&PingActor{}), "ping")

	srv.StartServer()
}
