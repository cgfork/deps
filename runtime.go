package deps

import (
	"context"
	"fmt"
	"log/slog"
	"reflect"
	"strings"
	"sync"

	"github.com/BurntSushi/toml"
)

// Runtime is the interface for the deps runtime which provides
// the way to resolve the references or implementations.
type Runtime interface {
	// GetImpl returns the instance of the given type.
	GetImpl(reflect.Type) (any, error)
	// GetIntf returns the implementation instance of the given interface Type
	// with the given name.
	GetIntf(reflect.Type, string) (any, error)
}

type Config struct {
	Config  string
	Fakes   map[reflect.Type]any
	Present map[string]any
	Root    *slog.Logger
}

type runtime struct {
	depsByName map[string]*Dep
	depsByIntf map[reflect.Type]map[string]*Dep
	depsByImpl map[reflect.Type]*Dep

	ctx      context.Context
	config   Config
	sections map[string]string

	mu    sync.Mutex
	impls map[string]any
}

// NewRuntime returns a new Runtime.
func NewRuntime(ctx context.Context, regs []*Dep, config Config) (Runtime, error) {
	return newRuntime(ctx, regs, config)
}
func newRuntime(ctx context.Context, deps []*Dep, config Config) (*runtime, error) {
	sections, err := ParseTOML(config.Config)
	if err != nil {
		return nil, err
	}

	depsByName := map[string]*Dep{}
	depsByIntf := map[reflect.Type]map[string]*Dep{}
	depsByImpl := map[reflect.Type]*Dep{}
	for _, dep := range deps {
		_, ok := depsByName[dep.id]
		if ok {
			return nil, fmt.Errorf("multiple deps found for %s", dep.id)
		}
		depsByName[dep.id] = dep

		intfs, ok := depsByIntf[dep.iface]
		if !ok {
			intfs = make(map[string]*Dep)
			depsByIntf[dep.iface] = intfs
		}

		_, ok = intfs[dep.name]
		if ok {
			return nil, fmt.Errorf("multiple deps implemented the same interface found for %s", dep.name)
		}

		intfs[dep.name] = dep

		_, ok = depsByImpl[dep.impl]
		if ok {
			return nil, fmt.Errorf("multiple deps found for the same implementation %v", dep.impl)
		}
		depsByImpl[dep.impl] = dep
	}

	impls := map[string]any{}
	for k, v := range config.Present {
		impls[k] = v
	}

	if config.Root == nil {
		config.Root = slog.Default()
	}

	return &runtime{
		depsByName: depsByName,
		depsByIntf: depsByIntf,
		depsByImpl: depsByImpl,
		ctx:        ctx,
		config:     config,
		sections:   sections,
		impls:      impls,
	}, nil
}

func (r *runtime) GetImpl(t reflect.Type) (any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getImpl(t)
}

func (r *runtime) GetIntf(t reflect.Type, name string) (any, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.getIntf(t, name, "root")
}

func (r *runtime) getImpl(t reflect.Type) (any, error) {
	dep, ok := r.depsByImpl[t]
	if !ok {
		return nil, fmt.Errorf("the implementation %v not found", t)
	}

	return r.get(dep)
}

func (r *runtime) hook(reg *Dep, impl any, requester string) (any, error) {
	if reg.hook != nil {
		return reg.hook(impl, requester), nil
	}
	return impl, nil
}

func (r *runtime) getIntf(t reflect.Type, name, requester string) (any, error) {
	deps, ok := r.depsByIntf[t]
	if !ok {
		return nil, fmt.Errorf("dep %v not found; maybe you forgot to register the Registration", t)
	}

	get := func(reg *Dep) (any, error) {
		v, err := r.get(reg)
		if err != nil {
			return nil, err
		}
		return r.hook(reg, v, requester)
	}

	if name != "" {
		dep, ok := deps[name]
		if !ok {
			return nil, fmt.Errorf("dep %v not found; maybe you forgot to register the Registration", t)
		}
		return get(dep)
	}

	var dep *Dep
	for _, v := range deps {
		if v.name == v.id {
			// Anonymous dependency
			dep = v
			break
		}
	}

	if dep == nil {
		return nil, fmt.Errorf("no anonymous dep found for %v", t)
	}

	return get(dep)
}

func (r *runtime) get(dep *Dep) (any, error) {
	if c, ok := r.impls[dep.id]; ok {
		return c, nil
	}

	if fake, ok := r.config.Fakes[dep.iface]; ok {
		return fake, nil
	}

	v := reflect.New(dep.impl)
	obj := v.Interface()

	// Setup
	if err := setupConfig(dep.name, v, r.sections); err != nil {
		return nil, err
	}

	setupLog(obj, r.config.Root)

	if err := setupImpl(obj, r); err != nil {
		return nil, err
	}

	if err := setupRefs(obj, func(t reflect.Type, name string) (any, error) {
		return r.getIntf(t, name, dep.name)
	}); err != nil {
		return nil, err
	}

	// Call the Init method.
	if i, ok := obj.(interface{ Init(context.Context) error }); ok {
		if err := i.Init(r.ctx); err != nil {
			return nil, fmt.Errorf("dep %q initialization failed: %w", dep.name, err)
		}
	}
	r.impls[dep.id] = obj
	return obj, nil
}

// ParseTOML parses the provided TOML input and returns a map of sections.
func ParseTOML(input string) (map[string]string, error) {
	var sections map[string]toml.Primitive
	_, err := toml.Decode(input, &sections)
	if err != nil {
		return nil, err
	}

	config := struct{ Sections map[string]string }{
		Sections: map[string]string{},
	}
	for k, v := range sections {
		var buf strings.Builder
		err := toml.NewEncoder(&buf).Encode(v)
		if err != nil {
			return nil, fmt.Errorf("encoding section %q: %w", k, err)
		}
		config.Sections[k] = buf.String()
	}

	return config.Sections, nil
}
