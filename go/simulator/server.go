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
			var command WebSocketMessage
			if err := wsjson.Read(ctx, conn, &command); err != nil {
				return
			}
			switch command.Type {
			case "start":
				go s.sim.Start()
			case "stop":
				s.sim.Stop()
			case "spawn":
				name := command.Payload["name"].(string)
				s.sim.SpawnActor(name, -1)
			}
		}
	}()

	for event := range s.sim.events {
		if err := wsjson.Write(ctx, conn, event); err != nil {
			return
		}
	}
}

func (s *Server) Close() error {
	return s.server.Close()
}

func (s *Server) SendActorTypes(ctx context.Context, conn *websocket.Conn) {
	var msg WebSocketMessage = WebSocketMessage{"type", make(map[string]any)}
	msg.Payload["actors"] = s.sim.actorTypes
	if err := wsjson.Write(ctx, conn, msg); err != nil {
		log.Printf("failed to send actorTypes: %v", err)
	}
}
