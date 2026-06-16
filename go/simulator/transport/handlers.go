package transport

import (
	"fmt"

	simulator "github.com/JonathanVil/protocol-visualizer"
	"github.com/JonathanVil/protocol-visualizer/transport/frames"
)

// registerHandlers wires all commands to their sim methods using the per-Server registry.
func registerHandlers(sim *simulator.Simulator, reg map[string]Handler) {
	// --- Simulation control ---

	Register(reg, "start", func(_ struct{}) (any, error) {
		go sim.Start()
		return nil, nil
	})

	Register(reg, "stop", func(_ struct{}) (any, error) {
		sim.Stop()
		return nil, nil
	})

	Register(reg, "spawn", func(p frames.SpawnPayload) (any, error) {
		actor := sim.SpawnActor(p.Name, -1)
		if actor == nil {
			return nil, fmt.Errorf("%w: unknown actor type %q", simulator.ErrInvalidArgument, p.Name)
		}
		return actor.ID(), nil
	})

	Register(reg, "requestTypes", func(_ struct{}) (any, error) {
		return sim.GetActorTypes(), nil
	})

	// --- Message manipulation ---
	Register(reg, "message.send", func(p frames.MessageSendPayload) (any, error) {
		return nil, sim.Send(p.From, p.To, p.Payload)
	})

	Register(reg, "message.drop", func(p frames.MessageDropPayload) (any, error) {
		return nil, sim.DropMessage(p.MessageID)
	})

	Register(reg, "message.delay", func(p frames.MessageDelayPayload) (any, error) {
		return nil, sim.DelayMessage(p.MessageID, p.Ticks)
	})

	Register(reg, "message.deliverNow", func(p frames.MessageDeliverNowPayload) (any, error) {
		return nil, sim.DeliverNow(p.MessageID)
	})

	// --- Actor manipulation ---

	Register(reg, "actor.setField", func(p frames.ActorSetFieldPayload) (any, error) {
		return nil, sim.SetActorField(p.ActorID, p.Field, p.Value)
	})

	Register(reg, "actor.invoke", func(p frames.ActorInvokePayload) (any, error) {
		return sim.InvokeActor(p.ActorID, p.Method, p.Args)
	})

	// --- Sim settings ---

	Register(reg, "sim.setSpeed", func(p frames.SimSetSpeedPayload) (any, error) {
		return nil, sim.SetSpeed(p.Multiplier)
	})

	Register(reg, "sim.setTransitTime", func(p frames.SimSetTransitTimePayload) (any, error) {
		return nil, sim.SetTransitTime(p.Ticks)
	})
}
