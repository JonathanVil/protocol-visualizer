package transport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
	"time"

	simulator "github.com/JonathanVil/protocol-visualizer"
	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

// testActor is a minimal actor for use in tests.
type testActor struct {
	id  int
	sim *simulator.Simulator
}

func (a *testActor) ID() int                             { return a.id }
func (a *testActor) Init(id int, s *simulator.Simulator) { a.id = id; a.sim = s }
func (a *testActor) OnMessage(_ simulator.Message)       {}

func newTestServer(t *testing.T) (*httptest.Server, *simulator.Simulator) {
	t.Helper()
	sim := simulator.New()
	srv := NewServer(sim, "") // addr unused — httptest provides the listener
	hs := httptest.NewServer(http.HandlerFunc(srv.handleWS))
	t.Cleanup(hs.Close)
	return hs, sim
}

func dialWS(t *testing.T, hs *httptest.Server) (*websocket.Conn, context.CancelFunc) {
	t.Helper()
	url := "ws" + strings.TrimPrefix(hs.URL, "http") + "/ws"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	conn, _, err := websocket.Dial(ctx, url, nil)
	if err != nil {
		cancel()
		t.Fatalf("dial: %v", err)
	}
	t.Cleanup(func() { conn.Close(websocket.StatusNormalClosure, "") })
	return conn, cancel
}

func readFrame(t *testing.T, ctx context.Context, conn *websocket.Conn) map[string]any {
	t.Helper()
	var frame map[string]any
	if err := wsjson.Read(ctx, conn, &frame); err != nil {
		t.Fatalf("readFrame: %v", err)
	}
	return frame
}

func TestIntegration_SnapshotOnConnect(t *testing.T) {
	hs, _ := newTestServer(t)
	conn, cancel := dialWS(t, hs)
	defer cancel()
	ctx := context.Background()

	frame := readFrame(t, ctx, conn)

	if frame["type"] != "snapshot" {
		t.Fatalf("first frame type = %v, want snapshot", frame["type"])
	}
	if _, ok := frame["seq"]; !ok {
		t.Error("snapshot missing seq")
	}
	if _, ok := frame["tick"]; !ok {
		t.Error("snapshot missing tick")
	}
	payload, ok := frame["payload"].(map[string]any)
	if !ok {
		t.Fatal("snapshot payload is not an object")
	}
	if _, ok := payload["Actors"]; !ok {
		t.Error("snapshot payload missing actors")
	}
	if _, ok := payload["Messages"]; !ok {
		t.Error("snapshot payload missing inTransit")
	}
}

func TestIntegration_MessageDropAckAndEvent(t *testing.T) {
	hs, sim := newTestServer(t)

	// Two actors, message that won't auto-deliver during the test.
	a0 := &testActor{}
	a0.Init(0, sim)
	a1 := &testActor{}
	a1.Init(1, sim)
	sim.Register(a0)
	sim.Register(a1)
	sim.TransitTicks = 99
	sim.Send(0, 1, "hello")

	go sim.Start()
	t.Cleanup(sim.Stop)

	conn, cancel := dialWS(t, hs)
	defer cancel()
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	// Consume the snapshot and extract the in-transit message ID.
	snapFrame := readFrame(t, ctx, conn)
	if snapFrame["type"] != "snapshot" {
		t.Fatalf("expected snapshot, got %v", snapFrame["type"])
	}
	snapPayloadRaw, _ := json.Marshal(snapFrame["payload"])
	var snap simulator.Snapshot
	if err := json.Unmarshal(snapPayloadRaw, &snap); err != nil {
		t.Fatalf("unmarshal snapshot: %v", err)
	}
	if len(snap.Messages) == 0 {
		t.Fatal("expected at least one in-transit message in snapshot")
	}
	msgID := snap.Messages[0].Id

	// Send the drop command.
	if err := wsjson.Write(ctx, conn, map[string]any{
		"id":      "c-1",
		"type":    "message.drop",
		"payload": map[string]any{"messageId": msgID},
	}); err != nil {
		t.Fatalf("write command: %v", err)
	}

	// Collect frames until we see both the ack and the message.dropped event.
	sawAck, sawEvent := false, false
	deadline := time.Now().Add(5 * time.Second)
	for !sawAck || !sawEvent {
		if time.Now().After(deadline) {
			t.Fatalf("timed out (ack=%v event=%v)", sawAck, sawEvent)
		}
		frame := readFrame(t, ctx, conn)
		switch frame["type"] {
		case "ack":
			if frame["replyTo"] != "c-1" {
				t.Errorf("ack replyTo = %v, want c-1", frame["replyTo"])
			}
			sawAck = true
		case "message.dropped":
			payload, ok := frame["payload"].(map[string]any)
			if !ok {
				t.Fatalf("message.dropped payload not an object")
			}
			if payload["messageId"] != msgID {
				t.Errorf("dropped messageId = %v, want %v", payload["messageId"], msgID)
			}
			if _, ok := frame["seq"]; !ok {
				t.Error("message.dropped missing seq")
			}
			if _, ok := frame["tick"]; !ok {
				t.Error("message.dropped missing tick")
			}
			sawEvent = true
		}
	}
}

func TestIntegration_StartStopSpawnRequestTypes(t *testing.T) {
	hs, sim := newTestServer(t)
	sim.RegisterActorType(reflect.TypeOf(&testActor{}), "test")

	conn, cancel := dialWS(t, hs)
	defer cancel()
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	readFrame(t, ctx, conn) // consume snapshot

	// start — no payload
	if err := wsjson.Write(ctx, conn, map[string]any{"id": "c-start", "type": "start"}); err != nil {
		t.Fatalf("write start: %v", err)
	}
	f := readFrame(t, ctx, conn)
	if f["type"] != "ack" || f["replyTo"] != "c-start" {
		t.Fatalf("start: expected ack, got %v", f)
	}

	// spawn — get an ack with the actor id as result
	if err := wsjson.Write(ctx, conn, map[string]any{
		"id": "c-spawn", "type": "spawn",
		"payload": map[string]any{"name": "test"},
	}); err != nil {
		t.Fatalf("write spawn: %v", err)
	}
	// drain events until we see the ack (actor.spawned event may arrive first)
	var spawnAck map[string]any
	for spawnAck == nil {
		f = readFrame(t, ctx, conn)
		if f["type"] == "ack" && f["replyTo"] == "c-spawn" {
			spawnAck = f
		}
	}
	if spawnAck["result"] == nil {
		t.Error("spawn ack should carry actor id as result")
	}

	// requestTypes — result should be a list containing "test"
	if err := wsjson.Write(ctx, conn, map[string]any{"id": "c-types", "type": "requestTypes"}); err != nil {
		t.Fatalf("write requestTypes: %v", err)
	}
	var typesAck map[string]any
	for typesAck == nil {
		f = readFrame(t, ctx, conn)
		if f["type"] == "ack" && f["replyTo"] == "c-types" {
			typesAck = f
		}
	}
	types, ok := typesAck["result"].([]any)
	if !ok || len(types) == 0 {
		t.Fatalf("requestTypes result = %v, want non-empty list", typesAck["result"])
	}

	// stop
	if err := wsjson.Write(ctx, conn, map[string]any{"id": "c-stop", "type": "stop"}); err != nil {
		t.Fatalf("write stop: %v", err)
	}
	var stopAck map[string]any
	for stopAck == nil {
		f = readFrame(t, ctx, conn)
		if f["type"] == "ack" && f["replyTo"] == "c-stop" {
			stopAck = f
		}
	}
}

func TestIntegration_SpawnUnknownTypeError(t *testing.T) {
	hs, sim := newTestServer(t)
	go sim.Start()
	t.Cleanup(sim.Stop)

	conn, cancel := dialWS(t, hs)
	defer cancel()
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	readFrame(t, ctx, conn) // consume snapshot

	if err := wsjson.Write(ctx, conn, map[string]any{
		"id": "c-bad-spawn", "type": "spawn",
		"payload": map[string]any{"name": "nonexistent"},
	}); err != nil {
		t.Fatalf("write: %v", err)
	}
	frame := readFrame(t, ctx, conn)
	if frame["type"] != "error" {
		t.Fatalf("expected error, got %v", frame["type"])
	}
	if frame["code"] != "INVALID_ARGUMENT" {
		t.Fatalf("expected INVALID_ARGUMENT, got %v", frame["code"])
	}
}

func TestIntegration_UnknownCommandError(t *testing.T) {
	hs, sim := newTestServer(t)
	go sim.Start()
	t.Cleanup(sim.Stop)

	conn, cancel := dialWS(t, hs)
	defer cancel()
	ctx, ctxCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer ctxCancel()

	readFrame(t, ctx, conn) // consume snapshot

	if err := wsjson.Write(ctx, conn, map[string]any{
		"id":   "c-99",
		"type": "no.such.command",
	}); err != nil {
		t.Fatalf("write: %v", err)
	}

	frame := readFrame(t, ctx, conn)
	if frame["type"] != "error" {
		t.Fatalf("expected error frame, got %v", frame["type"])
	}
	if frame["code"] != "UNKNOWN_COMMAND" {
		t.Fatalf("expected UNKNOWN_COMMAND, got %v", frame["code"])
	}
	if frame["replyTo"] != "c-99" {
		t.Fatalf("replyTo = %v, want c-99", frame["replyTo"])
	}
}
