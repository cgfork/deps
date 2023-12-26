package deps

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/BurntSushi/toml"
)

const PkgPath = "github.com/cgfork/deps"

// WithConfig[T] is a type that can be embedded inside a implementation struct.
// implementation. The runtime will take per-construct configuration information
// found in the application config file and use it to initialize the contents of T.
type WithConfig[T any] struct {
	config T
}

// Config returns the configuration information for the implementation that embeds
// this [deps.WithConfig].
//
// Any fields in T that were not present in the application config file will
// have their default values.
//
// Any fields in the application config file that are not present in T will be
// flagged as an error at application startup.
func (wc *WithConfig[T]) Config() *T {
	return &wc.config
}

// xxx_getConfig returns the underlying config.
// nolint
func (wc *WithConfig[T]) xxx_getConfig() any {
	return &wc.config
}

func setupConfig(name string, value reflect.Value, sections map[string]string) error {
	v, shortKey := resolveConfigAndName(value)
	if v == nil {
		return nil
	}

	if err := unmarshalTOML(name, shortKey, sections, v); err != nil {
		return err
	}

	if x, ok := v.(interface{ Validate() error }); ok {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("validate %T: %w", x, err)
		}
	}
	return nil
}

// resolveConfigAndName calls the WithConfig.Config method on the provided value
// and get the tag of field which embeded WithConfig.
func resolveConfigAndName(v reflect.Value) (any, string) {
	if v.Kind() != reflect.Pointer || v.Elem().Kind() != reflect.Struct {
		panic(fmt.Errorf("invalid non pointer to struct value: %v", v))
	}
	s := v.Elem()
	t := s.Type()
	for i := 0; i < t.NumField(); i++ {
		// Check that f is an embedded field of type deps.WithConfig[T].
		f := t.Field(i)
		if !f.Anonymous ||
			f.Type.PkgPath() != PkgPath ||
			!strings.HasPrefix(f.Type.Name(), "WithConfig[") {
			continue
		}

		sectionName := f.Tag.Get("section")

		// Call the Config method to get a *T.
		config := s.Field(i).Addr().MethodByName("Config")
		return config.Call(nil)[0].Interface(), sectionName
	}
	return nil, ""
}

// unmarshalTOML decodes the specified TOML section into dst.
func unmarshalTOML(key, shortKey string, sections map[string]string, dst any) error {
	section, ok := sections[key]
	if shortKey != "" && shortKey != key {
		if sSection, ok2 := sections[shortKey]; ok2 {
			if ok {
				return fmt.Errorf("confliction sections %q and %q", shortKey, key)
			}
			key, section, ok = shortKey, sSection, ok2
		}
	}

	if !ok {
		return nil
	}

	md, err := toml.Decode(section, dst)
	if err != nil {
		return err
	}

	if unknown := md.Undecoded(); len(unknown) != 0 {
		return fmt.Errorf("section %q has unknown keys %v", key, unknown)
	}

	if x, ok := dst.(interface{ Validate() error }); ok {
		if err := x.Validate(); err != nil {
			return fmt.Errorf("section %q: %w", key, err)
		}
	}
	return nil
}

// HasConfig returns true if the provided implementation has an embeded
// deps.WithConfig field.
func HasConfig(impl any) bool {
	_, ok := impl.(interface{ xxx_getConfig() any })
	return ok
}

// GetConfig returns the config stored in the provided implementation, or returns
// nil if the implementation does not have a config.
func GetConfig(impl any) any {
	if c, ok := impl.(interface{ xxx_getConfig() any }); ok {
		return c.xxx_getConfig()
	}
	return nil
}
