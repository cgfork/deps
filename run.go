package deps

import (
	"context"
	"errors"
	"fmt"
	"reflect"
)

// System is the interface implemented by a deps system which is started by the
// `RunSystem` function.
type System interface {
}

// PointerToSystem is a type constraint that asserts *T is an instance of System.
type PointerToSystem[T any] interface {
	*T

	// Ensure that the T type embeds Implements[Job].
	InstanceOf[System]
}

// Run starts a deps system.
func Run[T any, P PointerToSystem[T]](ctx context.Context, config Config, start func(context.Context, *T) error) error {
	regs := Registered()
	if err := ValidateDeps(regs); err != nil {
		return err
	}

	r, err := newRuntime(ctx, regs, config)
	if err != nil {
		return err
	}

	sys, err := r.GetImpl(Type[T]())
	if err != nil {
		return err
	}
	return start(ctx, sys.(*T))
}

// ValidateDeps validates the given registrations.
// It makes sure that every type which is refered by the runtime.Ref field of the impl type
// has been registered.
func ValidateDeps(deps []*Dep) error {
	// Gather the set of registered interfaces.
	intfs := map[reflect.Type]struct{}{}
	for _, reg := range deps {
		intfs[reg.iface] = struct{}{}
	}

	// Check that for every deps.Ref[T] field in an implementation
	// struct, T is a registered interface.
	var errs []error
	for _, dep := range deps {
		for i := 0; i < dep.impl.NumField(); i++ {
			f := dep.impl.Field(i)
			switch {
			case f.Type.Implements(Type[interface{ isRef() }]()):
				// f is a deps.Ref[T].
				v := f.Type.Field(0) // a Ref[T]'s value field
				if _, ok := intfs[v.Type]; !ok {
					// T is not a registered runtime interface.
					err := fmt.Errorf(
						"the implementation struct %v has reference field %v, but %v was not registered; maybe you forgot to register it",
						dep.impl, f.Type, v.Type,
					)
					errs = append(errs, err)
				}

			}
		}
	}
	return errors.Join(errs...)
}
