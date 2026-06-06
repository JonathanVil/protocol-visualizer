package simulator

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Server struct {
	sim  *Simulator
	addr string
}

func NewServer(sim *Simulator, addr string) *Server {
	return &Server{sim: sim, addr: addr}
}

func (s *Server) Listen() error {
	http.HandleFunc("/ws", s.handleWS)
	return http.ListenAndServe(s.addr, nil)
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		for {
			_, data, err := conn.Read(ctx)
			if err != nil {
				return
			}
			switch string(data) {
			case "start":
				go s.sim.Start()
			case "stop":
				s.sim.Stop()
			}
		}
	}()

	for event := range s.sim.events {
		if err := wsjson.Write(ctx, conn, event); err != nil {
			return
		}
	}
}
