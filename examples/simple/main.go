package main

import (
	"context"
	"fmt"

	"github.com/cgfork/deps"
)

type Foo interface {
	Display()
}

type fooConfig struct {
	Name string
}

func (c *fooConfig) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("name is empty")
	}
	return nil
}

type foo struct {
	deps.Implements[Foo]
	deps.WithConfig[fooConfig]
}

func (f *foo) Display() {
	fmt.Println("foo", f.Config().Name)
}

type fooA struct {
	deps.Implements[Foo]       `impl:"fooA"`
	deps.WithConfig[fooConfig] `section:"fooA"`
}

func (f *fooA) Display() {
	fmt.Println("fooA", f.Config().Name)
}

type fooB struct {
	deps.Implements[Foo]       `impl:"fooB"`
	deps.WithConfig[fooConfig] // Default to use fooB section
}

func (f *fooB) Display() {
	fmt.Println("fooB", f.Config().Name)
}

type app struct {
	deps.Implements[deps.System]

	foo  deps.Ref[Foo] // Anonymouse reference
	fooA deps.Ref[Foo] `ref:"fooA"`
	fooB deps.Ref[Foo] `ref:"fooB"`
}

func main() {
	deps.MustProvide[deps.System, app]()
	deps.MustProvide[Foo, foo]()
	deps.MustProvide[Foo, fooA]()
	deps.MustProvide[Foo, fooB]()
	var config = `
	[fooB]
	name = "xyz"

	[fooA]
	name = "abc"

	['main.Foo'] # full-packaged name for iface type
	name = "foo"
	`
	if err := deps.Run[app](context.Background(), deps.Config{
		Config: config,
	}, func(_ context.Context, app *app) error {
		app.foo.Get().Display()
		app.fooA.Get().Display()
		app.fooB.Get().Display()
		f, err := deps.GetIntf[Foo](app, "")
		if err != nil {
			return err
		}
		f.Display()

		fa, err := deps.GetImpl[fooA](app)
		if err != nil {
			return err
		}
		fa.Display()
		return nil
	}); err != nil {
		panic(err)
	}
}
