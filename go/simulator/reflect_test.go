package simulator

import (
	"errors"
	"reflect"
	"testing"
)

type reflectActor struct {
	id      int
	Count   int `json:"count"`
	Name    string
	Ratio   float64
	Enabled bool
	Hidden  string `json:"-"`
}

func (a *reflectActor) ID() int                   { return a.id }
func (a *reflectActor) Init(id int, _ *Simulator) { a.id = id }
func (a *reflectActor) OnMessage(Message)         {}
func (a *reflectActor) Add(n int) int             { a.Count += n; return a.Count }
func (a *reflectActor) Greet(name string) string  { return "hi " + name }
func (a *reflectActor) Pair() (int, string)       { return 1, "a" }
func (a *reflectActor) Fail() error               { return errors.New("boom") }
func (a *reflectActor) Panic()                    { panic("oops") }
func (a *reflectActor) Variadic(xs ...int)        {}

func newReflectSim(t *testing.T) (*Simulator, *reflectActor) {
	t.Helper()
	sim := New()
	sim.RegisterActorType(reflect.TypeOf(&reflectActor{}), "r")
	return sim, sim.SpawnActor("r", -1).(*reflectActor)
}

func TestSetActorField(t *testing.T) {
	sim, a := newReflectSim(t)

	// JSON numbers arrive as float64; fields are addressed by json tag or Go name.
	cases := []struct {
		field string
		value any
	}{
		{"count", float64(7)},
		{"Name", "bob"},
		{"Ratio", 0.5},
		{"Enabled", true},
	}
	for _, c := range cases {
		if err := sim.SetActorField(a.ID(), c.field, c.value); err != nil {
			t.Fatalf("SetActorField(%q, %v): %v", c.field, c.value, err)
		}
	}
	if a.Count != 7 || a.Name != "bob" || a.Ratio != 0.5 || !a.Enabled {
		t.Fatalf("fields not set: %+v", a)
	}

	if err := sim.SetActorField(a.ID(), "count", 1.5); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("non-integer for int field: got %v, want ErrInvalidArgument", err)
	}
	if err := sim.SetActorField(a.ID(), "Name", float64(1)); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("number for string field: got %v, want ErrInvalidArgument", err)
	}
	for _, f := range []string{"id", "missing", "Hidden"} {
		if err := sim.SetActorField(a.ID(), f, "x"); !errors.Is(err, ErrFieldNotFound) {
			t.Errorf("field %q: got %v, want ErrFieldNotFound", f, err)
		}
	}
	if err := sim.SetActorField(99, "count", float64(1)); !errors.Is(err, ErrActorNotFound) {
		t.Errorf("unknown actor: got %v, want ErrActorNotFound", err)
	}
}

func TestInvokeActor(t *testing.T) {
	sim, a := newReflectSim(t)

	got, err := sim.InvokeActor(a.ID(), "Add", []any{float64(3)})
	if err != nil || got != 3 {
		t.Fatalf("Add(3) = %v, %v; want 3, nil", got, err)
	}
	got, err = sim.InvokeActor(a.ID(), "Greet", []any{"ann"})
	if err != nil || got != "hi ann" {
		t.Fatalf("Greet = %v, %v", got, err)
	}
	got, err = sim.InvokeActor(a.ID(), "Pair", nil)
	if err != nil || !reflect.DeepEqual(got, []any{1, "a"}) {
		t.Fatalf("Pair = %v, %v", got, err)
	}
	if _, err = sim.InvokeActor(a.ID(), "Fail", nil); err == nil || err.Error() != "boom" {
		t.Errorf("Fail: got %v, want boom", err)
	}
	if _, err = sim.InvokeActor(a.ID(), "Panic", nil); err == nil {
		t.Error("Panic: expected error from recovered panic")
	}
	if _, err = sim.InvokeActor(a.ID(), "Add", nil); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("wrong arg count: got %v, want ErrInvalidArgument", err)
	}
	if _, err = sim.InvokeActor(a.ID(), "Add", []any{"x"}); !errors.Is(err, ErrInvalidArgument) {
		t.Errorf("wrong arg type: got %v, want ErrInvalidArgument", err)
	}
	for _, m := range []string{"OnMessage", "Init", "Missing"} {
		if _, err = sim.InvokeActor(a.ID(), m, nil); !errors.Is(err, ErrMethodNotFound) {
			t.Errorf("%s: got %v, want ErrMethodNotFound", m, err)
		}
	}
}

func TestSnapshotActorFieldsAndMethods(t *testing.T) {
	sim, _ := newReflectSim(t)
	snap := sim.GetSnapshot()
	if len(snap.Actors) != 1 {
		t.Fatalf("got %d actors", len(snap.Actors))
	}
	a := snap.Actors[0]

	if _, ok := a.Fields["count"]; !ok {
		t.Error("fields missing json-tagged count")
	}
	for _, hidden := range []string{"id", "Hidden", "-"} {
		if _, ok := a.Fields[hidden]; ok {
			t.Errorf("fields should not contain %q", hidden)
		}
	}

	methods := map[string][]string{}
	for _, m := range a.Methods {
		methods[m.Name] = m.Args
	}
	if !reflect.DeepEqual(methods["Add"], []string{"int"}) {
		t.Errorf("Add args = %v, want [int]", methods["Add"])
	}
	for _, excluded := range []string{"ID", "Init", "OnMessage", "Variadic"} {
		if _, ok := methods[excluded]; ok {
			t.Errorf("methods should not contain %s", excluded)
		}
	}
}
