package context

import (
	"fmt"
	"github.com/midoks/imail/internal/conf"
	"gopkg.in/macaron.v1"
)

type ToggleOptions struct {
	SignInRequired  bool
	SignOutRequired bool
	AdminRequired   bool
	DisableCSRF     bool
}

func Toggle(options *ToggleOptions) macaron.Handler {

	return func(c *Context) {
		if !conf.Security.InstallLock {

			fmt.Println("not install")
			c.RedirectSubpath("/install")
			return
		}
	}
}
