package deptest

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/cgfork/deps"
)

// Test runs a sub-test of t that tests the supplied code.
// It fails at runtime if body is not a function like:
//
//	func(*testing.T, IfaceType)
//	func(*testing.T, IfaceType1, IfaceType2)
//
// # example
//
//	 func TestFoo(t *testing.T) {
//			depstest.Test(t, config, func(t *testing.T, foo Foo) {
//				// Testing code
//			})
//	 }
func Test(t *testing.T, config deps.Config, body any) {
	t.Helper()
	t.Run("depstest", func(t *testing.T) {
		run(t, config, body)
	})
}

// Bench runs a sub-benchmark of b that benchmarks the supplied code.
func Bench(b *testing.B, config deps.Config, body any) {
	b.Helper()
	b.Run("depsbench", func(b *testing.B) {
		run(b, config, body)
	})
}

func run(t testing.TB, config deps.Config, testBody any) {
	t.Helper()
	body, _, err := checkRunFunc(t, testBody)
	if err != nil {
		t.Fatal(fmt.Errorf("depstest.run argument: %v", err))
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	runner, err := deps.NewRuntime(ctx, deps.Registered(), config)
	if err != nil {
		t.Fatal(err)
	}
	if err := body(ctx, runner); err != nil {
		t.Fatal(err)
	}
}

// checkRunFunc checks that the type of the function passed to depstest.Run
// is correct (its first argument matches t and its remaining arguments are
// either component interfaces or pointer to component implementations). On
// success it returns (1) a function that gets the components and passes them
// to fn and (2) the interface types of the component implementation arguments.
func checkRunFunc(t testing.TB, fn any) (func(context.Context, deps.Runtime) error, []reflect.Type, error) {
	fnType := reflect.TypeOf(fn)
	if fnType == nil || fnType.Kind() != reflect.Func {
		return nil, nil, fmt.Errorf("not a func")
	}
	if fnType.IsVariadic() {
		return nil, nil, fmt.Errorf("must not be variadic")
	}
	n := fnType.NumIn()
	if n < 2 {
		return nil, nil, fmt.Errorf("must have at least two args")
	}
	if fnType.NumOut() > 0 {
		return nil, nil, fmt.Errorf("must have no return outputs")
	}
	if fnType.In(0) != reflect.TypeOf(t) {
		return nil, nil, fmt.Errorf("function first argument type %v does not match first deptest.Test argument %T", fnType.In(0), t)
	}
	var intfs []reflect.Type
	for i := 1; i < n; i++ {
		switch fnType.In(i).Kind() {
		case reflect.Interface:
			// Do nothing.
		case reflect.Pointer:
			intf, err := extractComponentInterfaceType(fnType.In(i).Elem())
			if err != nil {
				return nil, nil, err
			}
			intfs = append(intfs, intf)
		default:
			return nil, nil, fmt.Errorf("function argument %d type %v must be a component interface or pointer to component implementation", i, fnType.In(i))
		}
	}

	return func(_ context.Context, runner deps.Runtime) error {
		args := make([]reflect.Value, n)
		args[0] = reflect.ValueOf(t)
		for i := 1; i < n; i++ {
			argType := fnType.In(i)
			switch argType.Kind() {
			case reflect.Interface:
				comp, err := runner.GetIntf(argType, "")
				if err != nil {
					return err
				}
				args[i] = reflect.ValueOf(comp)
			case reflect.Pointer:
				comp, err := runner.GetImpl(argType.Elem())
				if err != nil {
					return err
				}
				args[i] = reflect.ValueOf(comp)
			default:
				return fmt.Errorf("argument %v has unexpected type %v", i, argType)
			}
		}
		reflect.ValueOf(fn).Call(args)
		return nil
	}, intfs, nil
}

// extractComponentInterfaceType extracts the component interface type from the
// provided implementation. For example, calling
// extractComponentInterfaceType on a struct that embeds deps.Implements[Foo]
// returns Foo.
//
// extractComponentInterfaceType returns an error if the provided type is not a
// implementation.
func extractComponentInterfaceType(t reflect.Type) (reflect.Type, error) {
	if t.Kind() != reflect.Struct {
		return nil, fmt.Errorf("type %v is not a struct", t)
	}
	// See the definition of deps.Implements.
	f, ok := t.FieldByName("xxx_ifaceType")
	if !ok {
		return nil, fmt.Errorf("type %v does not embed deps.Implements", t)
	}
	return f.Type, nil
}
