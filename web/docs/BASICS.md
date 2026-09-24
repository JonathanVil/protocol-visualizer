# System Basics
Welcome!
This is a tool for visualizing, simulating, debugging and experimenting with distributed protocols.
You implement the actors of a protocol in Go, run your program, and this page connects to it
to show how the actors interact in a simulated network.

## Actors
Actors are an abstraction of a machine running a protocol.
Actors interact with other actors by sending and receiving messages.
Spawn actors of the types your program registered using the panel in the bottom left.

You can see the state and methods of an actor by pressing the icons at the top of the box next to it.
Through this box you can also modify the state by pressing the small pencil next to a field,
or run a method by pressing the small arrow.

## Killing and reviving actors
Press **Kill** in an actor's box to simulate a crash. A dead actor is shown in grey. Messages that
arrive at it are dropped, and it can't send messages or run methods. Messages it sent before
dying still arrive. Its state is kept, and you can still edit it.

Press **Revive** to bring it back with the state it had. If the actor has an `OnRevive()` method,
it is called as the actor comes back, which you can use to reset state a real node would lose in a crash.

## Ticks
The simulator is built on a "tick" system, like you might see in a video game.
Ticks serve as the smallest unit of time in the simulation.

## Messages
Every tick each message currently in transit moves one tick closer to its recipient.
The number of ticks in transit is the transit time, which may be changed in the settings in the top right.
Once a message reaches the end, it is delivered and the recipient handles it according to the protocol.

Click a message to deliver it immediately, drop it, or delay it.

## Manually sending messages
You can manually send a message from one actor to another using the panel in the bottom left.
The payload is sent as JSON if it is valid JSON (e.g. `5`, `true` or `{"term": 2}`), and as a string otherwise.

## Controlling the simulation
Start and pause the simulation using the controls in the bottom right.
Some basic settings are available in the top right.

## The log
The bottom tab on the top left is the log. Here you can see a list of all events that have occurred on a given tick.

# Implementation Basics
An actor is a Go type that implements the `simulator.Actor` interface:

```go
type Actor interface {
    ID() int
    OnMessage(Message)
    Init(id int, sim *Simulator)
}
```

A bare minimum implementation looks like this:

```go
type PingActor struct {
    id  int
    sim *simulator.Simulator

    PingsReceived int // exported fields are shown in the visualizer
}

func (a *PingActor) ID() int { return a.id }

func (a *PingActor) Init(id int, sim *simulator.Simulator) {
    a.id = id
    a.sim = sim
}

func (a *PingActor) OnMessage(msg simulator.Message) {
    if msg.Payload == "ping" {
        a.PingsReceived++
        a.sim.Send(a.id, msg.From, "pong")
    }
}

// Exported methods can be run from the visualizer.
func (a *PingActor) SendPing(to int) error {
    return a.sim.Send(a.id, to, "ping")
}
```

Then register the actor type and start the server in your `main` function:

```go
func main() {
    sim := simulator.New()
    srv := transport.NewServer(sim, ":8067")
    defer srv.Close()

    sim.RegisterActorType(reflect.TypeOf(&PingActor{}), "ping")

    srv.StartServer()
}
```

Run it with `go run .` and reload this page. See `go/examples/pingpong` in the repository for a complete example.

## Fields
Exported fields are shown next to each actor and can be edited from the visualizer.
A `json` struct tag changes the name a field is shown under, and `json:"-"` hides it.
Editing supports strings, numbers and booleans.

## Methods
Exported methods (other than `ID`, `Init`, `OnMessage` and `OnRevive`) can be run from the visualizer.
Arguments of type string, number or bool are supported. If the last return value is an `error`, it is shown when it is not `nil`.

## Reviving
An actor can optionally implement `OnRevive()`. It's called when the actor is revived after being killed:

```go
func (a *PingActor) OnRevive() {
    a.PingsReceived = 0 // state kept only in memory is lost in a crash
}
```

## Sending messages
`sim.Send(from, to, payload)` puts a message in transit from actor `from` to actor `to`.
The payload can be any value that can be encoded as JSON. `OnMessage` receives it as `msg.Payload`,
along with the sender in `msg.From`.
