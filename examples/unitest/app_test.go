package unitest

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/cgfork/deps"
	"github.com/cgfork/deps/deptest"
)

type mock struct {
	deps.Implements[App]
}

func (m *mock) Display() {
	fmt.Println("Mock is OK!!!")
}

func TestApp(t *testing.T) {
	deptest.Test(t, deps.Config{}, func(t *testing.T, app App) {
		app.Display()
	})
	deptest.Test(t, deps.Config{}, func(t *testing.T, app *app) {
		app.Display()
	})
	deptest.Test(t, deps.Config{
		Fakes: map[reflect.Type]any{
			deps.Type[App](): &mock{},
		},
	}, func(t *testing.T, app App) {
		app.Display()
	})
}
