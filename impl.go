package deps

import (
	"fmt"
)

// / Implements is a struct that can be embedded in an implementation.
type Implements[T any] struct {
	runtime Runtime

	// The iface type
	// nolint
	xxx_ifaceType T
}

// InstanceOf[T] is the interface implemented by a struct that embeds
// deps.Implements[T].
type InstanceOf[T any] interface {
	implements(T)
}

// implements is a method that can only be implemented inside the
// deps package. It exists so that a struct that embeds
// Implements[T] implements that InstanceOf[T] interface.
// nolint
func (Implements[T]) implements(T) {}

// setRuntime sets the runtime when the it runs it.
func (i *Implements[T]) setRuntime(runtime Runtime) {
	i.runtime = runtime
}

// getRuntime is the interface for visiting the runtime
// stored in Implements[T].
type getRuntime interface {
	xxx_getRuntime() Runtime
}

// xxx_getRuntime returns the runtime.
func (i Implements[T]) xxx_getRuntime() Runtime {
	return i.runtime
}

// GetIntf returns the implementation of the interface T and with the specified name.
func GetIntf[T any](gr getRuntime, name string) (T, error) {
	var t T
	v, err := gr.xxx_getRuntime().GetIntf(Type[T](), name)
	if err != nil {
		return t, err
	}
	return v.(T), nil
}

// GetImpl returns the object instance of the implementation T.
func GetImpl[T any](gr getRuntime) (*T, error) {
	v, err := gr.xxx_getRuntime().GetImpl(Type[T]())
	if err != nil {
		return nil, err
	}
	return v.(*T), nil
}

func setupImpl(impl any, runtime Runtime) error {
	x, ok := impl.(interface{ setRuntime(Runtime) })
	if !ok {
		return fmt.Errorf("%T does not embed deps.Implements", impl)
	}
	x.setRuntime(runtime)
	return nil
}
