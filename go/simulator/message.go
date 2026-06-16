package simulator

import (
	"fmt"
	"sync/atomic"
)

var msgCounter atomic.Int64

type Message struct {
	ID      string
	From    int
	To      int
	Payload any
}

func newMessageID() string {
	return fmt.Sprintf("m-%d", msgCounter.Add(1))
}
