package deps

import (
	"errors"
	"reflect"
)

// Dep represents the relationship between an interface and its implementation
type Dep struct {
	// id is the unique id for this dependency in the runtime.
	// Default is full package-prefixed iface name + tag name.
	id string
	// human-readable name
	// Default is tag name or full package-prefixed iface name.
	name string
	// interface type for the implementation
	iface reflect.Type
	// implementation type (struct)
	impl reflect.Type

	// indicates the Iface is singleton
	singleton bool
	// Functions that return different types of stubs.
	hook func(impl any, caller string) any
}

// Option is used to setup the Dep.
type Option func(*Dep)

func NewDep[Iface any, Impl any](opts ...Option) (Dep, error) {
	iface, err := IfaceType[Iface]()
	if err != nil {
		return Dep{}, err
	}
	impl, err := StructType[Impl]()
	if err != nil {
		return Dep{}, err
	}

	dep := Dep{iface: iface, impl: impl}
	for _, o := range opts {
		o(&dep)
	}

	_, ok := dep.impl.FieldByName("xxx_ifaceType")
	if !ok {
		return Dep{}, errors.New("implementation does not embed deps.Implements")
	}

	for i := 0; i < dep.impl.NumField(); i++ {
		// name := f.Tag.Get("impl")
		f := dep.impl.Field(i)
		switch {
		case f.Type.Implements(Type[interface{ embeddedImplement() }]()):
			dep.name = f.Tag.Get("impl")
		}
	}

	fullname := dep.iface.PkgPath() + "." + dep.iface.Name()

	// Anonymous dependency
	if dep.name == "" {
		dep.name = fullname
	}

	if dep.id == "" {
		if dep.name == fullname {
			dep.id = fullname
		} else {
			dep.id = fullname + "$" + dep.name
		}
	}

	return dep, nil
}
