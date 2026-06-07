package simulator

type Actor interface {
	ID() int
	OnMessage(Message)
	Init(id int, sim *Simulator)
}
