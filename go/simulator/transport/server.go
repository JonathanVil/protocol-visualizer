package simulator

import (
	"context"
	"log"
	"net/http"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type Server struct {
	sim    *Simulator
	addr   string
	server *http.Server
}

type WebSocketMessage struct {
	Type    string         `json:"type"`
	Payload map[string]any `json:"payload"`
}

func NewServer(sim *Simulator, addr string) *Server {
	return &Server{sim: sim, addr: addr}
}

func (s *Server) StartServer() {
	http.HandleFunc("/ws", s.handleWS)
	err := http.ListenAndServe(s.addr, nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true,
	})
	if err != nil {
		return
	}
	defer conn.Close(websocket.StatusNormalClosure, "")

	ctx, cancel := context.WithCancel(r.Context())
	defer cancel()

	go func() {
		for {
			var request WebSocketMessage
			if err := wsjson.Read(ctx, conn, &request); err != nil {
				return
			}
			switch request.Type {
			case "start":
				go s.sim.Start()
			case "stop":
				s.sim.Stop()
			case "spawn":
				name, ok := request.Payload["name"].(string)
				if !ok || name == "" {
					log.Printf("invalid spawn request: missing or null actor name")
					continue
				}
				s.sim.SpawnActor(name, -1)
			case "requestTypes":
				var reply = WebSocketMessage{"actorTypes", map[string]any{"actors": s.sim.GetActorTypes()}}
				if err := wsjson.Write(ctx, conn, reply); err != nil {
					log.Printf("failed to send actorTypes: %v", err)
				}
			}
		}
	}()

	for event := range s.sim.events {
		if err := wsjson.Write(ctx, conn, WebSocketMessage{"event", map[string]any{"event": event}}); err != nil {
			return
		}
	}
}

func (s *Server) Close() error {
	return s.server.Close()
}
