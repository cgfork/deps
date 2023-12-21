package deps

import "fmt"

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

// embeddedImplement is a method that can on be implemented inside
// the deps package. It used to check whether an struct embedded
// the Implements[T] and help to get tag of the implementation.
func (i Implements[T]) embeddedImplement() {}

// setRuntime sets the runtime when the it runs it.
func (i *Implements[T]) setRuntime(runtime Runtime) {
	i.runtime = runtime
}

// func (i Implements[T]) GetImpl[V any]() (*V, error) {
// 	v, err := i.runtime.GetImpl(Type[T])
// 	if err != nil {
// 		return nil ,err
// 	}
// 	return v.(*V), nil
// }

func setupImpl(impl any, runtime Runtime) error {
	x, ok := impl.(interface{ setRuntime(Runtime) })
	if !ok {
		return fmt.Errorf("%T does not embed deps.Implements", impl)
	}
	x.setRuntime(runtime)
	return nil
}
