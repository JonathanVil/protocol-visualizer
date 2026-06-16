// Package transport implements the bidirectional WebSocket protocol for the simulator.
package transport

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	simulator "github.com/JonathanVil/protocol-visualizer"
	"github.com/JonathanVil/protocol-visualizer/transport/frames"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// Server wraps a Simulator and exposes it over WebSocket.
type Server struct {
	sim      *simulator.Simulator
	addr     string
	server   *http.Server
	registry map[string]Handler
}

func NewServer(sim *simulator.Simulator, addr string) *Server {
	s := &Server{
		sim:      sim,
		addr:     addr,
		registry: make(map[string]Handler),
	}
	registerHandlers(sim, s.registry)
	return s
}

func (s *Server) StartServer() {
	mux := http.NewServeMux()
	mux.HandleFunc("/ws", s.handleWS)
	s.server = &http.Server{Addr: s.addr, Handler: mux}
	if err := s.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *Server) Close() error {
	if s.server != nil {
		return s.server.Close()
	}
	return nil
}

// handleWS is the per-connection WebSocket handler.
func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{InsecureSkipVerify: true})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	// writes funnels all outbound frames through a single goroutine.
	// coder/websocket forbids concurrent writes.
	writes := make(chan any, 64)

	var seq atomic.Int64

	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case frame := <-writes:
				if err := wsjson.Write(ctx, conn, frame); err != nil {
					return
				}
			}
		}
	}()

	sendFrame := func(frame any) {
		select {
		case writes <- frame:
		case <-ctx.Done():
		}
	}

	// Request a consistent snapshot from the sim goroutine and send it first.
	// Enqueue executes synchronously when the sim is not yet running.
	snapResult := <-s.sim.Enqueue(func() (any, error) {
		return s.sim.GetSnapshot(), nil
	})
	snap := snapResult.Value.(simulator.Snapshot)
	sendFrame(frames.SnapshotFrame{
		Type:    "snapshot",
		Seq:     seq.Load(),
		Tick:    snap.Tick,
		Payload: snap,
	})

	// Forward sim events to the write goroutine with per-connection seq numbers.
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case event := <-s.sim.Events():
				sendFrame(frames.EventFrame{
					Seq:     seq.Add(1),
					Tick:    event.Tick,
					Type:    string(event.Type),
					Payload: event.Payload,
				})
			}
		}
	}()

	// Read loop: decode envelopes and dispatch through the registry.
	for {
		var env frames.Envelope
		if err := wsjson.Read(ctx, conn, &env); err != nil {
			return
		}

		handler, ok := s.registry[env.Type]
		if !ok {
			sendFrame(frames.ErrorFrame{
				Type:    "error",
				ReplyTo: env.ID,
				Code:    "UNKNOWN_COMMAND",
				Message: fmt.Sprintf("unrecognized command type %q", env.Type),
			})
			continue
		}

		capturedEnv := env
		capturedHandler := handler

		// Wait for the sim-goroutine result in a separate goroutine so the read
		// loop is never blocked by in-flight commands.
		go func() {
			result := <-s.sim.Enqueue(func() (any, error) {
				return capturedHandler(capturedEnv.Payload)
			})
			if result.Err != nil {
				sendFrame(frames.ErrorFrame{
					Type:    "error",
					ReplyTo: capturedEnv.ID,
					Code:    simulator.GetCode(result.Err),
					Message: result.Err.Error(),
				})
			} else {
				sendFrame(frames.AckFrame{
					Type:    "ack",
					ReplyTo: capturedEnv.ID,
					Result:  result.Value,
				})
			}
		}()
	}
}
