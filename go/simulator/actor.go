package simulator

type Actor interface {
	ID() int
	OnMessage(Message)
	Init(id int, sim *Simulator)
}

// Reviver is optionally implemented by actors that need to react to being
// revived after a kill, e.g. to reset state that would not survive a crash.
type Reviver interface {
	OnRevive()
}
