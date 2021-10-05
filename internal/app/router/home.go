package router

import (
	// "fmt"
	// "net/http"

	"github.com/midoks/imail/internal/app/context"
	// "github.com/midoks/imail/internal/db"
	// "github.com/midoks/imail/internal/tools/paginater"
)

const (
	HOME = "home"
)

// func RenderUserSearch(c *context.Context, opts *UserSearchOptions) {
// 	page := c.QueryInt("page")
// 	if page <= 1 {
// 		page = 1
// 	}

// 	var (
// 		users []*db.User
// 		count int64
// 		err   error
// 	)

// 	keyword := c.Query("q")
// 	if len(keyword) == 0 {
// 		users, err = opts.Ranger(page, opts.PageSize)
// 		if err != nil {
// 			c.Error(err, "ranger")
// 			return
// 		}
// 		count = opts.Counter()
// 	} else {
// 		users, count, err = db.SearchUserByName(&db.SearchUserOptions{
// 			Keyword:  keyword,
// 			Type:     opts.Type,
// 			OrderBy:  opts.OrderBy,
// 			Page:     page,
// 			PageSize: opts.PageSize,
// 		})
// 		if err != nil {
// 			c.Error(err, "search user by name")
// 			return
// 		}
// 	}
// 	c.Data["Keyword"] = keyword
// 	c.Data["Total"] = count
// 	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
// 	c.Data["Users"] = users

// 	c.Success(opts.TplName)
// }

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
