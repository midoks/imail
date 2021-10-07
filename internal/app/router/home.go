package router

import (
	"github.com/midoks/imail/internal/app/context"
	"github.com/midoks/imail/internal/app/router/mail"
	"github.com/midoks/imail/internal/db"
)

const (
	HOME = "mail/list"
)

func Home(c *context.Context) {
	c.Data["PageIsHome"] = true

	mail.RenderMailSearch(c, &mail.MailSearchOptions{
		PageSize: 10,
		OrderBy:  "id ASC",
		TplName:  HOME,
		Type:     db.MailSearchOptionsTypeInbox,
	})
}
