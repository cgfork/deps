package deps

import (
	"fmt"
	"reflect"
)

// Ref[T] is a field that can be placed inside an implementation
// struct. T must be a registered type. Runtime will automatically
// wire such a field with a handle to the corresponding implementation.
type Ref[T any] struct {
	value T
}

// Get returns a handle to implementation of type T.
func (r Ref[T]) Get() T { return r.value }

// isRef is an internal method that is only implemented by Ref[T] and is
// used internally to check that a value is of type Ref[T].
func (r Ref[T]) isRef() {}

// setRef sets the underlying value of a Ref.
func (r *Ref[T]) setRef(value any) {
	var ok bool
	r.value, ok = value.(T)
	if !ok {
		panic(fmt.Errorf("value type assertion failed, %T is not %T", value, r.value))
	}
}

func setupRefs(impl any, get func(reflect.Type, string) (any, error)) error {
	p := reflect.ValueOf(impl)
	if p.Kind() != reflect.Pointer {
		return fmt.Errorf("%T not a pointer", impl)
	}
	s := p.Elem()
	if s.Kind() != reflect.Struct {
		return fmt.Errorf("%T not a struct pointer", impl)
	}

	typ := s.Type()
	for i, n := 0, s.NumField(); i < n; i++ {
		f := s.Field(i)
		if !f.CanAddr() {
			continue
		}
		p := reflect.NewAt(f.Type(), f.Addr().UnsafePointer()).Interface()
		x, ok := p.(interface{ setRef(any) })
		if !ok {
			continue
		}

		// Set the dep.
		valueField := f.Field(0)
		name := typ.Field(i).Tag.Get("ref")
		dep, err := get(valueField.Type(), name)
		if err != nil {
			return fmt.Errorf("setting field %v.%s: %w", typ, typ.Field(i).Name, err)
		}
		x.setRef(dep)
	}
	return nil
}
