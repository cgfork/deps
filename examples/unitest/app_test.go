package unitest

import (
	"testing"

	"github.com/cgfork/deps"
	"github.com/cgfork/deps/deptest"
)

func TestApp(t *testing.T) {
	deptest.Test(t, deps.Config{}, func(t *testing.T, app App) {
		app.Display()
	})
	deptest.Test(t, deps.Config{}, func(t *testing.T, app *app) {
		app.Display()
	})
}
