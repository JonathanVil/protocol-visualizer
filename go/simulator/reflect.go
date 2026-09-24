package simulator

import (
	"fmt"
	"math"
	"reflect"
	"strings"
)

// MethodInfo describes an exported actor method that can be invoked remotely.
type MethodInfo struct {
	Name string   `json:"name"`
	Args []string `json:"args"` // parameter type names, e.g. "int", "string"
}

var actorInterfaceMethods = func() map[string]bool {
	names := make(map[string]bool)
	for _, t := range []reflect.Type{reflect.TypeFor[Actor](), reflect.TypeFor[Reviver]()} {
		for method := range t.Methods() {
			names[method.Name] = true
		}
	}
	return names
}()

// isActorInterfaceMethod reports whether name is part of the Actor or Reviver
// interface. These are driven by the simulator and not exposed for remote invocation.
func isActorInterfaceMethod(name string) bool {
	return actorInterfaceMethods[name]
}

// actorMethods lists the exported, non-variadic methods of an actor that are
// not part of the Actor interface.
func actorMethods(a Actor) []MethodInfo {
	t := reflect.TypeOf(a)
	methods := make([]MethodInfo, 0, t.NumMethod())
	for m := range t.Methods() {
		if isActorInterfaceMethod(m.Name) || m.Type.IsVariadic() {
			continue
		}
		args := make([]string, 0, m.Type.NumIn()-1)
		for j := 1; j < m.Type.NumIn(); j++ { // skip the receiver
			args = append(args, m.Type.In(j).String())
		}
		methods = append(methods, MethodInfo{Name: m.Name, Args: args})
	}
	return methods
}

// wireName is the name a struct field is exposed under in snapshots: its json
// tag name if present, otherwise the Go field name. It returns "" for fields
// tagged `json:"-"`.
func wireName(field reflect.StructField) string {
	tag := field.Tag.Get("json")
	if tag == "-" {
		return ""
	}
	if name := strings.Split(tag, ",")[0]; name != "" {
		return name
	}
	return field.Name
}

// fieldByWireName finds an exported struct field by its wire name or Go name.
// Fields hidden from snapshots (`json:"-"`) are not found.
func fieldByWireName(v reflect.Value, name string) reflect.Value {
	t := v.Type()
	for i := range t.NumField() {
		f := t.Field(i)
		wire := wireName(f)
		if f.IsExported() && wire != "" && (wire == name || f.Name == name) {
			return v.Field(i)
		}
	}
	return reflect.Value{}
}

// convertValue converts a loosely typed value (as decoded from JSON, where all
// numbers are float64) to a reflect.Value of type t.
func convertValue(value any, t reflect.Type) (reflect.Value, error) {
	if value == nil {
		switch t.Kind() {
		case reflect.Pointer, reflect.Interface, reflect.Map, reflect.Slice:
			return reflect.Zero(t), nil
		}
		return reflect.Value{}, fmt.Errorf("%w: cannot use null as %s", ErrInvalidArgument, t)
	}

	rv := reflect.ValueOf(value)
	out := reflect.New(t).Elem()
	switch t.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		f, ok := asFloat(rv)
		if !ok || f != math.Trunc(f) || out.OverflowInt(int64(f)) {
			return reflect.Value{}, fmt.Errorf("%w: %v is not a valid %s", ErrInvalidArgument, value, t)
		}
		out.SetInt(int64(f))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		f, ok := asFloat(rv)
		if !ok || f < 0 || f != math.Trunc(f) || out.OverflowUint(uint64(f)) {
			return reflect.Value{}, fmt.Errorf("%w: %v is not a valid %s", ErrInvalidArgument, value, t)
		}
		out.SetUint(uint64(f))
	case reflect.Float32, reflect.Float64:
		f, ok := asFloat(rv)
		if !ok {
			return reflect.Value{}, fmt.Errorf("%w: %v is not a valid %s", ErrInvalidArgument, value, t)
		}
		out.SetFloat(f)
	case reflect.String:
		if rv.Kind() != reflect.String {
			return reflect.Value{}, fmt.Errorf("%w: %v is not a string", ErrInvalidArgument, value)
		}
		out.SetString(rv.String())
	case reflect.Bool:
		if rv.Kind() != reflect.Bool {
			return reflect.Value{}, fmt.Errorf("%w: %v is not a bool", ErrInvalidArgument, value)
		}
		out.SetBool(rv.Bool())
	default:
		if !rv.Type().AssignableTo(t) {
			return reflect.Value{}, fmt.Errorf("%w: cannot use %v as %s", ErrInvalidArgument, value, t)
		}
		out.Set(rv)
	}
	return out, nil
}

func asFloat(v reflect.Value) (float64, bool) {
	switch v.Kind() {
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return float64(v.Int()), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return float64(v.Uint()), true
	}
	return 0, false
}
