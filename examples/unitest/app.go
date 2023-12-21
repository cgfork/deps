package unitest

import (
	"fmt"

	"github.com/cgfork/deps"
)

func init() {
	deps.MustProvide[App, app]()
}

type App interface {
	Display()
}

type app struct {
	deps.Implements[App]
}

func (a *app) Display() {
	fmt.Println("It's OK!!!")
}
