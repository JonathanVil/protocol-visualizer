package simulator

import (
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

type Simulator struct {
	mu           sync.Mutex
	actors       map[int]Actor
	queue        []Message
	tick         int
	tickDuration time.Duration
	running      bool
}

func New() *Simulator {
	return &Simulator{
		actors:       make(map[int]Actor),
		tick:         0,
		tickDuration: time.Second,
	}
}

func (s *Simulator) Register(a Actor) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.actors[a.ID()] = a
	fmt.Printf("Registered actor %d\n", a.ID())
}

func (s *Simulator) Send(from, to int, payload any) {
	s.mu.Lock()
	defer s.mu.Unlock()

	i := rand.IntN(len(s.queue) + 1)
	msg := Message{From: from, To: to, Payload: payload}
	s.queue = append(s.queue[:i], append([]Message{msg}, s.queue[i:]...)...)
}

func (s *Simulator) Tick() {
	s.mu.Lock()
	if len(s.queue) == 0 {
		return
	}

	fmt.Printf("-- Tick %d --\n", s.tick)

	// get message to deliver
	msg := s.queue[0]
	s.queue = s.queue[1:]

	// deliver it
	actor := s.actors[msg.To]
	if actor == nil {
		fmt.Printf("No actor found %d\n", msg.To)
		return
	}
	s.tick++
	s.mu.Unlock()

	actor.OnMessage(msg)
}

func (s *Simulator) Start() {
	s.running = true

	for s.running {
		s.Tick()
		time.Sleep(s.tickDuration)
	}
}

func (s *Simulator) Stop() {
	s.running = false
}
