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
	Type     int
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
	opt := &db.MailSearchOptions{
		Keyword:  keyword,
		OrderBy:  opts.OrderBy,
		Page:     page,
		PageSize: opts.PageSize,
		Type:     opts.Type,
	}

	if len(keyword) == 0 {
		mail, err = db.MailList(page, opts.PageSize, opt)
		count = db.MailCountWithOpts(opt)
	} else {
		mail, count, err = db.MailSearchByName(opt)
		if err != nil {
			c.Error(err, "search user by name")
			return
		}
	}
	c.Data["Keyword"] = keyword
	c.Data["Total"] = count
	c.Data["Page"] = paginater.New(int(count), opts.PageSize, page, 5)

	fmt.Println("page:", c.Data["Page"], count)
	c.Data["Mail"] = mail

	c.Success(opts.TplName)
}

func Flags(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.flags")
	c.Data["PageIsWriteMail"] = true

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeFlags,
	})
}

func Sent(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.sent")
	c.Data["PageIsWriteMail"] = true

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeSend,
	})
}

func Junk(c *context.Context) {
	c.Data["Title"] = c.Tr("mail.junk")
	c.Data["PageIsWriteMail"] = true

	RenderMailSearch(c, &MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  MAIL,
		Type:     db.MailSearchOptionsTypeJunk,
	})
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
