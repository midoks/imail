package router

import (
	// "fmt"
	// "net/http"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/tools/paginater"
)

const (
	HOME = "home"
)

type HomeSearchOptions struct {
	page     int
	PageSize int
	OrderBy  string
	TplName  string
}

func RenderUserSearch(c *context.Context, opts *HomeSearchOptions) {
	page := c.QueryInt("page")
	if page <= 1 {
		page = 1
	}

	var (
		mail  []*db.Mail
		count int64
		err   error
	)

	keyword := c.Query("q")
	if len(keyword) == 0 {
		mail, err = db.MailList(page, opts.PageSize)
		count = db.MailCount()
	} else {
		mail, count, err = db.MailSearchByName(&db.MailSearchOptions{
			Keyword:  keyword,
			OrderBy:  opts.OrderBy,
			Page:     page,
			PageSize: opts.PageSize,
		})
		if err != nil {
			c.Error(err, "search user by name")
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)
	c.Data["Mail"] = mail

	c.Success(opts.TplName)
}

func Home(c *context.Context) {

	c.Data["PageIsHome"] = true
	c.Success(HOME)
}
