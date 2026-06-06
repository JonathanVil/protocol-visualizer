package simulator

type Actor interface {
	ID() int
	OnMessage(Message)
}
