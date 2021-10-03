package router

import (
	// "fmt"
	// "net/http"

	"github.com/midoks/imail/internal/app/context"
	// "github.com/midoks/imail/internal/conf"
)

const (
	HOME = "home"
)

func Home(c *context.Context) {
	// if c.IsLogged {
	// 	if !c.User.IsActive && conf.Auth.RequireEmailConfirmation {
	// 		c.Data["Title"] = c.Tr("auth.active_your_account")
	// 		c.Success(user.ACTIVATE)
	// 	} else {
	// 		user.Dashboard(c)
	// 	}
	// 	return
	// }

	// Check auto-login.
	// uname := c.GetCookie(conf.Security.CookieUsername)
	// if len(uname) != 0 {
	// 	c.Redirect(conf.Web.Subpath + "/user/login")
	// 	return
	// }

	c.Data["PageIsHome"] = true
	c.Success(HOME)
}
