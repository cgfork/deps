package deps

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
)

// globalRegistry is the global registry used by Register and Registered.
var globalRegistry registry

// Registered returns the types registered with Register.
func Registered() []*Dep {
	return globalRegistry.alldeps()
}

// Search returns the registrations of dependencies that implement the given type.
func Search(typ reflect.Type) ([]*Dep, bool) {
	return globalRegistry.search(typ)
}

// registry is a repository for registered dependencies.
// Entries are typically added to the default registry by calls
// to Register in init functions.
type registry struct {
	m    sync.Mutex
	deps map[reflect.Type][]*Dep // the set of registered deps, by their interface types
	byId map[string]*Dep         // map from full dependency name to registration
}

func (r *registry) register(dep Dep) error {
	if err := verifyDep(dep); err != nil {
		return fmt.Errorf("Register(%q): %w", dep.name, err)
	}

	r.m.Lock()
	defer r.m.Unlock()
	if old, ok := r.byId[dep.id]; ok {
		return fmt.Errorf("dep %s already registered for type %v when registering %v",
			dep.name, old.impl, dep.impl)
	}

	if r.deps == nil {
		r.deps = map[reflect.Type][]*Dep{}
	}

	if r.byId == nil {
		r.byId = map[string]*Dep{}
	}

	if dep.singleton && len(r.deps[dep.iface]) > 0 {
		return fmt.Errorf("dep %s already registered when registering %v", dep.name, dep.impl)
	}

	ptr := &dep
	r.deps[dep.iface] = append(r.deps[dep.iface], ptr)
	r.byId[dep.id] = ptr
	return nil
}

func verifyDep(dep Dep) error {
	if dep.iface == nil {
		return errors.New("missing dep type")
	}
	if dep.iface.Kind() != reflect.Interface {
		return errors.New("dep type is not an interface")
	}
	if dep.impl == nil {
		return errors.New("missing implementation type")
	}
	if dep.impl.Kind() != reflect.Struct {
		return errors.New("implementation type is not a struct")
	}

	if dep.id == "" {
		return errors.New("missing id")
	}

	if dep.name == "" {
		return errors.New("missing name")
	}

	return nil
}

// alldeps returns all of the registered dependencies, keyed by name.
func (r *registry) alldeps() []*Dep {
	r.m.Lock()
	defer r.m.Unlock()

	deps := make([]*Dep, 0, len(r.deps))
	for _, infos := range r.deps {
		deps = append(deps, infos...)
	}
	return deps
}

func (r *registry) search(typ reflect.Type) ([]*Dep, bool) {
	r.m.Lock()
	defer r.m.Unlock()
	regs, ok := r.deps[typ]
	return regs, ok
}

func Provide[Iface any, Impl any](opts ...Option) error {
	dep, err := NewDep[Iface, Impl]()
	if err != nil {
		return err
	}
	return globalRegistry.register(dep)
}

func MustProvide[Iface any, Impl any](opts ...Option) {
	err := Provide[Iface, Impl](opts...)
	if err != nil {
		panic(err)
	}
}
