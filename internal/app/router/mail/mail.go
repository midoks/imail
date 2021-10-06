package mail

import (
	"fmt"

	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/db"
	"github.com/midoks/imail/internal/tools/paginater"
)

const (
	MAIL     = "mail/list"
	USER_NEW = "mail/new"
)

type MailSearchOptions struct {
	page     int
	PageSize int
	OrderBy  string
	TplName  string
}

func RenderMailSearch(c *context.Context, opts *MailSearchOptions) {
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

	fmt.Println(c.Data["Page"])
	c.Data["Mail"] = mail

	c.Success(opts.TplName)
}

func Mail(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	c.Success(MAIL)
}

func New(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.write_letter")
	c.Data["PageIsWriteMail"] = true

	c.Success(USER_NEW)
}
